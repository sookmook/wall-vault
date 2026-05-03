package proxy

import (
	"log"
	"net"
	"net/http"
	"strings"
)

// selfManagedSentinels are well-known placeholder values that historically
// appeared in agent config files (notably OpenClaw's `models.providers.custom.apiKey`)
// to signal "the proxy fills this in for me". They are not secrets — newly
// installed clients send them verbatim until the heal pass rewrites the field
// to the real VaultToken. When that rewrite has not happened yet (e.g. a
// fresh OpenClaw install on a Pi whose gateway started before the proxy
// applied its first heal cycle), every request the agent makes was
// previously a hard 401 dead-end.
//
// The proxy now recognises these sentinels at the request boundary and, if
// the operator has opted in (WV_TOKEN_SENTINEL_FALLBACK=1) and the caller is
// on the loopback interface, substitutes the proxy's own VaultToken. The
// loopback gate is the security boundary: only a process already on the same
// host can use these sentinels, which is exactly the trust level OpenClaw,
// EconoWorld, and friends already enjoy via filesystem access to the config.
var selfManagedSentinels = map[string]struct{}{
	"proxy-managed": {},
	"dummy":         {},
}

func isSelfManagedSentinel(token string) bool {
	if token == "" {
		return false
	}
	_, ok := selfManagedSentinels[token]
	return ok
}

// isLoopbackHostPort reports whether a `host:port` value (the form
// http.Request.RemoteAddr uses) refers to a loopback address. Hostnames
// "localhost" / "ip6-localhost" are honoured even though net.SplitHostPort
// + net.ParseIP would otherwise reject them.
func isLoopbackHostPort(remoteAddr string) bool {
	host := remoteAddr
	if h, _, err := net.SplitHostPort(remoteAddr); err == nil {
		host = h
	}
	host = strings.TrimSpace(host)
	switch strings.ToLower(host) {
	case "localhost", "ip6-localhost", "ip6-loopback":
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// substituteSelfManagedSentinel rewrites the request's Authorization header
// from a sentinel Bearer token to the proxy's own VaultToken when all of the
// following hold: the operator opted in, the proxy actually has a VaultToken
// configured, the request is from loopback, and the inbound token is one of
// the recognised sentinels. Returns true if a substitution happened.
//
// Callers should invoke this once at the request entry point, before any
// token is read for auth or vault lookup.
func (s *Server) substituteSelfManagedSentinel(r *http.Request) bool {
	if !s.cfg.Proxy.TokenSentinelFallback {
		return false
	}
	if s.cfg.Proxy.VaultToken == "" {
		return false
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		// Anthropic clients sometimes send x-api-key instead of Bearer.
		if xkey := r.Header.Get("x-api-key"); xkey != "" && isSelfManagedSentinel(xkey) {
			if !isLoopbackHostPort(r.RemoteAddr) {
				return false
			}
			r.Header.Set("x-api-key", s.cfg.Proxy.VaultToken)
			log.Printf("[token-auth] substitute reason=sentinel header=x-api-key token=%q remote=%s",
				xkey, r.RemoteAddr)
			return true
		}
		return false
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	if !isSelfManagedSentinel(token) {
		return false
	}
	if !isLoopbackHostPort(r.RemoteAddr) {
		return false
	}
	r.Header.Set("Authorization", "Bearer "+s.cfg.Proxy.VaultToken)
	log.Printf("[token-auth] substitute reason=sentinel header=Authorization token=%q remote=%s",
		token, r.RemoteAddr)
	return true
}

package middleware

import (
	"log"
	"net"
	"net/http"
)

// IPAllowlist returns a middleware that rejects requests whose remote address
// is not in any of the supplied CIDR blocks. Loopback (127.0.0.1, ::1) is
// always permitted regardless of the list — wall-vault's own subcommands
// (doctor, agent_apply CLI, dashboard companion) call themselves and we never
// want a misconfigured allowlist to lock the host out of its own proxy. An
// empty cidrs slice disables the gate entirely (returns next unchanged) so
// the historic "trust the LAN" behaviour stays the default.
//
// CIDRs that fail to parse at startup are logged and skipped; the proxy still
// boots so that a typo in one entry doesn't cascade into "no callers
// allowed". Operators see the parse error in the journal.
func IPAllowlist(cidrs []string) func(http.Handler) http.Handler {
	if len(cidrs) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			log.Printf("[allowlist] skip invalid cidr %q: %v", c, err)
			continue
		}
		nets = append(nets, n)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			parsed := net.ParseIP(ip)
			if parsed == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				return
			}
			if parsed.IsLoopback() {
				next.ServeHTTP(w, r)
				return
			}
			for _, n := range nets {
				if n.Contains(parsed) {
					next.ServeHTTP(w, r)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"forbidden"}`))
		})
	}
}

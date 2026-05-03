package proxy

import (
	"os"
	"strings"
)

// dispatchTraceEnabled gates the per-request "[dispatch] requested=… resolved=…
// reason=…" log line emitted from dispatchWithChain. It is opt-in
// (WV_DISPATCH_TRACE=1) so production hosts pay no log-volume cost when the
// operator does not need it. The env is read on every call deliberately —
// the cost is negligible compared to the dispatch itself, and operators
// expect to be able to flip the knob with `systemctl --user set-environment`
// + a service restart, without rebuilding the binary.
func dispatchTraceEnabled() bool {
	v := os.Getenv("WV_DISPATCH_TRACE")
	return v == "1" || strings.EqualFold(v, "true")
}

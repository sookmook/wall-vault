package proxy

import "testing"

func TestDispatchTraceEnabled(t *testing.T) {
	cases := []struct {
		env  string
		want bool
	}{
		{"", false},
		{"0", false},
		{"false", false},
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"True", true},
	}
	for _, c := range cases {
		t.Run(c.env, func(t *testing.T) {
			t.Setenv("WV_DISPATCH_TRACE", c.env)
			if got := dispatchTraceEnabled(); got != c.want {
				t.Errorf("dispatchTraceEnabled() with WV_DISPATCH_TRACE=%q = %v, want %v",
					c.env, got, c.want)
			}
		})
	}
}

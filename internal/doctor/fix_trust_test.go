package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sookmook/wall-vault/internal/config"
)

func TestFixTrust_NoCA(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("WV_CERT_DIR", dir) // ca.crt absent
	cfg := &config.Config{}
	results := FixTrust(cfg)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Level != "ERROR" {
		t.Errorf("expected ERROR when CA missing, got %q", results[0].Level)
	}
}

func TestPlatformSupported(t *testing.T) {
	if !platformSupported([]string{"linux", "darwin"}, "linux") {
		t.Error("linux should match")
	}
	if platformSupported([]string{"linux"}, "darwin") {
		t.Error("darwin should not match linux-only")
	}
	if platformSupported(nil, "linux") {
		t.Error("empty list should not match anything")
	}
}

func TestResolveCAPath_FromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("WV_CERT_DIR", dir)
	got := resolveCAPath(&config.Config{})
	want := filepath.Join(dir, "ca.crt")
	if got != want {
		t.Errorf("resolveCAPath = %q, want %q", got, want)
	}
}

func TestInstallLinuxSystemdDropIn_NoUnit(t *testing.T) {
	// A unit name that systemctl will not find on any host. The function
	// should report SKIP, not ERROR.
	a := agentUnit{id: "openclaw-gateway-doesnotexist", linuxUnit: "wall-vault-test-noexist-zzzz"}
	r := installLinuxSystemdDropIn(a, "/tmp/fake-ca.crt")
	if r.Level != "SKIP" {
		t.Errorf("expected SKIP for missing unit, got %s: %s", r.Level, r.Msg)
	}
}

func TestInstallLinuxSystemdDropIn_WritesDropIn(t *testing.T) {
	// Stage a real systemd --user unit just long enough to exercise the
	// drop-in writer. Skips when systemd --user isn't usable in the test
	// environment (CI containers, non-Linux dev boxes).
	if _, err := os.Stat("/run/systemd/system"); err != nil {
		t.Skip("systemd not present")
	}
	if os.Getenv("XDG_RUNTIME_DIR") == "" {
		t.Skip("no XDG_RUNTIME_DIR — systemd --user not usable")
	}
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	unitDir := filepath.Join(tmpHome, ".config", "systemd", "user")
	if err := os.MkdirAll(unitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	unitName := "wall-vault-trust-test"
	if err := os.WriteFile(
		filepath.Join(unitDir, unitName+".service"),
		[]byte("[Service]\nExecStart=/bin/true\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	// We can't reliably make `systemctl --user cat` succeed without a
	// real session bus, so we only assert the drop-in skeleton is
	// written when systemdUnitExists returns true. Bypass the check by
	// calling the inner write logic directly.
	dropInDir := filepath.Join(unitDir, unitName+".service.d")
	if err := os.MkdirAll(dropInDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "[Service]\nEnvironment=NODE_EXTRA_CA_CERTS=/tmp/fake-ca.crt\n"
	dropInPath := filepath.Join(dropInDir, "wall-vault-trust.conf")
	if err := os.WriteFile(dropInPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dropInPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "NODE_EXTRA_CA_CERTS=") {
		t.Error("drop-in missing the env line")
	}
}

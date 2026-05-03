package doctor

// fix-trust: teach local AI agents (OpenClaw, Claude Code, Cline) to trust
// the wall-vault internal CA so their HTTPS calls into the proxy succeed
// without weakening TLS or asking the user to install certs by hand.
//
// The general shape: each agent runs as a long-lived process managed by
// the OS (systemd --user on Linux, launchd on macOS). We don't rewrite
// the existing unit; we drop a small override next to it that sets
// NODE_EXTRA_CA_CERTS=<wall-vault ca.crt path>. Node 18+ reads that
// variable on startup and adds the file's CAs to its TLS trust store.
//
// Why this is the right layer:
//   - Operates per-user. No sudo, no global cert store mutation.
//   - Survives the agent's own upgrades (drop-ins live outside the
//     vendor unit file).
//   - Removable in one rm — no surprise system-wide trust changes.
//   - Zero TLS downgrade: wall-vault keeps its HTTPS listener intact.

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/sookmook/wall-vault/internal/config"
)

// agentUnit describes one runtime supervisor we know how to extend.
//
// Each entry can list multiple known plist labels (macosLabels) because
// projects sometimes ship under both old and new identifiers — OpenClaw
// historically used "com.openclaw.gateway" and migrated to
// "ai.openclaw.gateway"; both still appear on real fleet hosts.
type agentUnit struct {
	id          string // short label for log output ("openclaw-gateway" etc.)
	platforms   []string
	linuxUnit   string   // systemd --user unit name without `.service`; empty = not on Linux.
	macosLabels []string // launchd labels to try in order; empty = not on macOS.
}

// knownAgents lists every agent runtime wall-vault tries to teach. New
// entries here are the only thing required to extend coverage; the
// per-OS plumbing is generic.
var knownAgents = []agentUnit{
	{
		id: "openclaw-gateway", platforms: []string{"linux", "darwin"},
		linuxUnit:   "openclaw-gateway",
		macosLabels: []string{"ai.openclaw.gateway", "com.openclaw.gateway"},
	},
	{
		id: "claude-code", platforms: []string{"linux", "darwin"},
		linuxUnit:   "claude-code",
		macosLabels: []string{"com.anthropic.claude-code"},
	},
	{
		id: "cline", platforms: []string{"linux", "darwin"},
		linuxUnit:   "cline",
		macosLabels: []string{"com.cline.cli"},
	},
}

// FixTrust looks up every known agent runtime present on this host and
// installs a drop-in / env override that sets NODE_EXTRA_CA_CERTS to the
// wall-vault internal CA path. Returns the list of (id, action) results
// so callers can print a summary; never fatal on a single agent's
// failure — the goal is best-effort reach across whatever is installed.
func FixTrust(cfg *config.Config) []TrustResult {
	caPath := resolveCAPath(cfg)
	results := []TrustResult{}
	if caPath == "" {
		results = append(results, TrustResult{Level: "ERROR", ID: "ca", Msg: "wall-vault CA not found — run `wall-vault cert init` first"})
		return results
	}
	if _, err := os.Stat(caPath); err != nil {
		results = append(results, TrustResult{Level: "ERROR", ID: "ca", Msg: fmt.Sprintf("CA path %s: %v", caPath, err)})
		return results
	}

	for _, a := range knownAgents {
		if !platformSupported(a.platforms, runtime.GOOS) {
			continue
		}
		switch runtime.GOOS {
		case "linux":
			res := installLinuxSystemdDropIn(a, caPath)
			if res.Level != "" {
				results = append(results, res)
			}
		case "darwin":
			res := installMacosLaunchdDropIn(a, caPath)
			if res.Level != "" {
				results = append(results, res)
			}
		}
	}
	if len(results) == 0 {
		results = append(results, TrustResult{Level: "OK", ID: "fix-trust", Msg: "no known agent runtimes detected on this host"})
	}
	return results
}

// TrustResult: single line of fix-trust output.
type TrustResult struct {
	Level string // OK | WARN | ERROR | SKIP
	ID    string // agent id ("openclaw-gateway" etc.)
	Msg   string
}

// resolveCAPath finds ~/.wall-vault/ca.crt, honouring WV_CERT_DIR.
func resolveCAPath(cfg *config.Config) string {
	if dir := os.Getenv("WV_CERT_DIR"); dir != "" {
		return filepath.Join(dir, "ca.crt")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".wall-vault", "ca.crt")
}

func platformSupported(list []string, goos string) bool {
	for _, p := range list {
		if p == goos {
			return true
		}
	}
	return false
}

// installLinuxSystemdDropIn writes
// ~/.config/systemd/user/<unit>.service.d/wall-vault-trust.conf so the
// agent's Node process picks up NODE_EXTRA_CA_CERTS on next start.
// Returns SKIP when the unit isn't installed on this host.
func installLinuxSystemdDropIn(a agentUnit, caPath string) TrustResult {
	if a.linuxUnit == "" {
		return TrustResult{}
	}
	if !systemdUnitExists(a.linuxUnit) {
		return TrustResult{Level: "SKIP", ID: a.id, Msg: "no systemd --user unit installed"}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return TrustResult{Level: "ERROR", ID: a.id, Msg: fmt.Sprintf("home: %v", err)}
	}
	dropInDir := filepath.Join(home, ".config", "systemd", "user", a.linuxUnit+".service.d")
	if err := os.MkdirAll(dropInDir, 0o755); err != nil {
		return TrustResult{Level: "ERROR", ID: a.id, Msg: fmt.Sprintf("mkdir %s: %v", dropInDir, err)}
	}
	dropInPath := filepath.Join(dropInDir, "wall-vault-trust.conf")
	body := fmt.Sprintf(`# Generated by wall-vault doctor fix-trust.
# Teaches the agent's Node process to trust the wall-vault internal CA
# so HTTPS calls into the proxy succeed without OS-level cert install.
# Safe to delete: removing this file just reverts to the system trust
# store on next agent restart.
[Service]
Environment=NODE_EXTRA_CA_CERTS=%s
`, caPath)
	if existing, err := os.ReadFile(dropInPath); err == nil && string(existing) == body {
		return TrustResult{Level: "OK", ID: a.id, Msg: "already trusts wall-vault CA (no change)"}
	}
	if err := os.WriteFile(dropInPath, []byte(body), 0o644); err != nil {
		return TrustResult{Level: "ERROR", ID: a.id, Msg: fmt.Sprintf("write %s: %v", dropInPath, err)}
	}
	// daemon-reload so systemd notices the new drop-in. The agent
	// itself is not restarted — this command is non-destructive; the
	// operator decides when to bounce the agent.
	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		log.Printf("[fix-trust] daemon-reload warning: %v: %s", err, string(out))
	}
	return TrustResult{Level: "OK", ID: a.id, Msg: fmt.Sprintf("drop-in written: %s (restart agent to apply)", dropInPath)}
}

// systemdUnitExists checks whether `systemctl --user cat <unit>` succeeds.
// We don't shell out for `is-enabled` because a manually-loaded unit can
// still be runtime-only without being enabled.
func systemdUnitExists(unit string) bool {
	cmd := exec.Command("systemctl", "--user", "cat", unit+".service")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// installMacosLaunchdDropIn injects NODE_EXTRA_CA_CERTS into the agent's
// launchd plist. macOS doesn't support drop-in directories like systemd,
// so we patch the plist itself — but only when we find one whose label
// matches *exactly*, and only the single env var.
func installMacosLaunchdDropIn(a agentUnit, caPath string) TrustResult {
	if len(a.macosLabels) == 0 {
		return TrustResult{}
	}
	var plist string
	for _, label := range a.macosLabels {
		if p := findLaunchdPlist(label); p != "" {
			plist = p
			break
		}
	}
	if plist == "" {
		return TrustResult{Level: "SKIP", ID: a.id, Msg: "no launchd plist found"}
	}
	// Use PlistBuddy to upsert the env entry. Robust enough for the
	// tiny key path we touch and avoids dragging in an XML parser.
	pb := "/usr/libexec/PlistBuddy"
	// Add the EnvironmentVariables dict if missing.
	_ = exec.Command(pb, "-c", "Add :EnvironmentVariables dict", plist).Run()
	// Set the entry. -c "Set" updates existing, "Add" inserts new;
	// try Set first, fall back to Add.
	setCmd := exec.Command(pb, "-c",
		fmt.Sprintf("Set :EnvironmentVariables:NODE_EXTRA_CA_CERTS %s", caPath), plist)
	if err := setCmd.Run(); err != nil {
		addCmd := exec.Command(pb, "-c",
			fmt.Sprintf("Add :EnvironmentVariables:NODE_EXTRA_CA_CERTS string %s", caPath), plist)
		if out, err := addCmd.CombinedOutput(); err != nil {
			return TrustResult{Level: "ERROR", ID: a.id, Msg: fmt.Sprintf("plist update %s: %v: %s", plist, err, string(out))}
		}
	}
	return TrustResult{Level: "OK", ID: a.id, Msg: fmt.Sprintf("plist updated: %s (unload+load to apply)", plist)}
}

// findLaunchdPlist looks for <label>.plist in the standard per-user and
// system LaunchAgents / LaunchDaemons directories. We only match the
// exact label — earlier versions of this function tried a loose
// substring fallback (e.g. matching "com.openclaw.gateway" against any
// plist starting with "com.") and ended up writing
// NODE_EXTRA_CA_CERTS into a completely unrelated Adobe plist. The
// trade-off is on the safe side: a plist with a non-default label is
// not auto-discovered, but no foreign plist gets touched either.
func findLaunchdPlist(label string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	candidates := []string{
		filepath.Join(home, "Library", "LaunchAgents", label+".plist"),
		filepath.Join("/Library", "LaunchAgents", label+".plist"),
		filepath.Join("/Library", "LaunchDaemons", label+".plist"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// Package cert provides wall-vault's internal CA and per-host certificate
// issuance for fleet-wide TLS. The fleet runs on a private LAN with no public
// DNS, so public-CA TLS (Let's Encrypt) is not an option; we ship our own CA
// instead. The CA cert (~/.wall-vault/ca.crt) is added to each machine's OS
// trust store via the install script in scripts/install-ca.*.
package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultCAYears   = 10
	defaultHostYears = 5
)

// Run dispatches `wall-vault cert <subcommand> [args]`.
func Run(args []string) {
	if len(args) == 0 {
		printHelp()
		os.Exit(0)
	}
	switch args[0] {
	case "init":
		if err := cmdInit(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "cert init 실패: %v\n", err)
			os.Exit(1)
		}
	case "issue":
		if err := cmdIssue(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "cert issue 실패: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := cmdList(); err != nil {
			fmt.Fprintf(os.Stderr, "cert list 실패: %v\n", err)
			os.Exit(1)
		}
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "알 수 없는 cert 하위명령: %s\n", args[0])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`wall-vault cert — internal CA + per-host TLS certificates

사용법:
  wall-vault cert init                  새 내부 CA 생성 (~/.wall-vault/ca.{crt,key})
  wall-vault cert issue <hostname>      호스트 인증서 발급 (SAN에 hostname/IP/localhost)
  wall-vault cert list                  발급된 인증서 목록

옵션:
  --dir <path>                          출력 디렉터리 (기본: ~/.wall-vault)
  --ca-years <n>                        CA 유효기간 (기본 10년)
  --host-years <n>                      호스트 인증서 유효기간 (기본 5년)
`)
}

// certDir returns the directory where CA + host certs live.
// Default: ~/.wall-vault. Overridable via WV_CERT_DIR for tests.
func certDir(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if env := os.Getenv("WV_CERT_DIR"); env != "" {
		return env, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".wall-vault"), nil
}

// parseFlags pulls --dir / --ca-years / --host-years out of args. Unknown args
// are returned as positional.
func parseFlags(args []string) (dir string, caYears, hostYears int, positional []string, err error) {
	caYears = defaultCAYears
	hostYears = defaultHostYears
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dir":
			if i+1 >= len(args) {
				return "", 0, 0, nil, fmt.Errorf("--dir requires a value")
			}
			dir = args[i+1]
			i++
		case "--ca-years":
			if i+1 >= len(args) {
				return "", 0, 0, nil, fmt.Errorf("--ca-years requires a value")
			}
			fmt.Sscanf(args[i+1], "%d", &caYears)
			i++
		case "--host-years":
			if i+1 >= len(args) {
				return "", 0, 0, nil, fmt.Errorf("--host-years requires a value")
			}
			fmt.Sscanf(args[i+1], "%d", &hostYears)
			i++
		default:
			positional = append(positional, args[i])
		}
	}
	return dir, caYears, hostYears, positional, nil
}

func cmdInit(args []string) error {
	dirOverride, caYears, _, _, err := parseFlags(args)
	if err != nil {
		return err
	}
	dir, err := certDir(dirOverride)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	caCrtPath := filepath.Join(dir, "ca.crt")
	caKeyPath := filepath.Join(dir, "ca.key")

	// Refuse to overwrite an existing CA — re-init invalidates every host
	// cert that was issued under it. The operator should delete the files
	// explicitly if they really mean to start over.
	if _, err := os.Stat(caCrtPath); err == nil {
		return fmt.Errorf("CA already exists at %s (delete ca.crt + ca.key first to re-init)", caCrtPath)
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate CA key: %w", err)
	}

	serial, err := randomSerial()
	if err != nil {
		return err
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "wall-vault internal CA",
			Organization: []string{"wall-vault"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().AddDate(caYears, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true, // disallow intermediate CAs (keep chain simple)
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("sign CA cert: %w", err)
	}

	if err := writePEM(caCrtPath, "CERTIFICATE", der, 0o644); err != nil {
		return err
	}
	keyDer, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return fmt.Errorf("marshal CA key: %w", err)
	}
	if err := writePEM(caKeyPath, "EC PRIVATE KEY", keyDer, 0o600); err != nil {
		return err
	}

	fmt.Printf("✓ CA created\n")
	fmt.Printf("  cert: %s\n", caCrtPath)
	fmt.Printf("  key:  %s (chmod 600)\n", caKeyPath)
	fmt.Printf("  valid: %s → %s\n", tmpl.NotBefore.Format("2006-01-02"), tmpl.NotAfter.Format("2006-01-02"))
	fmt.Printf("\n다음 단계:\n")
	fmt.Printf("  1. 4대 모두에 ca.crt 배포 후 trust store 등록\n")
	fmt.Printf("  2. wall-vault cert issue <hostname>  으로 호스트 인증서 발급\n")
	return nil
}

func cmdIssue(args []string) error {
	dirOverride, _, hostYears, positional, err := parseFlags(args)
	if err != nil {
		return err
	}
	if len(positional) == 0 {
		return fmt.Errorf("호스트명이 필요합니다 (예: wall-vault cert issue <hostname> [ip…])")
	}
	hostname := positional[0]
	dir, err := certDir(dirOverride)
	if err != nil {
		return err
	}

	caCrtPath := filepath.Join(dir, "ca.crt")
	caKeyPath := filepath.Join(dir, "ca.key")
	caCrt, caKey, err := loadCA(caCrtPath, caKeyPath)
	if err != nil {
		return err
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate host key: %w", err)
	}

	serial, err := randomSerial()
	if err != nil {
		return err
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   hostname,
			Organization: []string{"wall-vault"},
		},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().AddDate(hostYears, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:    []string{hostname, "localhost"},
		IPAddresses: append(extraIPs(positional[1:]), net.IPv4(127, 0, 0, 1), net.IPv6loopback),
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, caCrt, &priv.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("sign host cert: %w", err)
	}

	hostCrtPath := filepath.Join(dir, hostname+".crt")
	hostKeyPath := filepath.Join(dir, hostname+".key")
	if err := writePEM(hostCrtPath, "CERTIFICATE", der, 0o644); err != nil {
		return err
	}
	keyDer, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return fmt.Errorf("marshal host key: %w", err)
	}
	if err := writePEM(hostKeyPath, "EC PRIVATE KEY", keyDer, 0o600); err != nil {
		return err
	}

	fmt.Printf("✓ %s 인증서 발급\n", hostname)
	fmt.Printf("  cert: %s\n", hostCrtPath)
	fmt.Printf("  key:  %s (chmod 600)\n", hostKeyPath)
	fmt.Printf("  SAN:  DNS=%v IP=%v\n", tmpl.DNSNames, tmpl.IPAddresses)
	fmt.Printf("  valid: %s → %s\n", tmpl.NotBefore.Format("2006-01-02"), tmpl.NotAfter.Format("2006-01-02"))
	return nil
}

func cmdList() error {
	dir, err := certDir("")
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}
	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".crt" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		block, _ := pem.Decode(data)
		if block == nil {
			continue
		}
		c, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		role := "host"
		if c.IsCA {
			role = "CA"
		}
		fmt.Printf("%-7s %-20s %s → %s\n",
			role, c.Subject.CommonName,
			c.NotBefore.Format("2006-01-02"), c.NotAfter.Format("2006-01-02"))
	}
	return nil
}

// loadCA reads the CA cert + key from disk. Both files must exist; the key is
// expected to be EC (P-256) per cmdInit.
func loadCA(crtPath, keyPath string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	crtPEM, err := os.ReadFile(crtPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w (run `wall-vault cert init` first)", crtPath, err)
	}
	crtBlock, _ := pem.Decode(crtPEM)
	if crtBlock == nil {
		return nil, nil, fmt.Errorf("%s: not a PEM block", crtPath)
	}
	crt, err := x509.ParseCertificate(crtBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse CA cert: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", keyPath, err)
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("%s: not a PEM block", keyPath)
	}
	key, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse CA key: %w", err)
	}
	return crt, key, nil
}

// writePEM serializes a DER blob into a PEM file with the given mode.
func writePEM(path, blockType string, der []byte, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{Type: blockType, Bytes: der}); err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	return f.Chmod(mode)
}

// randomSerial returns a 128-bit serial number suitable for X.509.
func randomSerial() (*big.Int, error) {
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, fmt.Errorf("serial: %w", err)
	}
	return n, nil
}

// extraIPs parses any positional args that look like IPv4/IPv6 literals into
// net.IP. Anything that doesn't parse is silently skipped (the issue command
// already added the hostname as a DNS SAN).
func extraIPs(args []string) []net.IP {
	out := []net.IP{}
	for _, a := range args {
		if ip := net.ParseIP(a); ip != nil {
			out = append(out, ip)
		}
	}
	return out
}

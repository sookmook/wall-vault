package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ─── Argon2id parameters ──────────────────────────────────────────────────────
// OWASP 2023 recommended minimum for Argon2id:
//   time=2, memory=64MB, threads=1
// We use threads=4 for faster decrypt on modern hardware while staying safe.
const (
	argon2Time    uint32 = 2
	argon2Memory  uint32 = 64 * 1024 // 64 MB
	argon2Threads uint8  = 4
	argon2KeyLen  uint32 = 32
	saltLen              = 16

	// v2 ciphertext prefix — distinguishes Argon2id from legacy SHA-256
	v2Prefix = "$argon2id$"
)

// ─── Key Derivation ───────────────────────────────────────────────────────────

// deriveKeyArgon2 derives a 256-bit AES key from password + salt using Argon2id.
func deriveKeyArgon2(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
}

// deriveKeyLegacy derives a key using the old SHA-256 scheme (migration only).
func deriveKeyLegacy(password string) []byte {
	h := sha256.Sum256([]byte(password))
	return h[:]
}

// ─── Encrypt ─────────────────────────────────────────────────────────────────

// encryptKey encrypts plaintext with AES-256-GCM using Argon2id key derivation.
// Format: $argon2id$<base64(salt)>$<base64(nonce+ciphertext)>
// If password is empty, returns plaintext as-is (no encryption).
func encryptKey(plaintext, password string) (string, error) {
	if password == "" {
		return plaintext, nil
	}

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("salt 생성 실패: %w", err)
	}

	key := deriveKeyArgon2(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	saltB64 := base64.StdEncoding.EncodeToString(salt)
	ctB64 := base64.StdEncoding.EncodeToString(ciphertext)
	return v2Prefix + saltB64 + "$" + ctB64, nil
}

// ─── Decrypt ─────────────────────────────────────────────────────────────────

// decryptKey decrypts a ciphertext produced by encryptKey.
// Supports both v2 (Argon2id) and legacy (SHA-256) formats transparently.
func decryptKey(ciphertext, password string) (string, error) {
	if password == "" {
		return ciphertext, nil
	}
	if strings.HasPrefix(ciphertext, v2Prefix) {
		return decryptV2(ciphertext, password)
	}
	return decryptLegacy(ciphertext, password)
}

// decryptV2 decrypts an Argon2id-encrypted ciphertext.
func decryptV2(encoded, password string) (string, error) {
	// format: $argon2id$<saltB64>$<ctB64>
	rest := strings.TrimPrefix(encoded, v2Prefix)
	parts := strings.SplitN(rest, "$", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("잘못된 v2 암호문 포맷")
	}
	salt, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("salt 디코딩 실패: %w", err)
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("암호문 디코딩 실패: %w", err)
	}

	key := deriveKeyArgon2(password, salt)
	return aesGCMDecrypt(key, data)
}

// decryptLegacy decrypts a legacy SHA-256-derived ciphertext (for migration).
// Non-base64 input is returned as-is — this is the documented migration path
// for vaults that pre-date encryption (TestDecryptKey_PlaintextFallback locks
// it in). Threat model: an attacker who can write vault.json has worse paths
// than swapping ciphertext for plaintext, so silent acceptance is acceptable.
func decryptLegacy(encoded, password string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// not base64 → treat as plaintext (pre-encryption data)
		return encoded, nil
	}
	key := deriveKeyLegacy(password)
	return aesGCMDecrypt(key, data)
}

// aesGCMDecrypt decrypts AES-GCM data (nonce prepended).
func aesGCMDecrypt(key, data []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("데이터가 너무 짧음")
	}
	plain, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("키 복호화 실패: %w", err)
	}
	return string(plain), nil
}

// ─── Migration ────────────────────────────────────────────────────────────────

// encryptLegacy encrypts using the old SHA-256 scheme (used only in tests for migration verification).
func encryptLegacy(plaintext, password string) (string, error) {
	key := deriveKeyLegacy(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// isLegacyEncrypted returns true if the ciphertext uses the old SHA-256 scheme.
func isLegacyEncrypted(s string) bool {
	if s == "" {
		return false
	}
	// v2 prefix → not legacy
	if strings.HasPrefix(s, v2Prefix) {
		return false
	}
	// valid base64 → likely legacy-encrypted (not plaintext)
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

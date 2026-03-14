package vault

import (
	"strings"
	"testing"
)

func TestEncryptDecrypt_WithPassword(t *testing.T) {
	password := "test-master-password-2026"
	original := "AIzaSyTest1234567890abcdef"

	encrypted, err := encryptKey(original, password)
	if err != nil {
		t.Fatalf("encryptKey 실패: %v", err)
	}
	if encrypted == original {
		t.Fatal("암호화된 값이 원본과 같음")
	}
	// v2 prefix must be present
	if !strings.HasPrefix(encrypted, v2Prefix) {
		t.Fatalf("v2 prefix 없음: %q", encrypted)
	}

	decrypted, err := decryptKey(encrypted, password)
	if err != nil {
		t.Fatalf("decryptKey 실패: %v", err)
	}
	if decrypted != original {
		t.Fatalf("복호화 불일치: want %q, got %q", original, decrypted)
	}
}

func TestEncryptDecrypt_NoPassword(t *testing.T) {
	original := "plain-api-key"

	encrypted, err := encryptKey(original, "")
	if err != nil {
		t.Fatalf("encryptKey(no pass) 실패: %v", err)
	}
	if encrypted != original {
		t.Fatalf("비밀번호 없으면 평문 저장 기대: got %q", encrypted)
	}

	decrypted, err := decryptKey(encrypted, "")
	if err != nil {
		t.Fatalf("decryptKey(no pass) 실패: %v", err)
	}
	if decrypted != original {
		t.Fatalf("복호화 불일치: want %q, got %q", original, decrypted)
	}
}

func TestEncryptDecrypt_WrongPassword(t *testing.T) {
	original := "secret-key"
	encrypted, _ := encryptKey(original, "correct-password")

	_, err := decryptKey(encrypted, "wrong-password")
	if err == nil {
		t.Fatal("잘못된 비밀번호로 복호화 성공 — 오류 기대")
	}
}

func TestEncryptDecrypt_Randomness(t *testing.T) {
	password := "same-password"
	original := "same-key"

	enc1, _ := encryptKey(original, password)
	enc2, _ := encryptKey(original, password)

	// different salt → different ciphertext every time
	if enc1 == enc2 {
		t.Fatal("salt 랜덤화 미작동 — 같은 암호문이 두 번 생성됨")
	}

	dec1, _ := decryptKey(enc1, password)
	dec2, _ := decryptKey(enc2, password)
	if dec1 != original || dec2 != original {
		t.Fatal("랜덤 salt 후 복호화 실패")
	}
}

func TestDecryptKey_PlaintextFallback(t *testing.T) {
	// non-base64 plaintext with password → returned as-is (legacy fallback)
	plain := "not-base64-!!!"
	result, err := decryptKey(plain, "some-password")
	if err != nil {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
	if result != plain {
		t.Fatalf("평문 폴백 기대, got %q", result)
	}
}

// TestMigration verifies that legacy SHA-256 encrypted data can be decrypted
// and that new encryptions use the v2 (Argon2id) scheme.
func TestMigration_LegacyDecrypt(t *testing.T) {
	password := "migration-test-password"
	original := "legacy-api-key-value"

	// encrypt with legacy scheme directly
	legacyEnc, err := encryptLegacy(original, password)
	if err != nil {
		t.Fatalf("legacy encrypt 실패: %v", err)
	}
	if strings.HasPrefix(legacyEnc, v2Prefix) {
		t.Fatal("legacy 암호문에 v2 prefix가 있으면 안 됨")
	}

	// decryptKey must transparently handle legacy format
	decrypted, err := decryptKey(legacyEnc, password)
	if err != nil {
		t.Fatalf("legacy 복호화 실패: %v", err)
	}
	if decrypted != original {
		t.Fatalf("legacy 복호화 불일치: want %q, got %q", original, decrypted)
	}
}

func TestIsLegacyEncrypted(t *testing.T) {
	password := "test"
	original := "api-key"

	v2, _ := encryptKey(original, password)
	legacy, _ := encryptLegacy(original, password)

	if isLegacyEncrypted(v2) {
		t.Error("v2 암호문이 legacy로 판정됨")
	}
	if !isLegacyEncrypted(legacy) {
		t.Error("legacy 암호문이 legacy로 판정 안 됨")
	}
	if isLegacyEncrypted("not-base64-!!!") {
		t.Error("평문이 legacy로 판정됨")
	}
	if isLegacyEncrypted("") {
		t.Error("빈 문자열이 legacy로 판정됨")
	}
}

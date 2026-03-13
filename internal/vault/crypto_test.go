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

	if enc1 == enc2 {
		t.Fatal("nonce 랜덤화 미작동 — 같은 암호문이 두 번 생성됨")
	}

	dec1, _ := decryptKey(enc1, password)
	dec2, _ := decryptKey(enc2, password)
	if dec1 != original || dec2 != original {
		t.Fatal("랜덤 nonce 후 복호화 실패")
	}
}

func TestDecryptKey_PlaintextFallback(t *testing.T) {
	// if non-base64 plaintext is given, return it as-is
	plain := "not-base64-!!!"
	result, err := decryptKey(plain, "some-password")
	if err != nil {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
	// base64 decode failure → return plaintext
	if !strings.Contains(result, plain) && result != plain {
		t.Fatalf("평문 폴백 기대, got %q", result)
	}
}

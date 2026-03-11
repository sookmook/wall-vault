package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

func deriveKey(password string) []byte {
	h := sha256.Sum256([]byte(password))
	return h[:]
}

func encryptKey(plaintext, password string) (string, error) {
	if password == "" {
		return plaintext, nil // 마스터 비밀번호 없으면 평문 저장
	}
	key := deriveKey(password)
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

func decryptKey(ciphertext64, password string) (string, error) {
	if password == "" {
		return ciphertext64, nil
	}
	key := deriveKey(password)
	data, err := base64.StdEncoding.DecodeString(ciphertext64)
	if err != nil {
		// base64 디코드 실패 = 이미 평문
		return ciphertext64, nil
	}
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
		return ciphertext64, nil // 짧으면 평문으로 취급
	}
	nonce, ct := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("키 복호화 실패: %w", err)
	}
	return string(plain), nil
}

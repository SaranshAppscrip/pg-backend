package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// FieldEncryptor encrypts sensitive field values at rest using AES-256-GCM.
type FieldEncryptor struct {
	gcm cipher.AEAD
}

func NewFieldEncryptor(secret string) (*FieldEncryptor, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("encryption secret is required")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &FieldEncryptor{gcm: gcm}, nil
}

func (e *FieldEncryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *FieldEncryptor) Decrypt(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	nonceSize := e.gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plain, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// Last4 returns the last four digits from an ID number string.
func Last4(id string) string {
	var digits strings.Builder
	for _, r := range id {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}
	s := digits.String()
	if len(s) >= 4 {
		return s[len(s)-4:]
	}
	return s
}

// MaskIDNumber formats a masked display value from the last four digits.
func MaskIDNumber(last4 string) string {
	if last4 == "" {
		return ""
	}
	return "XXXX-XXXX-" + last4
}

package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Service provides AES-256-GCM encryption for sensitive data such as mailbox
// passwords. The key must be exactly 32 bytes (base64-encoded or raw).
type Service struct {
	key []byte
}

// NewService creates an encryption service from a base64-encoded 32-byte key.
func NewService(key string) (*Service, error) {
	if key == "" {
		return nil, errors.New("encryption key is empty")
	}
	k, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		// Fall back to raw key if not valid base64.
		k = []byte(key)
	}
	if len(k) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(k))
	}
	return &Service{key: k}, nil
}

// Encrypt returns a base64-encoded ciphertext for the given plaintext.
func (s *Service) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.key)
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

// Decrypt decrypts a base64-encoded ciphertext produced by Encrypt.
func (s *Service) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, encrypted := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// Package tenant gerencia configurações por tenant (multi-tenancy leve).
// Cada usuário tem suas próprias credenciais de API, destinos de alerta, etc.
package tenant

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// deriveKey deriva uma chave AES-256 (32 bytes) a partir de uma passphrase.
func deriveKey(passphrase string) []byte {
	h := sha256.Sum256([]byte(passphrase))
	return h[:]
}

// encryptionKey retorna a chave de criptografia do ambiente.
// Se ENCRYPTION_KEY não estiver definida, usa um fallback inseguro (só para dev).
func encryptionKey() string {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		// Fallback inseguro — aceitável apenas em dev local.
		// Em produção, ENCRYPTION_KEY deve estar definida no Cloud Run.
		return "garimpei-dev-insecure-key-do-not-use-in-prod"
	}
	return key
}

// Encrypt cifra um valor usando AES-256-GCM com a chave do ambiente.
// Retorna o ciphertext codificado em base64.
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	key := deriveKey(encryptionKey())
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
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decifra um valor criptografado por Encrypt.
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	key := deriveKey(encryptionKey())
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
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
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

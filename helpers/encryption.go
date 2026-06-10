// Package helpers provides utility types and functions used across the
// Adele framework, including symmetric encryption and file upload helpers.
package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// Encryption performs AES symmetric encryption and decryption using the
// provided Key. Key must be a valid AES key length (16, 24, or 32 bytes,
// for AES-128, AES-192, or AES-256 respectively).
type Encryption struct {
	Key []byte
}

// Encrypt encrypts the given plaintext using AES in CFB mode and returns the
// result as a base64 URL-encoded string. A random initialization vector is
// generated for each call and prepended to the ciphertext. It returns an error
// if the key is invalid or if reading random bytes for the IV fails.
func (e *Encryption) Encrypt(text string) (string, error) {
	plaintext := []byte(text)

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(plaintext))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts a base64 URL-encoded string previously produced by Encrypt
// and returns the original plaintext. It expects the initialization vector to
// be prepended to the ciphertext. It returns an error if the key is invalid.
func (e *Encryption) Decrypt(cryptoText string) (string, error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

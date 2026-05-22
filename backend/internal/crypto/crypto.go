package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"log/slog"

	"github.com/yunuskargi/configbox/internal/config"
)

var key []byte

func Init() {
	h := sha256.Sum256([]byte(config.EncryptionKey))
	key = h[:]
}

func Encrypt(plaintext string) string {
	if plaintext == "" {
		return plaintext
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		slog.Error("crypto: failed to create cipher for encryption", "error", err)
		return plaintext
	}

	padded := pkcs7Pad([]byte(plaintext), aes.BlockSize)

	iv := make([]byte, aes.BlockSize)
	rand.Read(iv)

	mode := cipher.NewCBCEncrypter(block, iv)
	ct := make([]byte, len(padded))
	mode.CryptBlocks(ct, padded)

	result := append(iv, ct...)
	return base64.StdEncoding.EncodeToString(result)
}

func Decrypt(ciphertext string) string {
	if ciphertext == "" {
		return ciphertext
	}

	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		slog.Warn("crypto: failed to base64 decode, returning as-is (may be plaintext)")
		return ciphertext
	}

	if len(raw) < 32 {
		slog.Warn("crypto: ciphertext too short, returning as-is (may be plaintext)")
		return ciphertext
	}

	iv := raw[:16]
	ct := raw[16:]

	block, err := aes.NewCipher(key)
	if err != nil {
		slog.Error("crypto: failed to create cipher for decryption", "error", err)
		return ciphertext
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plainBytes := make([]byte, len(ct))
	mode.CryptBlocks(plainBytes, ct)

	unpadded, err := pkcs7Unpad(plainBytes, aes.BlockSize)
	if err != nil {
		slog.Warn("crypto: unpad failed, data may be corrupted or wrong key")
		return ciphertext
	}

	return string(unpadded)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padByte := byte(padding)
	for i := 0; i < padding; i++ {
		data = append(data, padByte)
	}
	return data
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, aes.KeySizeError(len(data))
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize {
		return nil, aes.KeySizeError(padLen)
	}
	for i := len(data) - padLen; i < len(data); i++ {
		if data[i] != byte(padLen) {
			return nil, aes.KeySizeError(padLen)
		}
	}
	return data[:len(data)-padLen], nil
}

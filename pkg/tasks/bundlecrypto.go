package tasks

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/hkdf"

	"github.com/xbapps/xbvr/pkg/models"
)

// EncryptedValuePrefix marks KV values in a backup bundle whose plaintext is
// encrypted with a key derived from the user-supplied bundle password via
// HKDF-SHA256.
const EncryptedValuePrefix = "xbvrenc:v2:"

// IsEncryptedBundleValue reports whether an exported KV value is encrypted.
// Unencrypted bundles contain no such values, which is how encrypted and
// unencrypted bundles are told apart.
func IsEncryptedBundleValue(value string) bool {
	return strings.HasPrefix(value, EncryptedValuePrefix)
}

// IsCredentialKVKey reports whether a KV-store key holds credentials or other
// secrets: the app config (API keys, tokens, password hashes) and scraper
// HTTP configs (cookies/headers, stored under "<domain>-trailers" and
// "<domain>-scraper" keys).
func IsCredentialKVKey(key string) bool {
	return key == "config" ||
		strings.HasSuffix(key, "-trailers") ||
		strings.HasSuffix(key, "-scraper") ||
		strings.Contains(key, "cookie") ||
		strings.Contains(key, "header")
}

// deriveBundleKey derives the AES-256 key from the bundle password with
// HKDF-SHA256, using a random per-value salt.
func deriveBundleKey(password string, salt []byte) ([]byte, error) {
	key := make([]byte, 32)
	_, err := io.ReadFull(hkdf.New(sha256.New, []byte(password), salt, []byte("xbvr-bundle")), key)
	return key, err
}

// encryptValue seals plaintext with a raw AES-256-GCM key and returns the
// given prefix plus base64(salt || nonce || ciphertext).
func encryptValue(key, salt []byte, prefix, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload := append(salt, nonce...)
	payload = gcm.Seal(payload, nonce, []byte(plaintext), nil)
	return prefix + base64.StdEncoding.EncodeToString(payload), nil
}

func decryptValue(key, payload []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", errors.New("encrypted value too short")
	}
	nonce, ciphertext := payload[:gcm.NonceSize()], payload[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("cannot decrypt value — wrong bundle password?")
	}
	return string(plaintext), nil
}

// EncryptBundleSecret encrypts a value with the user-supplied bundle
// password (HKDF-SHA256-derived AES-256-GCM) and returns an
// EncryptedValuePrefix-tagged base64 string.
func EncryptBundleSecret(password, plaintext string) (string, error) {
	if password == "" {
		return "", errors.New("bundle password required to encrypt credentials")
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := deriveBundleKey(password, salt)
	if err != nil {
		return "", err
	}
	return encryptValue(key, salt, EncryptedValuePrefix, plaintext)
}

// DecryptBundleSecret decrypts an exported bundle value with the bundle
// password. Unencrypted values are rejected — callers restore those as-is.
func DecryptBundleSecret(password, encoded string) (string, error) {
	if !strings.HasPrefix(encoded, EncryptedValuePrefix) {
		return "", errors.New("not an encrypted bundle value")
	}
	if password == "" {
		return "", errors.New("bundle password required")
	}
	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encoded, EncryptedValuePrefix))
	if err != nil {
		return "", err
	}
	if len(payload) < 16+12 {
		return "", errors.New("encrypted value too short")
	}
	key, err := deriveBundleKey(password, payload[:16])
	if err != nil {
		return "", err
	}
	return decryptValue(key, payload[16:])
}

// VerifyBundlePassword pre-flights the bundle password against every
// encrypted KV value in a bundle. It returns nil when nothing is encrypted
// (unencrypted bundles restore without a password) and an error when the
// password is missing or fails to decrypt any value — callers must abort
// the restore instead of silently skipping fields.
func VerifyBundlePassword(kvs []models.KV, bundlePassword string) error {
	for _, kv := range kvs {
		if !IsEncryptedBundleValue(kv.Value) {
			continue
		}
		if bundlePassword == "" {
			return errors.New("bundle contains password-encrypted credential settings — provide the bundle password")
		}
		if _, err := DecryptBundleSecret(bundlePassword, kv.Value); err != nil {
			return fmt.Errorf("bundle password does not match (cannot decrypt setting %q)", kv.Key)
		}
	}
	return nil
}

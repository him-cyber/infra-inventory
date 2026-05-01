package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type SecureCookie struct {
	name   string
	aead   cipher.AEAD
	secure bool
}

func NewSecureCookie(name, key string, secure bool) (*SecureCookie, error) {
	secret, err := decodeKey(key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SecureCookie{name: name, aead: aead, secure: secure}, nil
}

func (c *SecureCookie) Set(w http.ResponseWriter, value any, expires time.Time) error {
	encoded, err := c.Seal(value)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     c.name,
		Value:    encoded,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (c *SecureCookie) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c *SecureCookie) Read(r *http.Request, value any) error {
	cookie, err := r.Cookie(c.name)
	if err != nil {
		return err
	}
	return c.Open(cookie.Value, value)
}

func (c *SecureCookie) Seal(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nonce, nonce, data, nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (c *SecureCookie) Open(encoded string, value any) error {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	if len(data) < c.aead.NonceSize() {
		return errors.New("secure cookie payload is too short")
	}
	nonce := data[:c.aead.NonceSize()]
	ciphertext := data[c.aead.NonceSize():]
	plain, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(plain, value)
}

func decodeKey(key string) ([]byte, error) {
	if key == "" {
		sum := sha256.Sum256([]byte("local-dev-infra-inventory-stream-session-key"))
		return sum[:], nil
	}
	data, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	if len(data) != 32 {
		return nil, errors.New("AUTH_SESSION_KEY must decode to 32 bytes")
	}
	return data, nil
}

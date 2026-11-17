// Package crypter provides AES-GCM authenticated string encryption, tagged ENCv1[...]; idempotent, thread-safe.
package crypter

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
)

const (
	// StartTag is the ENCv1[ prefix marking a string as encrypted by this package; plaintext must never start with it.
	StartTag = "ENCv1["

	// EndTag is the closing "]"; quote encrypted values in YAML/JSON so brackets are not parsed as a flow sequence.
	EndTag = "]"
)

var (
	// ErrInvalidCipherText marks a malformed ciphertext (bad end tag, short payload); wrong key failures use ErrOpen.
	ErrInvalidCipherText = errors.New("crypter: invalid ciphertext")

	// ErrInvalidKeyLength is returned by New when secret is not 16, 24, or 32 bytes (AES-128/192/256), measured in bytes.
	ErrInvalidKeyLength = errors.New("crypter: invalid key length")

	// ErrEmptyPayload is returned for a tagged string with no base64 payload, i.e. exactly "ENCv1[]".
	ErrEmptyPayload = errors.New("crypter: empty payload")

	// ErrNewCipher wraps an aes.NewCipher failure; should not occur since New already validates key length.
	ErrNewCipher = errors.New("crypter: aes.NewCipher failed")

	// ErrNewGCM wraps a cipher.NewGCM failure; should not occur in practice.
	ErrNewGCM = errors.New("crypter: cipher.NewGCM failed")

	// ErrReadNonce is returned when crypto/rand fails during encryption (an unrecoverable OS-level error).
	ErrReadNonce = errors.New("crypter: read nonce failed")

	// ErrBase64Decode is returned when the payload between the tags is not valid base64url.
	ErrBase64Decode = errors.New("crypter: base64 decode failed")

	// ErrOpen wraps an AES-GCM auth/decrypt failure (wrong key, tampered ciphertext or tag); unwrap for the original.
	ErrOpen = errors.New("crypter: aead open failed")
)

// startLen is StartTag byte length, cached to avoid repeated len() calls in IsEncrypted hot path.
const startLen = len(StartTag)

// Crypter encrypts and decrypts strings via AES-GCM; construct with New, reuse across goroutines; zero value unusable.
type Crypter struct {
	// aead is the AES-GCM AEAD instance from the caller key; safe for concurrent use.
	aead cipher.AEAD

	// nonceSize is aead.NonceSize(), cached to avoid repeated interface calls; always 12 for standard AES-GCM.
	nonceSize int

	// noncePool pools *[]byte nonce buffers per Encrypt call; pointers avoid heap-allocating the interface box.
	noncePool sync.Pool
}

// New initializes a Crypter from a 16/24/32-byte AES key (AES-128/192/256), used directly with no hashing.
func New(secret string) (*Crypter, error) {
	key := []byte(secret)
	switch len(key) {
	case 16, 24, 32:
	default:
		return nil, ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewCipher, err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewGCM, err)
	}

	nonceSize := aead.NonceSize()
	return &Crypter{
		aead:      aead,
		nonceSize: nonceSize,
		// Stores *[]byte (pointer-sized) so pool.Put never heap-allocates the interface box.
		noncePool: sync.Pool{New: func() any { b := make([]byte, nonceSize); return &b }},
	}, nil
}

// IsEncrypted is a fast heuristic (checks only StartTag prefix, no decoding); plaintext must never start with it.
func (c *Crypter) IsEncrypted(s string) bool {
	return len(s) > startLen && s[:startLen] == StartTag
}

// StartTag returns the StartTag constant so callers like crypter/yaml can obtain the tag via interface method.
func (c *Crypter) StartTag() string {
	return StartTag
}

// Encrypt seals data into ENCv1[...]; tagged input returns unchanged, so plaintext must never start with ENCv1[.
func (c *Crypter) Encrypt(data string) (string, error) {
	if c.IsEncrypted(data) {
		return data, nil
	}

	noncep, ok := c.noncePool.Get().(*[]byte)
	if !ok {
		// should never happen; pool.New always returns *[]byte
		b := make([]byte, c.nonceSize)
		noncep = &b
	}
	if _, err := io.ReadFull(rand.Reader, *noncep); err != nil {
		c.noncePool.Put(noncep)
		return "", fmt.Errorf("%w: %w", ErrReadNonce, err)
	}
	// Copies random bytes into an owned slice, returns the pool buffer immediately; nothing below reuses it.
	nonce := make([]byte, c.nonceSize)
	copy(nonce, *noncep)
	c.noncePool.Put(noncep)

	// Pre-allocates raw with the nonce prefix, then Seal appends ciphertext+tag in place; raw and nonce do not alias.
	raw := make([]byte, 0, c.nonceSize+len(data)+c.aead.Overhead())
	raw = append(raw, nonce...)
	raw = c.aead.Seal(raw, nonce, []byte(data), nil)

	encodedLen := base64.RawURLEncoding.EncodedLen(len(raw))
	buf := make([]byte, startLen, startLen+encodedLen+len(EndTag))
	copy(buf, StartTag)
	buf = base64.RawURLEncoding.AppendEncode(buf, raw)
	buf = append(buf, EndTag...)
	return string(buf), nil
}

// Decrypt reverses Encrypt: untagged input passes through unchanged, tagged input is base64-decoded and AES-GCM-opened.
func (c *Crypter) Decrypt(data string) (string, error) {
	if !c.IsEncrypted(data) {
		return data, nil
	}

	// Validate end tag and extract the base64url payload.
	if data[len(data)-1] != EndTag[0] {
		return "", ErrInvalidCipherText
	}
	payload := data[startLen : len(data)-1] // slice only, no alloc
	if payload == "" {
		return "", ErrEmptyPayload
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrBase64Decode, err)
	}

	if len(raw) < c.nonceSize {
		return "", ErrInvalidCipherText
	}

	nonce, ct := raw[:c.nonceSize], raw[c.nonceSize:]
	if len(ct) < c.aead.Overhead() {
		// Payload has a valid nonce but no room for the GCM authentication tag.
		return "", ErrInvalidCipherText
	}
	pt, err := c.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		// Map all auth/decrypt failures to a stable sentinel, but keep original for debugging via wrapping.
		return "", fmt.Errorf("%w: %w", ErrOpen, err)
	}
	return string(pt), nil
}

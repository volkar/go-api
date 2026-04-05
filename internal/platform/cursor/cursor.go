package cursor

import (
	"api/internal/platform/response"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Cursor struct {
	aead cipher.AEAD
}

/* Initialize a new cursor encryptor */
func New(key []byte, logger *slog.Logger) *Cursor {
	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Error("Cannot create cursor encryptor", "err", err)
		os.Exit(1)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("Cannot create cursor encryptor", "err", err)
		os.Exit(1)
	}

	return &Cursor{aead: aesGCM}
}

/* Create a base64 encoded string from timestamp and id */
func (s *Cursor) Encode(t time.Time, id string) (string, error) {
	// Create the raw payload
	payload := fmt.Sprintf("%s|%s", t.Format(time.RFC3339Nano), id)
	plaintext := []byte(payload)

	// Create a random nonce (initialization vector)
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and authenticate the payload. The result is appended to the nonce
	ciphertext := s.aead.Seal(nonce, nonce, plaintext, nil)

	// Return as a URL-safe Base64 string
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

/* Extract timestamp and id from the base64 cursor */
func (s *Cursor) Decode(encoded string) (time.Time, string, error) {
	if encoded == "" {
		return time.Time{}, "", nil
	}

	// Decode Base64
	ciphertext, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return time.Time{}, "", response.ErrInvalidCursor
	}

	// Extract nonce
	nonceSize := s.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return time.Time{}, "", response.ErrInvalidCursor
	}
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and verify integrity. If someone tampered with the string, Open() will return an error
	plaintext, err := s.aead.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return time.Time{}, "", response.ErrInvalidCursor
	}

	// Parse the original payload
	parts := strings.SplitN(string(plaintext), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", response.ErrInvalidCursor
	}

	parsedTime, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", response.ErrInvalidCursor
	}

	// Return timestamp and id
	return parsedTime, parts[1], nil
}

func (s *Cursor) Parse(cursor string) (pgtype.Timestamptz, pgtype.UUID, error) {
	var cursorDate pgtype.Timestamptz
	var cursorID pgtype.UUID
	if cursor != "" {
		parsedTime, parsedIDStr, err := s.Decode(cursor)
		if err != nil {
			return pgtype.Timestamptz{}, pgtype.UUID{}, err
		}
		cursorDate = pgtype.Timestamptz{Time: parsedTime, Valid: true}
		uid, err := uuid.Parse(parsedIDStr)
		if err != nil {
			return pgtype.Timestamptz{}, pgtype.UUID{}, err
		}
		cursorID = pgtype.UUID{Bytes: uid, Valid: true}
	}
	return cursorDate, cursorID, nil
}

package tokens

import (
	db "api/internal/platform/database/sqlc"
	"api/internal/platform/request"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
)

type Manager struct {
	key    paseto.V4SymmetricKey
	parser *paseto.Parser
	tokens *Repository
}

func NewService(secretKey string, repo *Repository, logger *slog.Logger) *Manager {
	// Use Paseto token with secret key
	key, err := paseto.V4SymmetricKeyFromBytes([]byte(secretKey))
	if err != nil {
		logger.Error("Failed to init token manager", "err", err)
		os.Exit(1)
	}

	parser := paseto.NewParser()

	return &Manager{
		key:    key,
		parser: &parser,
		tokens: repo,
	}
}

type ctxKey string

const userClaimsKey ctxKey = "user"

type UserClaims struct {
	UserID uuid.UUID
	Email  string
	Role   string
}

/* Generate random refresh token string */
func (m *Manager) GenerateRefreshString() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

/* Hash refresh token string */
func (m *Manager) Hash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

/* Get refresh token by hash */
func (m *Manager) GetRefreshByHash(ctx context.Context, hash string) (db.RefreshToken, error) {
	return m.tokens.GetByHash(ctx, hash)
}

/* Create refresh token */
func (m *Manager) CreateRefresh(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time, meta request.Metadata) (uuid.UUID, error) {
	return m.tokens.Create(ctx, userID, tokenHash, expiresAt, meta.IP, meta.UserAgent, meta.Device, meta.Os, meta.Browser, meta.Location)
}

/* Delete old token and create new one in single transaction */
func (m *Manager) ReplaceRefresh(ctx context.Context, userID uuid.UUID, oldHash string, newHash string, expiresAt time.Time, meta request.Metadata) error {
	return m.tokens.ReplaceInTransaction(ctx, userID, oldHash, newHash, expiresAt, meta.IP, meta.UserAgent, meta.Device, meta.Os, meta.Browser, meta.Location)
}

/* Consume one token by hash */
func (m *Manager) ConsumeRefreshByHash(ctx context.Context, hash string) error {
	return m.tokens.ConsumeByHash(ctx, hash)
}

/* Consume all tokens for given user id */
func (m *Manager) ConsumeAllRefreshForUser(ctx context.Context, userID uuid.UUID) error {
	return m.tokens.ConsumeOtherForUser(ctx, userID, "")
}

/* Consume other tokens for given user id except given hash */
func (m *Manager) ConsumeOtherRefreshForUser(ctx context.Context, userID uuid.UUID, hash string) error {
	return m.tokens.ConsumeOtherForUser(ctx, userID, hash)
}

/* Delete all tokens for given user id */
func (m *Manager) DeleteAllRefreshForUser(ctx context.Context, userID uuid.UUID) error {
	return m.tokens.DeleteAllRefreshForUser(ctx, userID)
}

/* Create access token string */
func (m *Manager) CreateAccess(userID uuid.UUID, role string, email string, ttl time.Duration) string {
	token := paseto.NewToken()

	now := time.Now()
	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetExpiration(now.Add(ttl))

	// Set claims
	token.SetString("user_id", userID.String())
	token.SetString("role", role)
	token.SetString("email", email)

	return token.V4Encrypt(m.key, nil)
}

/* Parse Paseto access token string */
func (m *Manager) ParseAccess(tokenStr string) (UserClaims, error) {
	token, err := m.parser.ParseV4Local(m.key, tokenStr, nil)
	if err != nil {
		return UserClaims{}, err
	}

	userID, err := token.GetString("user_id")
	if err != nil {
		return UserClaims{}, err
	}

	role, err := token.GetString("role")
	if err != nil {
		return UserClaims{}, err
	}

	email, err := token.GetString("email")
	if err != nil {
		return UserClaims{}, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return UserClaims{}, err
	}

	return UserClaims{
		UserID: userUUID,
		Role:   role,
		Email:  email,
	}, nil
}

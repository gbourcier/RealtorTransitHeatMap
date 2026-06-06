package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrBadCredentials  = errors.New("auth: invalid username or password")
	ErrInactive        = errors.New("auth: account is inactive")
	ErrSessionExpired  = errors.New("auth: session expired")
	ErrPasswordTooLong = errors.New("auth: password exceeds 72 bytes")
)

type Service struct {
	users     *user.Repository
	sessions  *sessionRepository
	cfg       config.AuthConfig
	dummyHash []byte
}

func NewService(db *gorm.DB, users *user.Repository, cfg config.AuthConfig) *Service {
	dummy, _ := bcrypt.GenerateFromPassword([]byte("bcrypt-timing-equalizer"), bcrypt.DefaultCost)
	return &Service{
		users:     users,
		sessions:  &sessionRepository{db: db},
		cfg:       cfg,
		dummyHash: dummy,
	}
}

func (svc *Service) HashPassword(plain string) (string, error) {
	if len(plain) > maxPasswordLength {
		return "", ErrPasswordTooLong
	}
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func (svc *Service) Login(ctx context.Context, username, password string) (string, *Session, *user.User, error) {
	u, err := svc.users.GetByUsername(ctx, username)
	if errors.Is(err, user.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(svc.dummyHash, []byte(password))
		return "", nil, nil, ErrBadCredentials
	}
	if err != nil {
		return "", nil, nil, err
	}

	compareErr := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if !u.IsActive {
		return "", nil, nil, ErrInactive
	}
	if compareErr != nil {
		return "", nil, nil, ErrBadCredentials
	}

	raw, err := generateToken()
	if err != nil {
		return "", nil, nil, err
	}

	sess := &Session{
		Token:     hashToken(raw),
		UserID:    u.ID,
		ExpiresAt: time.Now().Add(svc.cfg.SessionTTL),
	}
	if err := svc.sessions.create(ctx, sess); err != nil {
		return "", nil, nil, err
	}
	return raw, sess, u, nil
}

func (svc *Service) Validate(ctx context.Context, token string) (*user.User, error) {
	hashed := hashToken(token)
	sess, err := svc.sessions.get(ctx, hashed)
	if err != nil {
		return nil, ErrSessionExpired
	}
	if time.Now().After(sess.ExpiresAt) {
		_ = svc.sessions.delete(ctx, hashed)
		return nil, ErrSessionExpired
	}
	u, err := svc.users.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, ErrSessionExpired
	}
	if !u.IsActive {
		return nil, ErrInactive
	}
	return u, nil
}

func (svc *Service) Logout(ctx context.Context, token string) error {
	return svc.sessions.delete(ctx, hashToken(token))
}

func (svc *Service) MarkSeen(ctx context.Context, id uuid.UUID) {
	if err := svc.users.TouchLastSeen(ctx, id, lastSeenThrottle); err != nil {
		slog.Warn("touch last_seen", "err", err)
	}
}

func (svc *Service) PurgeExpiredSessions(ctx context.Context) (int64, error) {
	return svc.sessions.deleteExpired(ctx, time.Now())
}

func (svc *Service) Bootstrap(ctx context.Context) error {
	if len(svc.cfg.AdminPassword) < minPasswordLength {
		slog.Warn("admin password is shorter than recommended minimum", "min", minPasswordLength)
	}

	existing, err := svc.users.GetByUsername(ctx, svc.cfg.AdminUsername)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		return err
	}

	hash, err := svc.HashPassword(svc.cfg.AdminPassword)
	if err != nil {
		return err
	}

	if errors.Is(err, user.ErrNotFound) || existing == nil {
		u := &user.User{
			ID:           uuid.New(),
			Username:     svc.cfg.AdminUsername,
			PasswordHash: hash,
			Role:         user.RoleAdmin,
			IsActive:     true,
		}
		if createErr := svc.users.Create(ctx, u); createErr != nil {
			return createErr
		}
		slog.Info("admin user created", "username", svc.cfg.AdminUsername)
		return nil
	}

	if svc.cfg.AdminReset {
		ph := hash
		_, updateErr := svc.users.Update(ctx, existing.ID, user.Patch{PasswordHash: &ph})
		if updateErr != nil {
			return updateErr
		}
		slog.Info("admin password reset", "username", svc.cfg.AdminUsername)
	}
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

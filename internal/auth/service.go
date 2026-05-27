package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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
	ErrBadCredentials = errors.New("auth: invalid username or password")
	ErrInactive       = errors.New("auth: account is inactive")
	ErrSessionExpired = errors.New("auth: session expired")
)

type Service struct {
	users    *user.Repository
	sessions *sessionRepository
	cfg      config.AuthConfig
}

func NewService(db *gorm.DB, users *user.Repository, cfg config.AuthConfig) *Service {
	return &Service{
		users:    users,
		sessions: &sessionRepository{db: db},
		cfg:      cfg,
	}
}

func (svc *Service) HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func (svc *Service) Login(ctx context.Context, username, password string) (*Session, *user.User, error) {
	u, err := svc.users.GetByUsername(ctx, username)
	if errors.Is(err, user.ErrNotFound) {
		return nil, nil, ErrBadCredentials
	}
	if err != nil {
		return nil, nil, err
	}
	if !u.IsActive {
		return nil, nil, ErrInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrBadCredentials
	}

	token, err := generateToken()
	if err != nil {
		return nil, nil, err
	}

	sess := &Session{
		Token:     token,
		UserID:    u.ID,
		ExpiresAt: time.Now().Add(svc.cfg.SessionTTL),
	}
	if err := svc.sessions.create(ctx, sess); err != nil {
		return nil, nil, err
	}
	return sess, u, nil
}

func (svc *Service) Validate(ctx context.Context, token string) (*user.User, error) {
	sess, err := svc.sessions.get(ctx, token)
	if err != nil {
		return nil, ErrSessionExpired
	}
	if time.Now().After(sess.ExpiresAt) {
		_ = svc.sessions.delete(ctx, token)
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
	return svc.sessions.delete(ctx, token)
}

func (svc *Service) Bootstrap(ctx context.Context) error {
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

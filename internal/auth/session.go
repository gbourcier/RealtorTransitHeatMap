package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	Token     string    `gorm:"column:token;primaryKey"`
	UserID    uuid.UUID `gorm:"column:user_id;not null"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Session) TableName() string { return "sessions" }

type sessionRepository struct{ db *gorm.DB }

func (r *sessionRepository) create(ctx context.Context, s *Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *sessionRepository) get(ctx context.Context, token string) (*Session, error) {
	var s Session
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *sessionRepository) delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&Session{}).Error
}

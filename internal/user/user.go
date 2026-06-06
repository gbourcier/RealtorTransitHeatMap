package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("user: not found")

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	ID           uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	Username     string     `gorm:"column:username;not null;uniqueIndex"`
	PasswordHash string     `gorm:"column:password_hash;not null"`
	Role         string     `gorm:"column:role;not null;default:user"`
	IsActive     bool       `gorm:"column:is_active;not null;default:true"`
	LastSeenAt   *time.Time `gorm:"column:last_seen_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (User) TableName() string { return "users" }

type Patch struct {
	Role         *string
	IsActive     *bool
	PasswordHash *string
}

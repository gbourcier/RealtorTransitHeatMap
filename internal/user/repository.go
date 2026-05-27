package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *Repository) List(ctx context.Context) ([]User, error) {
	var out []User
	err := r.db.WithContext(ctx).Order("created_at ASC").Find(&out).Error
	return out, err
}

func (r *Repository) Create(ctx context.Context, u *User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, patch Patch) (*User, error) {
	updates := map[string]any{}
	if patch.Role != nil {
		updates["role"] = *patch.Role
	}
	if patch.IsActive != nil {
		updates["is_active"] = *patch.IsActive
	}
	if patch.PasswordHash != nil {
		updates["password_hash"] = *patch.PasswordHash
	}
	if len(updates) == 0 {
		return r.GetByID(ctx, id)
	}
	res := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&User{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) CountAdmins(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("role = ? AND is_active = TRUE", RoleAdmin).Count(&n).Error
	return n, err
}

package preference

import (
	"time"

	"github.com/google/uuid"
)

type Preference struct {
	UserID          uuid.UUID  `gorm:"column:user_id;primaryKey"`
	DefaultFilterID *uuid.UUID `gorm:"column:default_filter_id"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (Preference) TableName() string { return "user_preferences" }

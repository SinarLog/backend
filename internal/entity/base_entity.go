package entity

import (
	"time"

	"gorm.io/gorm"
)

type BaseModelID struct {
	ID string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
}

type BaseModelStamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BaseModelSoftDelete struct {
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

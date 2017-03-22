package models

import (
	"github.com/jinzhu/gorm"
)

// User model.
type User struct {
	gorm.Model

	Username string `gorm:"unique_index"`
	Password string `gorm:"not null"`
}

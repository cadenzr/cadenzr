package models

import (
	"github.com/jinzhu/gorm"
)

// Image model.
type Image struct {
	gorm.Model

	Path string `gorm:"not null; unique"`
	Link string `gorm:"not null; unique"`
	Mime string `gorm:"not null"`
	Hash string `gorm:"not null; unique"`
}

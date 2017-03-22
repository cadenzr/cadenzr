package models

import (
	"github.com/jinzhu/gorm"
)

// Artist model.
type Artist struct {
	gorm.Model

	Name string `gorm:"unique_index"`
}

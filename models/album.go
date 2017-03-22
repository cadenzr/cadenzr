package models

import (
	"github.com/jinzhu/gorm"
)

// Album model.
type Album struct {
	gorm.Model

	Name string `gorm:"not null,unique_index"`

	Cover   *Image `gorm:"ForeignKey:CoverID"`
	CoverID NullInt64

	Year NullInt64
}

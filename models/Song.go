package models

import (
	"github.com/jinzhu/gorm"
)

// Song model.
type Song struct {
	gorm.Model

	Name string `gorm:"not null"`

	Artist   *Artist `gorm:"ForeignKey:ArtistID"`
	ArtistID NullInt64

	Album   *Album `gorm:"ForeignKey:AlbumID"`
	AlbumID NullInt64

	Cover   *Image `gorm:"ForeignKey:CoverID"`
	CoverID NullInt64

	Year NullInt64

	Genre    NullString
	Duration NullFloat64
	Mime     string `gorm:"not null"`
	Path     string `gorm:"not null"`
	Played   uint   `gorm:"not null"`
}

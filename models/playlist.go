package models

import (
	"github.com/jinzhu/gorm"
)

// Playlist model.
type Playlist struct {
	gorm.Model

	Name  string  `gorm:"not null,unique_index"`
	Songs []*Song `gorm:"many2many:playlist_songs"`
}

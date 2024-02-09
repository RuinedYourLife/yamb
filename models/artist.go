package models

import "gorm.io/gorm"

type Artist struct {
	gorm.Model
	Name      string `gorm:"not null"`
	SpotifyID string `gorm:"not null;unique"`
}

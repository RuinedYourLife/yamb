package models

import (
	"time"

	"gorm.io/gorm"
)

type LatestRelease struct {
	gorm.Model
	Name        string    `gorm:"not null"`
	ReleaseDate time.Time `gorm:"not null"`
	SpotifyID   string    `gorm:"not null;unique"`
	ArtistID    uint      `gorm:"not null;uniqueIndex"`
}

package services

import (
	"time"

	"github.com/ruined.yamb/v1/db"
	"github.com/ruined.yamb/v1/models"
)

type LatestReleaseService struct{}

func NewLatestReleaseService() *LatestReleaseService {
	return &LatestReleaseService{}
}

func (rs *LatestReleaseService) Create(name string, releaseDate time.Time, spotifyID string, artistID uint) error {
	release := models.LatestRelease{
		Name:        name,
		ReleaseDate: releaseDate,
		SpotifyID:   spotifyID,
		ArtistID:    artistID,
	}

	result := db.DB.Create(&release)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

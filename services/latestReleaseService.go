package services

import (
	"time"

	"github.com/ruined/yamb/v1/db"
	"github.com/ruined/yamb/v1/models"
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

func (rs *LatestReleaseService) IsMoreRecent(artistID uint, latestRelease models.LatestRelease) (bool, error) {
	var currentLatestRelease models.LatestRelease
	result := db.DB.Where("artist_id = ?", artistID).First(&currentLatestRelease)
	if result.Error != nil {
		return false, result.Error
	}
	return currentLatestRelease.ReleaseDate.Before(latestRelease.ReleaseDate), nil
}

func (rs *LatestReleaseService) Update(artistID uint, latestRelease models.LatestRelease) error {
	var currentLatestRelease models.LatestRelease
	result := db.DB.Where("artist_id = ?", artistID).First(&currentLatestRelease)
	if result.Error != nil {
		return result.Error
	}

	currentLatestRelease.Name = latestRelease.Name
	currentLatestRelease.ReleaseDate = latestRelease.ReleaseDate
	currentLatestRelease.SpotifyID = latestRelease.SpotifyID

	return db.DB.Save(&currentLatestRelease).Error
}

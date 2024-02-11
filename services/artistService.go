package services

import (
	"errors"

	"github.com/ruined.yamb/v1/db"
	"github.com/ruined.yamb/v1/models"
	"gorm.io/gorm"
)

type ArtistService struct{}

func NewArtistService() *ArtistService {
	return &ArtistService{}
}

func (as *ArtistService) Create(name, spotifyID string) (uint, error) {
	artist := models.Artist{Name: name, SpotifyID: spotifyID}
	result := db.DB.Create(&artist)
	if result.Error != nil {
		return 0, result.Error
	}
	return artist.ID, nil
}

func (as *ArtistService) FindByName(name string) (*models.Artist, error) {
	var artist models.Artist
	result := db.DB.Where("name = ?", name).First(&artist)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &artist, nil
}

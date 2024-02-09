package services

import (
	"errors"
	"log"

	"github.com/ruined.yamb/v1/db"
	"github.com/ruined.yamb/v1/models"
	"gorm.io/gorm"
)

type ArtistService struct{}

func NewArtistService() *ArtistService {
	return &ArtistService{}
}

func (as *ArtistService) Create(name, spotifyID string) error {
	artist := models.Artist{Name: name, SpotifyID: spotifyID}
	result := db.DB.Create(&artist)
	if result.Error != nil {
		log.Printf("error creating artist: %v", result.Error)
		return result.Error
	}
	return nil
}

func (as *ArtistService) FindByName(name string) (*models.Artist, error) {
	var artist models.Artist
	result := db.DB.Where("name = ?", name).First(&artist)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Printf("error finding artist by name: %v", result.Error)
		return nil, result.Error
	}
	return &artist, nil
}

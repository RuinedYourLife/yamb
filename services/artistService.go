package services

import (
	"errors"

	"github.com/ruined.yamb/v1/db"
	"github.com/ruined.yamb/v1/models"
	"github.com/ruined.yamb/v1/yamberrors"
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
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return 0, &yamberrors.AlreadyTrackedError{Entity: name}
		}
		return 0, result.Error
	}
	return artist.ID, nil
}

func (as *ArtistService) GetAll() ([]models.Artist, error) {
	var artists []models.Artist
	result := db.DB.Find(&artists)
	if result.Error != nil {
		return nil, result.Error
	}
	return artists, nil
}

func (as *ArtistService) FindByID(id uint) (*models.Artist, error) {
	var artist models.Artist
	result := db.DB.First(&artist, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &artist, nil
}

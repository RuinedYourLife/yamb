package tasks

import (
	"log"
	"time"

	"github.com/ruined.yamb/v1/models"
	"github.com/ruined.yamb/v1/services"
)

type ArtistCheckTask struct {
	ArtistID  uint
	SpotifyID string
}

var artistCheckQueue = make(chan ArtistCheckTask, 100)

func CheckForNewReleases() {
	artistService := services.NewArtistService()

	artists, err := artistService.GetAll()
	if err != nil {
		log.Printf("failed to get all artists: %v", err)
		return
	}

	for _, artist := range artists {
		artistCheckQueue <- ArtistCheckTask{
			ArtistID:  artist.ID,
			SpotifyID: artist.SpotifyID,
		}
	}
}

func ProcessArtistCheckQueue() {
	for task := range artistCheckQueue {
		processArtistCheckTask(task)
		time.Sleep(time.Second)
	}
}

func processArtistCheckTask(task ArtistCheckTask) {
	spotifyService := services.NewSpotifyService()
	releaseService := services.NewLatestReleaseService()

	latestRelease, err := spotifyService.FindArtistLatestRelease(task.SpotifyID)
	if err != nil {
		log.Printf("failed to find latest release for artist id %d: %v", task.ArtistID, err)
		return
	}

	toUpdate := models.LatestRelease{
		Name:        latestRelease.Name,
		ReleaseDate: latestRelease.ReleaseDate,
		SpotifyID:   latestRelease.SpotifyID,
		ArtistID:    task.ArtistID,
	}

	moreRecent, err := releaseService.IsMoreRecent(task.ArtistID, toUpdate)
	if err != nil {
		log.Printf("failed to compare release dates for artist id %d: %v", task.ArtistID, err)
		return
	}

	if moreRecent {
		err = releaseService.Update(task.ArtistID, toUpdate)
		if err != nil {
			log.Printf("failed to update latest release for artist id %d: %v", task.ArtistID, err)
		}
	}
}

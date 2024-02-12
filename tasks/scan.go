package tasks

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ruined.yamb/v1/models"
	"github.com/ruined.yamb/v1/services"
)

type ArtistCheckTask struct {
	ArtistID  uint
	SpotifyID string
}

var artistCheckQueue = make(chan ArtistCheckTask, 100)

func ScanForReleases() {
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

func ProcessArtistCheckQueue(s *discordgo.Session) {
	for task := range artistCheckQueue {
		processArtistCheckTask(s, task)
		time.Sleep(time.Second)
	}
}

func processArtistCheckTask(s *discordgo.Session, task ArtistCheckTask) {
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
			return
		}
		postNewRelease(s, toUpdate)
	}
}

func postNewRelease(s *discordgo.Session, release models.LatestRelease) {
	artistService := services.NewArtistService()
	artist, err := artistService.FindByID(release.ArtistID)
	if err != nil {
		log.Printf("failed to get artist for release id %d: %v", release.ID, err)
	}

	spotifyService := services.NewSpotifyService()
	albumURL, err := spotifyService.FindAlbum(release.SpotifyID)
	if err != nil {
		log.Printf("failed to get album for release id %d: %v", release.ID, err)
	}

	channelID := os.Getenv("YAMB_CHANNEL_ID")
	message := fmt.Sprintf("New release from %s: %s - %s", artist.Name, release.Name, albumURL["spotify"])

	_, err = s.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Printf("failed to send message for release id %d: %v", release.ID, err)
	}
}

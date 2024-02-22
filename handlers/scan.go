package handlers

import (
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ruined/yamb/v1/models"
	"github.com/ruined/yamb/v1/services"
)

type ArtistCheckTask struct {
	ArtistID  uint
	SpotifyID string
}

var artistCheckQueue = make(chan ArtistCheckTask, 100)

func ScanCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	go ScanForReleases()

	err := ReplyEmbed(s, i, &discordgo.MessageEmbed{
		Title:       "Scanning",
		Description: "Please wait...",
		Color:       0xD4AF91,
	})

	if err != nil {
		SendErrorEmbed(s, os.Getenv("YAMB_CHANNEL_ID"), "Cannot check for new releases")
		return
	}
}

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

	latestRelease, err := spotifyService.FetchArtistLatestRelease(task.SpotifyID)
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
		postNewRelease(s, task.ArtistID, latestRelease.Name, latestRelease.ReleaseDate, latestRelease.SpotifyID)
	}
}

func postNewRelease(s *discordgo.Session, artistID uint, releaseName string, releaseDate time.Time, releaseID string) {
	artistService := services.NewArtistService()
	artist, err := artistService.FindByID(artistID)
	if err != nil {
		log.Printf("failed to get artist for release id %s: %v", releaseID, err)
	}

	spotifyService := services.NewSpotifyService()
	details, err := spotifyService.FetchAlbumDetails(releaseID)
	if err != nil {
		log.Printf("failed to get album for release id %s: %v", releaseID, err)
	}

	channelID := os.Getenv("YAMB_CHANNEL_ID")

	embed := &discordgo.MessageEmbed{
		Title: releaseName,
		URL:   details.URL,
		Color: 0xD4AF91,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Artist",
				Value:  artist.Name,
				Inline: true,
			},
			{
				Name:   "Release Date",
				Value:  releaseDate.Format("02/01/2006"),
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: details.ArtistImageURL,
		},
		Image: &discordgo.MessageEmbedImage{
			URL: details.ImageURL,
		},
	}

	err = SendEmbed(s, channelID, embed)
	if err != nil {
		SendErrorEmbed(s, channelID, "Could not send new release")
	}
}

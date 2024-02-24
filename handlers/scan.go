package handlers

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ruined/yamb/v1/models"
	"github.com/ruined/yamb/v1/services"
)

type ArtistCheckTask struct {
	ArtistID    uint
	ArtistName  string
	SpotifyID   string
	Interaction *discordgo.InteractionCreate
}

var artistCheckQueue = make(chan ArtistCheckTask, 100)

var startTime time.Time
var processedArtists int
var artistCount int

func ScanCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	go ScanForReleases(i)

	startTime = time.Now()
	processedArtists = 0

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

func ScanForReleases(i *discordgo.InteractionCreate) {
	artistService := services.NewArtistService()
	artists, err := artistService.GetAll()
	if err != nil {
		log.Printf("failed to get all artists: %v", err)
		return
	}
	artistCount = len(artists)

	for _, artist := range artists {
		artistCheckQueue <- ArtistCheckTask{
			ArtistID:    artist.ID,
			ArtistName:  artist.Name,
			SpotifyID:   artist.SpotifyID,
			Interaction: i,
		}
	}
}

func ProcessArtistCheckQueue(s *discordgo.Session) {
	for task := range artistCheckQueue {
		processArtistCheckTask(s, task)
		processedArtists++
		if task.Interaction != nil {
			updateProgress(s, task, time.Since(startTime), processedArtists)
		}
		time.Sleep(time.Second)

		if processedArtists == artistCount && task.Interaction != nil {
			err := UpdateEmbed(s, task.Interaction, &discordgo.MessageEmbed{
				Title:       "Scanning",
				Description: "Complete",
				Color:       0xD4AF91,
			})
			if err != nil {
				SendErrorEmbed(s, os.Getenv("YAMB_CHANNEL_ID"), "Cannot complete scan for new releases")
				return
			}
		}
	}
}

func updateProgress(s *discordgo.Session, task ArtistCheckTask, elapsedTime time.Duration, processedArtists int) {
	hours, minutes, seconds := elapsedTime.Hours(), elapsedTime.Minutes(), elapsedTime.Seconds()
	formattedElapsedTime := fmt.Sprintf("%02d:%02d:%02d", int(hours), int(minutes)%60, int(seconds)%60)

	progressBarLength := 38
	progress := max(0, min(processedArtists*progressBarLength/artistCount, progressBarLength))
	progressBar := strings.Repeat("#", progress) + strings.Repeat("-", max(0, progressBarLength-progress))

	remaining := max(1, artistCount-processedArtists)
	eta := elapsedTime / time.Duration(processedArtists) * time.Duration(remaining)
	formattedEta := eta.Truncate(time.Second).String()

	description := fmt.Sprintf(
		"[%s] [%s] %d/%d         \n(ETA: %s) %s",
		formattedElapsedTime, progressBar, processedArtists, artistCount, formattedEta, task.ArtistName,
	)

	UpdateEmbed(s, task.Interaction, &discordgo.MessageEmbed{
		Title:       "Scanning",
		Description: "```" + description + "```",
		Color:       0xD4AF91,
	})
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

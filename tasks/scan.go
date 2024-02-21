package tasks

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
	albumDetails, err := spotifyService.FindAlbumDetails(release.SpotifyID)
	if err != nil {
		log.Printf("failed to get album for release id %d: %v", release.ID, err)
	}

	channelID := os.Getenv("YAMB_CHANNEL_ID")

	embed := &discordgo.MessageEmbed{
		Title: release.Name,
		URL:   albumDetails["spotifyURL"],
		Color: 0xD4AF91,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Artist",
				Value:  artist.Name,
				Inline: true,
			},
			{
				Name:   "Release Date",
				Value:  release.ReleaseDate.Format("02/01/2006"),
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: albumDetails["artistImageURL"],
		},
		Image: &discordgo.MessageEmbedImage{
			URL: albumDetails["imageURL"],
		},
	}

	_, err = s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Printf("failed to send embed for release id %d: %v", release.ID, err)
	}
}

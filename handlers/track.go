package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ruined/yamb/v1/services"
	"github.com/ruined/yamb/v1/yamberrors"
)

func TrackCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	spotifyURL := i.ApplicationCommandData().Options[0].StringValue()

	spotifyService := services.NewSpotifyService()
	_, spotifyID := spotifyService.ExtractSpotifyInfos(spotifyURL)
	if spotifyID == "" || len(spotifyID) != 22 {
		ReplyErrorEmbed(s, i, "Invalid Spotify URL provided")
		return
	}

	artistDetails, err := spotifyService.FetchArtistDetails(spotifyID)
	if err != nil {
		ReplyErrorEmbed(s, i, "Could not find artist details for this URL")
		return
	}

	latestRelease, err := spotifyService.FetchArtistLatestRelease(spotifyID)
	if err != nil {
		ReplyErrorEmbed(s, i, "Could not find latest release for this artist")
		return
	}

	artistService := services.NewArtistService()
	artistID, err := artistService.Create(artistDetails.Name, artistDetails.SpotifyID)
	if err != nil {
		content := "Could not register artist"
		switch e := err.(type) {
		case *yamberrors.AlreadyTrackedError:
			content = e.Error()
		}

		ReplyErrorEmbed(s, i, content)
		return
	}

	releaseService := services.NewLatestReleaseService()
	err = releaseService.Create(latestRelease.Name, latestRelease.ReleaseDate, latestRelease.SpotifyID, artistID)
	if err != nil {
		ReplyErrorEmbed(s, i, "Could not register latest release for this artist")
		return
	}

	err = ReplyEmbed(s, i, &discordgo.MessageEmbed{
		Title:       artistDetails.Name,
		URL:         fmt.Sprintf("https://open.spotify.com/artist/%s", artistDetails.SpotifyID),
		Description: "is now being tracked",
		Color:       0xD4AF91,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: artistDetails.Images[0].URL,
		},
	})
	if err != nil {
		ReplyErrorEmbed(s, i, "Could not send reply")
	}
}

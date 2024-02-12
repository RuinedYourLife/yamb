package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ruined.yamb/v1/services"
	"github.com/ruined.yamb/v1/yamberrors"
)

func TrackCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	spotifyURL := i.ApplicationCommandData().Options[0].StringValue()

	spotifyService := services.NewSpotifyService()
	spotifyID := spotifyService.ExtractSpotifyID(spotifyURL)
	if spotifyID == "" || len(spotifyID) != 22 {
		sendErrorReply(s, i, "Invalid Spotify URL provided")
		return
	}

	artistDetails, err := spotifyService.FindArtistDetails(spotifyID)
	if err != nil {
		sendErrorReply(s, i, "Could not find artist details for this URL")
		return
	}

	latestRelease, err := spotifyService.FindArtistLatestRelease(spotifyID)
	if err != nil {
		sendErrorReply(s, i, "Could not find latest release for this artist")
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

		sendErrorReply(s, i, content)
		return
	}

	releaseService := services.NewLatestReleaseService()
	err = releaseService.Create(latestRelease.Name, latestRelease.ReleaseDate, latestRelease.SpotifyID, artistID)
	if err != nil {
		sendErrorReply(s, i, "Could not register latest release for this artist")
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       artistDetails.Name,
					URL:         fmt.Sprintf("https://open.spotify.com/artist/%s", artistDetails.SpotifyID),
					Description: "is now being tracked",
					Color:       0xD4AF91,
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: artistDetails.Images[0].URL,
					},
				},
			},
		},
	})
}

func sendErrorReply(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Description: message,
					Color:       0xBD5773,
				},
			},
		},
	})
}

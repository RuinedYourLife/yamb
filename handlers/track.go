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

	artistDetails, err := spotifyService.FindArtistDetails(spotifyID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Could not find artist with this URL.",
			},
		})
		return
	}

	latestRelease, err := spotifyService.FindLatestRelease(spotifyID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Could not find latest release for this artist.",
			},
		})
		return
	}

	artistService := services.NewArtistService()
	artistID, err := artistService.Create(artistDetails.Name, artistDetails.SpotifyID)
	if err != nil {
		content := "Could not register artist."
		switch e := err.(type) {
		case *yamberrors.AlreadyTrackedError:
			content = e.Error()
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		})
		return
	}

	releaseService := services.NewLatestReleaseService()
	err = releaseService.Create(latestRelease.Name, latestRelease.ReleaseDate, latestRelease.SpotifyID, artistID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Could not register latest release.",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Tracking new releases for %s", artistDetails.Name),
		},
	})
}

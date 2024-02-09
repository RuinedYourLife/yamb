package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/ruined.yamb/v1/services"
)

func TrackCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	spotifyURL := i.ApplicationCommandData().Options[0].StringValue()

	spotifyService := services.NewSpotifyService()

	spotifyID := spotifyService.ExtractSpotifyID(spotifyURL)
	artist, err := spotifyService.FindArtistDetails(spotifyID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Could not find artist with this URL.",
			},
		})
		return
	}

	artistService := services.NewArtistService()
	artistService.Create(artist.Name, artist.SpotifyID)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Tracking new releases for %s", artist.Name),
		},
	})
}

package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/ruined.yamb/v1/tasks"
)

func ScanCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	go tasks.CheckForNewReleases()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Checking for new releases...",
		},
	})

	if err != nil {
		log.Printf("failed to respond to interaction: %v", err)
	}
}

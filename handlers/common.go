package handlers

import "github.com/bwmarrin/discordgo"

func SendErrorReply(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
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

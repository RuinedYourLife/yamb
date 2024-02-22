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

func UpdateEmbedDescription(s *discordgo.Session, i *discordgo.InteractionCreate, e *discordgo.MessageEmbed, description string) {
	e.Description = description
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{e},
	})
}

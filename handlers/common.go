package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func SendEmbed(s *discordgo.Session, channelID string, e *discordgo.MessageEmbed) error {
	_, err := s.ChannelMessageSendEmbed(channelID, e)
	if err != nil {
		log.Printf("failed to send embed: %v", err)
		return err
	}

	return nil
}

func SendErrorEmbed(s *discordgo.Session, channelID string, message string) error {
	return SendEmbed(s, channelID, &discordgo.MessageEmbed{
		Description: message,
		Color:       0xBD5773,
	})
}

func ReplyEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, e *discordgo.MessageEmbed) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{e},
		},
	})
	if err != nil {
		log.Printf("failed to reply with embed: %v", err)
		return err
	}

	return nil
}

func ReplyErrorEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	return ReplyEmbed(s, i, &discordgo.MessageEmbed{
		Description: message,
		Color:       0xBD5773,
	})
}

func UpdateEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, e *discordgo.MessageEmbed) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{e},
	})
	if err != nil {
		log.Printf("failed to update embed: %v", err)
		return err
	}

	return nil
}

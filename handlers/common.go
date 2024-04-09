package handlers

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/ruined/yamb/v1/services"
	"github.com/ruined/yamb/v1/util"
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

func UpdateEmbedContent(s *discordgo.Session, i *discordgo.InteractionCreate, e *discordgo.MessageEmbed) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{e},
	})
	if err != nil {
		log.Printf("failed to update embed: %v", err)
		return err
	}

	return nil
}

func UpdateEmbedContentWithButton(s *discordgo.Session, i *discordgo.InteractionCreate, e *discordgo.MessageEmbed, buttonLabel string, buttonUrl string) error {
	button := discordgo.Button{
		Label: buttonLabel,
		Style: discordgo.LinkButton,
		URL:   buttonUrl,
	}

	actionRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{button},
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{e},
		Components: &[]discordgo.MessageComponent{actionRow},
	})
	if err != nil {
		log.Printf("failed to update embed: %v", err)
		return err
	}

	return nil
}

func PostSpotifyResource(s *discordgo.Session, i *discordgo.InteractionCreate, e *discordgo.MessageEmbed, details *services.SpotifyResourceDetails, url string) error {
	var err error
	isFromDownload := false

	if details == nil {
		isFromDownload = true
		spotifyService := services.NewSpotifyService()
		resourceType, resourceID := spotifyService.ExtractSpotifyInfos(url)

		switch resourceType {
		case "album":
			details, err = spotifyService.FetchAlbumDetails(resourceID)
		case "track":
			details, err = spotifyService.FetchTrackDetails(resourceID)
		case "playlist":
			details, err = spotifyService.FetchPlaylistDetails(resourceID)
		}

		if err != nil {
			ReplyErrorEmbed(s, i, "Could not find details for this URL")
		}
	}

	updateEmbed(details, e)

	if isFromDownload {
		username := util.SanitizeLowerString(i.Member.User.Username)
		err = UpdateEmbedContentWithButton(s, i, e, "Download", fmt.Sprintf("http://yamb.ruined.fr/%s.zip", username))
		if err != nil {
			return err
		}
	} else {
		err = SendEmbed(s, os.Getenv("YAMB_CHANNEL_ID"), e)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateEmbed(details *services.SpotifyResourceDetails, e *discordgo.MessageEmbed) {
	fields := []*discordgo.MessageEmbedField{}

	if details.ArtistName != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Artist",
			Value:  details.ArtistName,
			Inline: true,
		})
	}
	if details.ReleaseDate != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Release Date",
			Value:  details.ReleaseDate,
			Inline: true,
		})
	}
	if details.OwnerName != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Owner",
			Value:  details.OwnerName,
			Inline: true,
		})
	}
	if details.Public != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Public",
			Value:  details.Public,
			Inline: true,
		})
	}
	if details.ArtistImageURL != "" {
		e.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: details.ArtistImageURL,
		}
	}
	if details.OwnerImageURL != "" {
		e.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: details.OwnerImageURL,
		}
	}

	e.Description = ""
	e.Title = details.Name
	e.URL = details.URL
	e.Color = 0xD4AF91
	e.Fields = fields
	e.Image = &discordgo.MessageEmbedImage{
		URL: details.ImageURL,
	}
}

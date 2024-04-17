package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "track",
			Description: "Track new releases for an artist",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "spotify-url",
					Description: "Spotify URL for the artist",
					Required:    true,
				},
			},
		},
		{
			Name:        "scan",
			Description: "Scan for new releases for all artists",
		},
		{
			Name:        "dl",
			Description: "Download a specific resource from Spotify",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "spotify-url",
					Description: "Spotify URL for the resource to download",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "format",
					Description: "Format to download the resource in (default: flac)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "flac", Value: "flac"},
						{Name: "mp3", Value: "mp3"},
						{Name: "wav", Value: "wav"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "ordered",
					Description: "Prefixes the filename with its index (default: false)",
					Required:    false,
				},
			},
		},
	}

	command_handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"track": TrackCommandHandler,
		"scan":  ScanCommandHandler,
		"dl":    DownloadCommandHandler,
	}
)

func RegisterHandlers(s *discordgo.Session, guildID string) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := command_handlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	for _, command := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
		if err != nil {
			log.Fatalf("failed to register command %s: %v", command.Name, err)
		}
	}
}

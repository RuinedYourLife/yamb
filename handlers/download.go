package handlers

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/creack/pty"
	"github.com/ruined/yamb/v1/util"
)

func DownloadCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "Download",
		Description: "Starting",
		Color:       0xD4AF91,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	url := i.ApplicationCommandData().Options[0].StringValue()
	ordered := i.ApplicationCommandData().Options[1].BoolValue()

	go func() {
		err := spotifyDL(s, i, embed, url, ordered)
		if err != nil {
			SendErrorReply(s, i, "Could not use spotify-dl")
			return
		}
	}()
}

func spotifyDL(s *discordgo.Session, i *discordgo.InteractionCreate, e *discordgo.MessageEmbed, url string, ordered bool) error {
	cmdName := "spotify-dl"
	args := []string{
		"-u", os.Getenv("SPOTIFY_USERNAME"),
		"-p", os.Getenv("SPOTIFY_PASSWORD"),
		url,
		"-d", os.Getenv("YAMB_DOWNLOAD_DIR"),
	}
	if ordered {
		args = append([]string{"-o"}, args...)
	}

	cmd := exec.Command(cmdName, args...)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("[+] spotify-dl failed to start: %v", err)
		return err
	}
	defer func() { _ = ptmx.Close() }()

	scanner := bufio.NewScanner(ptmx)

	for scanner.Scan() {
		line := util.StripANSICodes(scanner.Text())
		if strings.Contains(line, "Adding all songs from") {
			continue
		}

		UpdateEmbedDescription(s, i, e, "```"+line+"```")
		time.Sleep(500 * time.Millisecond)
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("[+] spotify-dl failed to execute: %v", err)
		return err
	}

	UpdateEmbedDescription(s, i, e, "Completed")

	return nil
}

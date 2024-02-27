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
		Title:       "Downloading",
		Description: "Starting",
		Color:       0xD4AF91,
	}

	err := ReplyEmbed(s, i, embed)
	if err != nil {
		ReplyErrorEmbed(s, i, "Could not start download")
		return
	}

	url := ""
	ordered := false

	for _, option := range i.ApplicationCommandData().Options {
		if option.Name == "spotify-url" {
			url = option.StringValue()
		} else if option.Name == "ordered" && option.Type == discordgo.ApplicationCommandOptionBoolean {
			ordered = option.BoolValue()
		}
	}

	go func() {
		err := spotifyDL(s, i, embed, url, ordered)
		if err != nil {
			SendErrorEmbed(s, os.Getenv("YAMB_CHANNEL_ID"), "Could not use spotify-dl")
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

		e.Description = "```" + line + "```"
		err := UpdateEmbedContent(s, i, e)
		if err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("[+] spotify-dl failed to execute: %v", err)
		return err
	}

	PostSpotifyResource(s, i, e, nil, url)

	return nil
}

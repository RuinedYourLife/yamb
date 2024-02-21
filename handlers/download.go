package handlers

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
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
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	var outputLines []string
	for scanner.Scan() {
		outputLines = append(outputLines, scanner.Text())
		if len(outputLines) > 3 {
			outputLines = outputLines[len(outputLines)-3:]
		}

		content := strings.Join(outputLines, "\n")

		e.Description = "```" + content + "```"
		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{e},
		})
		if err != nil {
			e.Description = "Something went wrong"
			e.Color = 0xBD5773
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{e},
			})
			return err
		}
		time.Sleep(1 * time.Second)
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("[+] spotify-dl fail: %v\noutput", err)
		return err
	}

	e.Description = "Completed"
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{e},
	})

	return nil
}

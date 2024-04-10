package handlers

import (
	"archive/zip"
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
	username := util.SanitizeLowerString(i.Member.User.Username)
	downloadDir := path.Join(os.Getenv("YAMB_DOWNLOAD_DIR"), username)

	err := os.MkdirAll(downloadDir, os.ModeDir)
	if err != nil {
		log.Printf("failed to create download dir: %v", err)
		return err
	}

	cmdName := "spotify-dl"
	args := []string{
		"-u", os.Getenv("SPOTIFY_USERNAME"),
		"-p", os.Getenv("SPOTIFY_PASSWORD"),
		url,
		"-d", downloadDir,
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

	archivePath := filepath.Join(os.Getenv("YAMB_DOWNLOAD_DIR"), username+".zip")
	if err := createArchive(downloadDir, archivePath); err != nil {
		log.Printf("failed to create archive: %v", err)
		return err
	}

	err = os.RemoveAll(downloadDir)
	if err != nil {
		log.Printf("failed to remove download dir: %v", err)
		return err
	}

	PostSpotifyResource(s, i, e, nil, url)

	return nil
}

func createArchive(sourceDir, archivePath string) error {
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		log.Printf("failed to create archive: %v", err)
		return err
	}
	defer archiveFile.Close()

	archive := zip.NewWriter(archiveFile)
	defer archive.Close()

	err = filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("failed to walk through files: %v", err)
			return err
		}

		if info.IsDir() || filePath == archivePath {
			return nil
		}

		outputPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".mp3"
		err = convertToMp3(filePath, outputPath)
		if err != nil {
			log.Printf("failed to convert to mp3: %v", err)
			return err
		}

		srcFile, err := os.Open(outputPath)
		if err != nil {
			log.Printf("failed to open source file: %v", err)
			return err
		}
		defer srcFile.Close()

		relPath, err := filepath.Rel(sourceDir, outputPath)
		if err != nil {
			log.Printf("failed to get relative path: %v", err)
			return err
		}

		zipFile, err := archive.Create(relPath)
		if err != nil {
			log.Printf("failed to create zip file: %v", err)
			return err
		}

		_, err = io.Copy(zipFile, srcFile)
		return err
	})

	return err
}

func convertToMp3(inputPath, outputPath string) error {
	cmdName := "ffmpeg"
	args := []string{
		"-i", inputPath,
		"-vn",
		"-ar", "44100",
		"-ac", "2",
		"-b:a", "320k",
		"-f", "mp3",
		"-map_metadata", "0",
		"-id3v2_version", "3",
		outputPath,
	}

	cmd := exec.Command(cmdName, args...)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
	"github.com/ruined/yamb/v1/app"
	"github.com/ruined/yamb/v1/handlers"
)

func main() {
	app.Init()

	session, err := discordgo.New("Bot " + app.Token)
	if err != nil {
		log.Fatalf("failed to create discord session: %v", err)
	}

	err = session.Open()
	if err != nil {
		log.Fatalf("failed to open discord session: %v", err)
	}

	handlers.RegisterHandlers(session, *app.GuildID)

	e := echo.New()
	e.Static("/", os.Getenv("YAMB_DOWNLOAD_DIR"))

	go func() {
		if err := e.Start(fmt.Sprintf("0.0.0.0:%s", os.Getenv("WEB_SERVER_PORT"))); err != nil {
			log.Fatalf("failed to start web server: %v", err)
		}
	}()

	app.Run(session, func() {
		session.Close()
	})
}

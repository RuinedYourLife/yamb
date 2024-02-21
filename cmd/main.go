package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/ruined/yamb/v1/app"
	"github.com/ruined/yamb/v1/handlers"
)

func main() {
	app.Init()

	session, err := discordgo.New("Bot " + app.Token)
	if err != nil {
		log.Fatalf("failed to create discord session: %v", err)
	}

	defer session.Close()
	err = session.Open()
	if err != nil {
		log.Fatalf("failed to open discord session: %v", err)
	}

	handlers.RegisterHandlers(session, *app.GuildID)

	app.Run(session)

	log.Println("[+] yamb ready (:")
}

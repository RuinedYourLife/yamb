package app

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ruined/yamb/v1/db"
	"github.com/ruined/yamb/v1/handlers"
	"github.com/ruined/yamb/v1/tasks"
)

var (
	GuildID = flag.String("guild", "", "guild id")
	Token   = os.Getenv("YAMB_TOKEN")
)

func init() { flag.Parse() }

func Init() {
	db.Init()
	err := os.MkdirAll(os.Getenv("YAMB_DOWNLOAD_DIR"), os.ModeDir)
	if err != nil {
		log.Fatalf("failed to create download dir: %v", err)
	}
}

func Run(session *discordgo.Session) {
	go handlers.ProcessArtistCheckQueue(session)
	tasks.SetupCronJob()

	log.Println("[+] hello (:")

	listenSignal()
}

func listenSignal() {
	stop := make(chan os.Signal, 2)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	shutdown()
}

func shutdown() {
	tasks.StopCronJob()

	time.Sleep(time.Second)
	log.Println("[+] bye (:")

	os.Exit(0)
}

package tasks

import (
	"log"

	"github.com/robfig/cron/v3"
	"github.com/ruined/yamb/v1/handlers"
)

var (
	c = cron.New(cron.WithSeconds())
)

func SetupCronJob() {
	_, err := c.AddFunc("0 0 9 * * *", func() {
		handlers.ScanForReleases(nil)
	})

	if err != nil {
		log.Fatalf("failed to add cron job: %v", err)
	}

	c.Start()
}

func StopCronJob() {
	c.Stop()
}

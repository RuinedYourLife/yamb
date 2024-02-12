package tasks

import (
	"log"

	"github.com/robfig/cron/v3"
)

var (
	c = cron.New(cron.WithSeconds())
)

func SetupCronJob() {
	_, err := c.AddFunc("0 0 9 * * *", ScanForReleases)

	if err != nil {
		log.Fatalf("failed to add cron job: %v", err)
	}

	c.Start()
}

func StopCronJob() {
	c.Stop()
}

package tasks

import (
	"log"

	"github.com/robfig/cron/v3"
)

func SetupCronJobs() *cron.Cron {
	c := cron.New()

	_, err := c.AddFunc("0 0 9 * * *", CheckForNewReleases)

	if err != nil {
		log.Fatalf("failed to add cron job: %v", err)
	}

	c.Start()
	return c
}

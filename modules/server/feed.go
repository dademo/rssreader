package server

import (
	"fmt"

	"github.com/dademo/rssreader/modules/config"
	"github.com/dademo/rssreader/modules/feed"
	"github.com/dademo/rssreader/modules/scheduler"

	log "github.com/sirupsen/logrus"
)

type scheduledFeedReaderJob struct {
	Feed config.Feed
}

func ScheduleFromConfig(jobScheduler *scheduler.Scheduler, config config.Config) {

	for _, feed := range config.Feeds {
		jobScheduler.Schedule(scheduler.ScheduledJob{
			Job: scheduledFeedReaderJob{
				Feed: feed,
			},
		})
	}
}

func (scheduledFeedReaderJob scheduledFeedReaderJob) Run() {

	fetchedFeed, err := feed.Fetch(scheduledFeedReaderJob.Feed)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to fetch feed [%s]", scheduledFeedReaderJob.Feed.Name), err)
		return
	}

	err = fetchedFeed.Save()
	if err != nil {
		log.Error("Unable to persist values", err)
	}
}

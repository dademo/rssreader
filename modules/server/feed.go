package server

import (
	"github.com/dademo/rssreader/modules/config"
	"github.com/dademo/rssreader/modules/feed"
	"github.com/dademo/rssreader/modules/scheduler"
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
	feed.Fetch(scheduledFeedReaderJob.Feed) // TODO -> Save into the database
}

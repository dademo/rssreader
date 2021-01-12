package scheduler

import (
	"github.com/dademo/rssreader/modules/config"
	"github.com/dademo/rssreader/modules/feed"
)

type scheduledFeedReaderJob struct {
	Feed config.Feed
}

func ScheduleFromConfig(scheduler *Scheduler, config config.Config) {

	for _, feed := range config.Feeds {
		scheduler.Schedule(ScheduledJob{
			Job: scheduledFeedReaderJob{
				Feed: feed,
			},
		})
	}
}

func (scheduledFeedReaderJob scheduledFeedReaderJob) Run() {
	feed.Fetch(scheduledFeedReaderJob.Feed) // TODO -> Save into the database
}

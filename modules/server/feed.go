package server

import (
	"fmt"

	"github.com/dademo/rssreader/modules/config"
	"github.com/dademo/rssreader/modules/feed"
	appLog "github.com/dademo/rssreader/modules/log"
	"github.com/dademo/rssreader/modules/scheduler"
)

type scheduledFeedReaderJob struct {
	Feed *config.Feed
}

func ScheduleFromConfig(jobScheduler *scheduler.Scheduler, config *config.Config) {

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
		appLog.DebugError(err, fmt.Sprintf("Unable to fetch feed [%s]", scheduledFeedReaderJob.Feed.Name))
		return
	}

	err = fetchedFeed.Save()
	if err != nil {
		appLog.DebugError(err, "Unable to persist values")
	}
}

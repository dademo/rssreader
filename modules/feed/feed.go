package feed

import (
	"fmt"

	"github.com/dademo/rssreader/modules/config"
	"github.com/dademo/rssreader/modules/database"
	"github.com/mmcdole/gofeed"

	log "github.com/sirupsen/logrus"
)

func FetchAll(config config.Config) ([]*database.Feed, error) {

	log.Debug("Fetching all feeds")

	feeds := make([]*database.Feed, len(config.Feeds))

	for _, feed := range config.Feeds {
		fetchedFeed, err := Fetch(feed)

		if err != nil {
			return nil, err
		}

		feeds = append(feeds, fetchedFeed)
	}

	return feeds, nil
}

func Fetch(feedConfig config.Feed) (*database.Feed, error) {

	log.Debug(fmt.Sprintf("Fetching feed [%s]", feedConfig.Name))

	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(feedConfig.Url)

	if err != nil {
		return nil, err
	}

	return database.FromFeed(feed), nil
}

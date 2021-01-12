package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

const feedSql = `
CREATE TABLE feed_categories (
	id			INTEGER PRIMARY KEY NOT NULL,
	category	TEXT
);

CREATE TABLE feed_author (
	id			INTEGER PRIMARY KEY NOT NULL,
	name		TEXT,
	email		TEXT
);

CREATE TABLE feed_image (
	id			INTEGER PRIMARY KEY NOT NULL,
	url			TEXT,
	title		TEXT
);

CREATE TABLE feed_enclosure (
	id		INTEGER PRIMARY KEY NOT NULL,
	url		TEXT,
	length	TEXT,
	type	TEXT
);

CREATE TABLE feed_item (
	id				INTEGER PRIMARY KEY NOT NULL,
	id_author		INTEGER REFERENCES feed_author(id),
	id_image		INTEGER REFERENCES feed_image(id),
	title			TEXT,
	description		TEXT,
	content			TEXT,
	link			TEXT,
	updated			{{.SqlTimestamp}},
	published		{{.SqlTimestamp}},
	guid			TEXT
);

CREATE TABLE feed (
	id			INTEGER PRIMARY KEY NOT NULL,
	id_author	INTEGER REFERENCES feed_author(id),
	id_image	INTEGER REFERENCES feed_image(id),
	title		TEXT NOT NULL UNIQUE,
	description	TEXT,
	link		TEXT,
	feedlink	TEXT,
	updated	 	{{.SqlTimestamp}},
	published	{{.SqlTimestamp}},
	language	TEXT,
	copyright	TEXT,
	generator	TEXT,
	last_update	{{.SqlTimestamp}}
);

CREATE TABLE feed_categories_feed (
	id_feed_category	INTEGER NOT NULL REFERENCES feed_categories(id),
	id_feed				INTEGER NOT NULL REFERENCES feed(id)
);

CREATE TABLE feed_categories_item (
	id_feed_category	INTEGER NOT NULL REFERENCES feed_categories(id),
	id_feed_item		INTEGER NOT NULL REFERENCES feed_item(id)
);

CREATE TABLE feed_enclosure_item (
	id_feed_enclosure	INTEGER NOT NULL REFERENCES feed_enclosure(id),
	id_feed_item		INTEGER NOT NULL REFERENCES feed_item(id)
);
`

type Feed struct {
	Id          uint64
	Author      *FeedAuthor
	Image       *FeedImage
	Categories  []*FeedCategory
	Title       string
	Description string
	Link        string
	FeedLink    string
	Updated     *time.Time
	Published   *time.Time
	Language    string
	Copyright   string
	Generator   string
	LastUpdate  *time.Time
	dirty       bool
}

type FeedItem struct {
	Id          uint64
	Author      *FeedAuthor
	Image       *FeedImage
	Categories  []*FeedCategory
	Enclosures  []*FeedEnclosure
	Title       string
	Description string
	Content     string
	Link        string
	Updated     *time.Time
	Published   *time.Time
	GUID        string
	dirty       bool
}

type FeedCategory struct {
	Id       uint64
	Category string
	dirty    bool
}

type FeedAuthor struct {
	Id    uint64
	Name  string
	Email string
	dirty bool
}

type FeedImage struct {
	Id    uint64
	URL   string
	Title string
	dirty bool
}

type FeedEnclosure struct {
	Id     uint64
	URL    string
	Length string
	Type   string
	dirty  bool
}

func init() {
	RegisterDatabaseTableCreator(databaseFeedCreator)
}

func databaseFeedCreator(connection *sql.Conn) error {

	log.Debug("Creating feed tables")
	ctx := context.Background()

	sql, err := normalizedSql(feedSql)

	if err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("Running command :\n%s", sql))

	_, err = connection.ExecContext(ctx, sql)

	if err != nil {
		log.Error("Unable to create feed tables")
		return err
	}

	log.Debug("Feed tables created")
	return nil
}

func FromFeed(feed *gofeed.Feed) *Feed {
	return &Feed{
		Id:          0,
		Author:      FromPerson(feed.Author),
		Image:       FromImage(feed.Image),
		Categories:  mapCategories(feed.Categories),
		Title:       feed.Title,
		Description: feed.Description,
		Link:        feed.Link,
		FeedLink:    feed.FeedLink,
		Updated:     feed.UpdatedParsed,
		Published:   feed.PublishedParsed,
		Language:    feed.Language,
		Copyright:   feed.Copyright,
		Generator:   feed.Generator,
		LastUpdate:  nil,
		dirty:       false,
	}
}

func FromFeedItem(item *gofeed.Item) *FeedItem {
	return &FeedItem{
		Id:          0,
		Author:      FromPerson(item.Author),
		Image:       FromImage(item.Image),
		Categories:  mapCategories(item.Categories),
		Enclosures:  mapEnclosures(item.Enclosures),
		Title:       item.Title,
		Description: item.Description,
		Content:     item.Content,
		Link:        item.Link,
		Updated:     item.UpdatedParsed,
		Published:   item.PublishedParsed,
		GUID:        item.GUID,
		dirty:       false,
	}
}

func FromCategory(category string) *FeedCategory {
	return &FeedCategory{
		Id:       0,
		Category: category,
		dirty:    false,
	}
}

func FromPerson(author *gofeed.Person) *FeedAuthor {
	return &FeedAuthor{
		Id:    0,
		Name:  author.Name,
		Email: author.Email,
		dirty: false,
	}
}

func FromImage(image *gofeed.Image) *FeedImage {
	return &FeedImage{
		Id:    0,
		URL:   image.URL,
		Title: image.Title,
		dirty: false,
	}
}

func FromEnclosure(enclosure *gofeed.Enclosure) *FeedEnclosure {
	return &FeedEnclosure{
		Id:     0,
		URL:    enclosure.URL,
		Length: enclosure.Length,
		Type:   enclosure.Type,
		dirty:  false,
	}
}

func mapCategories(categories []string) []*FeedCategory {

	feedCategories := make([]*FeedCategory, len(categories))

	for _, category := range categories {
		feedCategories = append(feedCategories, FromCategory(category))
	}

	return feedCategories
}

func mapEnclosures(enclosures []*gofeed.Enclosure) []*FeedEnclosure {

	feedEnclosures := make([]*FeedEnclosure, len(enclosures))

	for _, enclosure := range enclosures {
		feedEnclosures = append(feedEnclosures, FromEnclosure(enclosure))
	}

	return feedEnclosures
}

func (f Feed) Save() error {
	return nil
}

func (f FeedItem) Save() error {
	return nil
}

func (f FeedCategory) Save() error {
	return nil
}

func (f FeedAuthor) Save() error {
	return nil
}

func (f FeedImage) Save() error {
	return nil
}

func (f FeedEnclosure) Save() error {
	return nil
}

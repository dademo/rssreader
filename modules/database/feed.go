package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

const feedSql = `
CREATE TABLE feed_category (
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
	id_feed				INTEGER NOT NULL REFERENCES feed(id),
	UNIQUE(id_feed_category, id_feed)
);

CREATE TABLE feed_categories_item (
	id_feed_category	INTEGER NOT NULL REFERENCES feed_categories(id),
	id_feed_item		INTEGER NOT NULL REFERENCES feed_item(id),
	UNIQUE(id_feed_category, id_feed_item)
);

CREATE TABLE feed_enclosure_item (
	id_feed_enclosure	INTEGER NOT NULL REFERENCES feed_enclosure(id),
	id_feed_item		INTEGER NOT NULL REFERENCES feed_item(id),
	UNIQUE(id_feed_enclosure, id_feed_item)
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
}

type FeedCategory struct {
	Id       uint64
	Category string
}

type FeedAuthor struct {
	Id    uint64
	Name  string
	Email string
}

type FeedImage struct {
	Id    uint64
	URL   string
	Title string
}

type FeedEnclosure struct {
	Id     uint64
	URL    string
	Length string
	Type   string
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
	}
}

func FromCategory(category string) *FeedCategory {
	return &FeedCategory{
		Id:       0,
		Category: category,
	}
}

func FromPerson(author *gofeed.Person) *FeedAuthor {
	return &FeedAuthor{
		Id:    0,
		Name:  author.Name,
		Email: author.Email,
	}
}

func FromImage(image *gofeed.Image) *FeedImage {
	return &FeedImage{
		Id:    0,
		URL:   image.URL,
		Title: image.Title,
	}
}

func FromEnclosure(enclosure *gofeed.Enclosure) *FeedEnclosure {
	return &FeedEnclosure{
		Id:     0,
		URL:    enclosure.URL,
		Length: enclosure.Length,
		Type:   enclosure.Type,
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

func (f *Feed) Save() error {

	if f.Author != nil {
		err := f.Author.Save()
		if err != nil {
			log.Debug("Unable to save a feed author")
			return err
		}
	}

	if f.Image != nil {
		err := f.Image.Save()
		if err != nil {
			log.Debug("Unable to save a feed image")
			return err
		}
	}

	if len(f.Categories) > 0 {
		for _, category := range f.Categories {
			err := category.Save()
			if err != nil {
				log.Debug("Unable to save a feed cateogry")
				return err
			}
		}
	}

	if f.Id == 0 {

		log.Debug("Adding a new feed")

		sql := `
			INSERT INTO feed (id_author, id_image, title, description, link, feedLink, updated, published, language, copyright, generator, last_update)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			log.Debug("Unable to create the statement for feed creation")
			return err
		}

		newId, err := sqlExec(stmt,
			nilIfZero(f.Author.Id),
			nilIfZero(f.Image.Id),
			f.Title,
			f.Description,
			f.Link,
			f.FeedLink,
			f.Updated,
			f.Published,
			f.Language,
			f.Copyright,
			f.Generator,
			time.Now(),
		)

		if err != nil {
			log.Debug("An error occured while saving a feed")
			return err
		} else {
			f.Id = PrimaryKey(newId)
		}

	} else {

		sql := `
			UPDATE feed SET
				id_author = ?,
				id_image = ?,
				title = ?,
				description = ?,
				link = ?,
				feedLink = ?,
				updated = ?,
				published = ?,
				language = ?,
				copyright = ?,
				generator = ?,
				last_update = ?
			WHERE id = ?
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			log.Debug("Unable to create the statement for feed creation")
			return err
		}

		newId, err := sqlExec(stmt,
			nilIfZero(f.Author.Id),
			nilIfZero(f.Image.Id),
			f.Title,
			f.Description,
			f.Link,
			f.FeedLink,
			f.Updated,
			f.Published,
			f.Language,
			f.Copyright,
			f.Generator,
			time.Now(),
			f.Id,
		)

		if err != nil {
			log.Debug(fmt.Sprintf("An error occured while updating a feed (%d)", f.Id))
			return err
		} else {
			f.Id = PrimaryKey(newId)
		}
	}

	log.Debug("Linking feeds to its categories")
	if len(f.Categories) > 0 {
		for _, category := range f.Categories {
			err := linkFeedCategoryToFeed(category, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *FeedItem) Save() error {

	if f.Author != nil {
		err := f.Author.Save()
		if err != nil {
			log.Debug("Unable to save a feed author")
			return err
		}
	}

	if f.Image != nil {
		err := f.Image.Save()
		if err != nil {
			log.Debug("Unable to save a feed image")
			return err
		}
	}

	if len(f.Categories) > 0 {
		for _, category := range f.Categories {
			err := category.Save()
			if err != nil {
				log.Debug("Unable to save a feed cateogry")
				return err
			}
		}
	}

	if len(f.Enclosures) > 0 {
		for _, enclosure := range f.Enclosures {
			err := enclosure.Save()
			if err != nil {
				log.Debug("Unable to save a feed enclosure")
				return err
			}
		}
	}

	if f.Id == 0 {

		log.Debug("Adding a new feed item")

		sql := `
			INSERT INTO feed_item (id_author, id_image, title, description, content, link, updated, published, guid)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			log.Debug("Unable to create the statement for feed item creation")
			return err
		}

		newId, err := sqlExec(stmt,
			nilIfZero(f.Author.Id),
			nilIfZero(f.Image.Id),
			f.Title,
			f.Description,
			f.Content,
			f.Link,
			f.Updated,
			f.Published,
			f.GUID,
			time.Now(),
		)

		if err != nil {
			log.Debug("An error occured while saving a feed item")
			return err
		} else {
			f.Id = PrimaryKey(newId)
		}

	} else {

		sql := `
			UPDATE feed_item SET
				id_author = ?,
				id_image = ?,
				title = ?,
				description = ?,
				content = ?,
				link = ?,
				updated = ?,
				published = ?,
				guid = ?
			WHERE id = ?
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			log.Debug("Unable to create the statement for feed item creation")
			return err
		}

		newId, err := sqlExec(stmt,
			nilIfZero(f.Author.Id),
			nilIfZero(f.Image.Id),
			f.Title,
			f.Description,
			f.Content,
			f.Link,
			f.Updated,
			f.Published,
			f.GUID,
			f.Id,
		)

		if err != nil {
			log.Debug(fmt.Sprintf("An error occured while updating a feed item (%d)", f.Id))
			return err
		} else {
			f.Id = PrimaryKey(newId)
		}
	}

	log.Debug("Linking feed item to its categories")
	if len(f.Enclosures) > 0 {
		for _, category := range f.Categories {
			err := linkFeedCategoryToFeedItem(category, f)
			if err != nil {
				return err
			}
		}
	}

	log.Debug("Linking feed item to its enclosure")
	if len(f.Enclosures) > 0 {
		for _, enclosure := range f.Enclosures {
			err := linkFeedEnclosureToFeedItem(enclosure, f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *FeedCategory) Save() error {

	if f.Id == 0 {

		existing, err := feedCategoryByCategory(f.Category)
		if err != nil {
			log.Debug("Unable to get category by name")
			return err
		}

		if existing != nil {
			f.Id = existing.Id
			return nil
		} else {

			log.Debug("Adding a new feed category")

			sql := `
			INSERT INTO feed_category (category)
			VALUES (?)
		`

			stmt, err := database.Prepare(sql)

			if err != nil {
				log.Debug("Unable to create the statement for feed category update")
				return err
			}

			newId, err := sqlExec(stmt,
				f.Category,
			)

			if err != nil {
				log.Debug("An error occured while saving a feed category")
				return err
			} else {
				f.Id = PrimaryKey(newId)
				return nil
			}
		}

	} else {

		sql := `
			UPDATE feed_category SET
				category = ?
			WHERE id = ?
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			log.Debug("Unable to create the statement for feed category update")
			return err
		}

		newId, err := sqlExec(stmt,
			f.Category,
			f.Id,
		)

		if err != nil {
			log.Debug(fmt.Sprintf("An error occured while updating a feed category (%d)", f.Id))
			return err
		} else {
			f.Id = PrimaryKey(newId)
			return nil
		}
	}
}

func (f *FeedAuthor) Save() error {

	if f.Id == 0 {

		existing, err := feedAuthorByName(f.Name)
		if err != nil {
			log.Debug("Unable to get author by name")
			return err
		}

		if existing != nil {
			f.Id = existing.Id
			return nil
		} else {

			log.Debug("Adding a new feed author")

			sql := `
			INSERT INTO feed_author (name, email)
			VALUES (?, ?)
		`

			stmt, err := database.Prepare(sql)

			if err != nil {
				log.Debug("Unable to create the statement for feed author update")
				return err
			}

			newId, err := sqlExec(stmt,
				f.Name,
				f.Email,
			)

			if err != nil {
				log.Debug("An error occured while saving a feed author")
				return err
			} else {
				f.Id = PrimaryKey(newId)
				return nil
			}
		}

	} else {

		sql := `
			UPDATE feed_author SET
				name = ?,
				email = ?
			WHERE id = ?
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			log.Debug("Unable to create the statement for feed author update")
			return err
		}

		newId, err := sqlExec(stmt,
			f.Name,
			f.Email,
			f.Id,
		)

		if err != nil {
			log.Debug(fmt.Sprintf("An error occured while updating a feed author (%d)", f.Id))
			return err
		} else {
			f.Id = PrimaryKey(newId)
			return nil
		}
	}
}

func (f *FeedImage) Save() error {

	if f.Id == 0 {

		log.Debug("Adding a new feed image")

		sql := `
			INSERT INTO feed_image (url, title)
			VALUES (?, ?)
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			log.Debug("Unable to create the statement for feed image update")
			return err
		}

		newId, err := sqlExec(stmt,
			f.URL,
			f.Title,
		)

		if err != nil {
			log.Debug("An error occured while saving a feed image")
			return err
		} else {
			f.Id = PrimaryKey(newId)
			return nil
		}

	} else {

		sql := `
			UPDATE feed_image SET
				url = ?,
				title = ?
			WHERE id = ?
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			log.Debug("Unable to create the statement for feed image update")
			return err
		}

		newId, err := sqlExec(stmt,
			f.URL,
			f.Title,
			f.Id,
		)

		if err != nil {
			log.Debug(fmt.Sprintf("An error occured while updating a feed image (%d)", f.Id))
			return err
		} else {
			f.Id = PrimaryKey(newId)
			return nil
		}
	}
}

func (f *FeedEnclosure) Save() error {

	if f.Id == 0 {

		log.Debug("Adding a new feed enclosure")

		sql := `
			INSERT INTO feed_enclosure (url, length, type)
			VALUES (?, ?, ?)
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			log.Debug("Unable to create the statement for feed enclosure update")
			return err
		}

		newId, err := sqlExec(stmt,
			f.URL,
			f.Length,
			f.Type,
		)

		if err != nil {
			log.Debug("An error occured while saving a feed enclosure")
			return err
		} else {
			f.Id = PrimaryKey(newId)
			return nil
		}

	} else {

		sql := `
			UPDATE feed_enclosure SET
				url = ?,
				length = ?,
				type = ?
			WHERE id = ?
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			log.Debug("Unable to create the statement for feed enclosure update")
			return err
		}

		newId, err := sqlExec(stmt,
			f.URL,
			f.Length,
			f.Type,
			f.Id,
		)

		if err != nil {
			log.Debug(fmt.Sprintf("An error occured while updating a feed enclosure (%d)", f.Id))
			return err
		} else {
			f.Id = PrimaryKey(newId)
			return nil
		}
	}
}

func feedCategoryByCategory(category string) (*FeedCategory, error) {

	sql := `
		SELECT
			id,
			category
		FROM feed_category
		WHERE category = ?
	`

	stmt, err := database.Prepare(sql)
	if err != nil {
		log.Debug("Unable to prepare statement")
		return nil, err
	}

	v := new(FeedCategory)

	row := stmt.QueryRow(category)
	if row.Err() != nil {
		log.Debug("Unable to get result row")
		return nil, err
	}

	err = row.Scan(v)
	if err != nil {
		log.Debug("Unable to affect results")
		return nil, err
	} else {
		return v, nil
	}
}

func feedAuthorByName(name string) (*FeedAuthor, error) {

	sql := `
		SELECT
			id,
			name,
			email
		FROM feed_author
		WHERE name = ?
	`

	stmt, err := database.Prepare(sql)
	if err != nil {
		log.Debug("Unable to prepare statement")
		return nil, err
	}

	v := new(FeedAuthor)

	row := stmt.QueryRow(name)
	if row.Err() != nil {
		log.Debug("Unable to get result row")
		return nil, err
	}

	err = row.Scan(v)
	if err != nil {
		log.Debug("Unable to affect results")
		return nil, err
	} else {
		return v, nil
	}
}

func linkFeedCategoryToFeed(category *FeedCategory, feed *Feed) error {

	type result struct {
		Cnt int `db:"cnt"`
	}

	if category.Id == 0 {
		return errors.New("You must provide a saved feed category")
	}

	if feed.Id == 0 {
		return errors.New("You must provide a saved feed")
	}

	sql := `
		SELECT
			COUNT(*) AS cnt
		FROM feed_categories_feed
		WHERE 
			    id_feed_category = ?
			AND id_feed = ?
	`

	stmt, err := database.Prepare(sql)
	if err != nil {
		log.Debug("Unable to prepare SELECT statement")
		return err
	}

	v := new(result)

	row := stmt.QueryRow(category.Id, feed.Id)
	if row.Err() != nil {
		log.Debug("Unable to get result row")
		return err
	}

	err = row.Scan(v)
	if err != nil {
		log.Debug("Unable to affect results")
		return err
	}

	if v.Cnt > 0 {
		return nil
	} else {
		// Adding value
		sql = `
			INSERT INTO feed_categories_feed(id_feed_category, id_feed)
			VALUES (?, ?)
		`
		stmt, err = database.Prepare(sql)
		if err != nil {
			log.Debug("Unable to prepare INSERT statement")
			return err
		}

		countInserted, err := sqlExec(stmt, category.Id, feed.Id)
		if err != nil {
			log.Debug("An error occured while saving a feed")
			return err
		} else {
			log.Debug(fmt.Sprintf("Inserted %d row", countInserted))
			return nil
		}
	}
}

func linkFeedCategoryToFeedItem(category *FeedCategory, item *FeedItem) error {

	type result struct {
		Cnt int `db:"cnt"`
	}

	if category.Id == 0 {
		return errors.New("You must provide a saved feed category")
	}

	if item.Id == 0 {
		return errors.New("You must provide a saved feed item")
	}

	sql := `
		SELECT
			COUNT(*) AS cnt
		FROM feed_categories_item
		WHERE 
			    id_feed_category = ?
			AND id_feed_item = ?
	`

	stmt, err := database.Prepare(sql)
	if err != nil {
		log.Debug("Unable to prepare SELECT statement")
		return err
	}

	v := new(result)

	row := stmt.QueryRow(category.Id, item.Id)
	if row.Err() != nil {
		log.Debug("Unable to get result row")
		return err
	}

	err = row.Scan(v)
	if err != nil {
		log.Debug("Unable to affect results")
		return err
	}

	if v.Cnt > 0 {
		return nil
	} else {
		// Adding value
		sql = `
			INSERT INTO feed_categories_item(id_feed_category, id_feed_item)
			VALUES (?, ?)
		`
		stmt, err = database.Prepare(sql)
		if err != nil {
			log.Debug("Unable to prepare INSERT statement")
			return err
		}

		countInserted, err := sqlExec(stmt, category.Id, item.Id)
		if err != nil {
			log.Debug("An error occured while saving a feed")
			return err
		} else {
			log.Debug(fmt.Sprintf("Inserted %d row", countInserted))
			return nil
		}
	}
}

func linkFeedEnclosureToFeedItem(enclosure *FeedEnclosure, item *FeedItem) error {

	type result struct {
		Cnt int `db:"cnt"`
	}

	if enclosure.Id == 0 {
		return errors.New("You must provide a saved feed enclosure")
	}

	if item.Id == 0 {
		return errors.New("You must provide a saved feed item")
	}

	sql := `
		SELECT
			COUNT(*) AS cnt
		FROM feed_enclosure_item
		WHERE 
			    id_feed_enclosure = ?
			AND id_feed_item = ?
	`

	stmt, err := database.Prepare(sql)
	if err != nil {
		log.Debug("Unable to prepare SELECT statement")
		return err
	}

	v := new(result)

	row := stmt.QueryRow(enclosure.Id, item.Id)
	if row.Err() != nil {
		log.Debug("Unable to get result row")
		return err
	}

	err = row.Scan(v)
	if err != nil {
		log.Debug("Unable to affect results")
		return err
	}

	if v.Cnt > 0 {
		return nil
	} else {
		// Adding value
		sql = `
			INSERT INTO feed_enclosure_item(id_feed_enclosure, id_feed_item)
			VALUES (?, ?)
		`
		stmt, err = database.Prepare(sql)
		if err != nil {
			log.Debug("Unable to prepare INSERT statement")
			return err
		}

		countInserted, err := sqlExec(stmt, enclosure.Id, item.Id)
		if err != nil {
			log.Debug("An error occured while saving a feed")
			return err
		} else {
			log.Debug(fmt.Sprintf("Inserted %d row", countInserted))
			return nil
		}
	}
}

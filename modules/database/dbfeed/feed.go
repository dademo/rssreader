package dbfeed

import (
	"fmt"
	"time"

	appDatabase "github.com/dademo/rssreader/modules/database"
	appLog "github.com/dademo/rssreader/modules/log"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

type Feed struct {
	Id          uint64
	Author      *FeedAuthor
	Image       *FeedImage
	Categories  []*FeedCategory
	Items       []*FeedItem
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

func FromFeed(feed *gofeed.Feed) *Feed {
	return &Feed{
		Id:          0,
		Author:      FromPerson(feed.Author),
		Image:       FromImage(feed.Image),
		Categories:  mapCategories(feed.Categories),
		Items:       mapItems(feed.Items),
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

func (f *Feed) Save() error {

	log.Debug("Saving a feed")

	if f.Author != nil {
		err := f.Author.Save()
		if err != nil {
			appLog.DebugError("Unable to save a feed author")
			return err
		}
	}

	if f.Image != nil {
		err := f.Image.Save()
		if err != nil {
			appLog.DebugError("Unable to save a feed image")
			return err
		}
	}

	if len(f.Categories) > 0 {
		for _, category := range f.Categories {
			err := category.Save()
			if err != nil {
				appLog.DebugError("Unable to save a feed cateogry")
				return err
			}
		}
	}

	existingFeed, err := feedByTitle(f.Title)
	if err != nil {
		appLog.DebugError("Unable to check for feed existance")
		return err
	}

	if existingFeed != nil {
		f.Id = existingFeed.Id
	}

	if f.Id == 0 {

		log.Debug("Adding a new feed")

		sql, err := appDatabase.NormalizedSql(`
			INSERT INTO feed (id_author, id_image, title, description, link, feed_link, updated, published, language, copyright, generator, last_update)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			appLog.DebugError(err)
			return err
		}

		stmt, err := database.Prepare(appDatabase.PrepareExecSQL(sql))

		if err != nil {
			appLog.DebugError("Unable to create the statement for feed creation")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		newId, err := appDatabase.SqlExecGetId(stmt,
			appDatabase.EntityId(f.Author),
			appDatabase.EntityId(f.Image),
			appDatabase.StrWithMaxLength(f.Title, 200),
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
			appLog.DebugError("An error occured while saving a feed")
			return err
		} else {
			f.Id = appDatabase.PrimaryKey(newId)
		}

	} else {

		log.Debug("Updating a feed")

		sql, err := appDatabase.NormalizedSql(`
			UPDATE feed SET
				id_author = ?,
				id_image = ?,
				title = ?,
				description = ?,
				link = ?,
				feed_link = ?,
				updated = ?,
				published = ?,
				language = ?,
				copyright = ?,
				generator = ?,
				last_update = ?
			WHERE id = ?
		`)
		if err != nil {
			appLog.DebugError(err)
			return err
		}

		stmt, err := database.Prepare(appDatabase.PrepareExecSQL(sql))

		if err != nil {
			appLog.DebugError("Unable to create the statement for feed creation")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		_, err = appDatabase.SqlExecGetId(stmt,
			appDatabase.EntityId(f.Author),
			appDatabase.EntityId(f.Image),
			appDatabase.StrWithMaxLength(f.Title, 200),
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
			appLog.DebugError(fmt.Sprintf("An error occured while updating a feed (%d)", f.Id))
			return err
		}
	}

	log.Debug("Saving feeds")
	if len(f.Items) > 0 {
		for _, item := range f.Items {
			item.Feed = f
			err := item.Save()
			if err != nil {
				return err
			}
		}
	}

	log.Debug("Linking feed to its categories")
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

func GetAllFeeds(withFeedItems bool) ([]*Feed, error) {

	sql, err := appDatabase.NormalizedSql(`
		SELECT
			id,
			id_author,
			id_image,
			title,
			description,
			link,
			feed_link,
			updated,
			published,
			language,
			copyright,
			generator,
			last_update
		FROM feed
	`)
	if err != nil {
		appLog.DebugError(err)
		return nil, err
	}

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare SELECT statement")
		return nil, err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	rows, err := stmt.Query()
	if err != nil {
		appLog.DebugError("Unable to get result row")
		return nil, err
	}
	if rows.Err() != nil {
		appLog.DebugError("Unable to get result row")
		return nil, err
	}
	defer appDatabase.DeferRowsCloseFct(rows)()

	allRows := make([]*Feed, 0)

	for rows.Next() {
		var authorId, imageId *appDatabase.PrimaryKey
		var updatedRawValue, publishedRawValue, lastUpdateRawValue interface{}
		v := new(Feed)

		err = rows.Scan(
			&v.Id,
			&authorId,
			&imageId,
			&v.Title,
			&v.Description,
			&v.Link,
			&v.FeedLink,
			&updatedRawValue,
			&publishedRawValue,
			&v.Language,
			&v.Copyright,
			&v.Generator,
			&lastUpdateRawValue,
		)
		if err != nil {
			appLog.DebugError("Unable to affect results")
			return nil, err
		}

		v.Updated, err = appDatabase.SqlDateParse(updatedRawValue)
		if err != nil {
			appLog.DebugError("Unable to parse updated date")
			return nil, err
		}

		v.Published, err = appDatabase.SqlDateParse(publishedRawValue)
		if err != nil {
			appLog.DebugError("Unable to parse published date")
			return nil, err
		}

		v.LastUpdate, err = appDatabase.SqlDateParse(lastUpdateRawValue)
		if err != nil {
			appLog.DebugError("Unable to parse last update date")
			return nil, err
		}

		if authorId != nil {
			v.Author, err = authorById(*authorId)
			if err != nil {
				appLog.DebugError("Unable to fetch feed author")
				return nil, err
			}
		} else {
			v.Author = nil
		}

		if imageId != nil {
			v.Image, err = imageById(*imageId)
			if err != nil {
				appLog.DebugError("Unable to fetch feed image")
				return nil, err
			}
		} else {
			v.Image = nil
		}

		v.Categories, err = categoriesOfFeed(v)
		if err != nil {
			appLog.DebugError("Unable to fetch feed categories")
			return nil, err
		}

		if withFeedItems {

			v.Items, err = itemsOfFeed(v)
			if err != nil {
				appLog.DebugError("Unable to fetch feeds")
				return nil, err
			}
		} else {
			v.Items = nil
		}

		allRows = append(allRows, v)
	}

	return allRows, nil
}

func feedByTitle(title string) (*Feed, error) {

	sql, err := appDatabase.NormalizedSql(`
		SELECT
			id,
			id_author,
			id_image,
			title,
			description,
			link,
			feed_link,
			updated,
			published,
			language,
			copyright,
			generator,
			last_update
		FROM feed
		WHERE title = ?
	`)
	if err != nil {
		appLog.DebugError(err)
		return nil, err
	}

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare SELECT statement")
		return nil, err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	rows, err := stmt.Query(title)
	if err != nil {
		appLog.DebugError("Unable to get result row")
		return nil, err
	}
	if rows.Err() != nil {
		appLog.DebugError("Unable to get result row")
		return nil, err
	}
	defer appDatabase.DeferRowsCloseFct(rows)()

	if rows.Next() {
		var authorId, imageId *appDatabase.PrimaryKey
		var updatedRawValue, publishedRawValue, lastUpdateRawValue interface{}
		result := new(Feed)

		err = rows.Scan(
			&result.Id,
			&authorId,
			&imageId,
			&result.Title,
			&result.Description,
			&result.Link,
			&result.FeedLink,
			&updatedRawValue,
			&publishedRawValue,
			&result.Language,
			&result.Copyright,
			&result.Generator,
			&lastUpdateRawValue,
		)
		if err != nil {
			appLog.DebugError("Unable to affect results")
			return nil, err
		}

		result.Updated, err = appDatabase.SqlDateParse(updatedRawValue)
		if err != nil {
			appLog.DebugError("Unable to parse updated date")
			return nil, err
		}

		result.Published, err = appDatabase.SqlDateParse(publishedRawValue)
		if err != nil {
			appLog.DebugError("Unable to parse published date")
			return nil, err
		}

		result.LastUpdate, err = appDatabase.SqlDateParse(lastUpdateRawValue)
		if err != nil {
			appLog.DebugError("Unable to parse last update date")
			return nil, err
		}

		if authorId != nil {
			result.Author, err = authorById(*authorId)
			if err != nil {
				appLog.DebugError("Unable to fetch feed author")
				return nil, err
			}
		} else {
			result.Author = nil
		}

		if imageId != nil {
			result.Image, err = imageById(*imageId)
			if err != nil {
				appLog.DebugError("Unable to fetch feed image")
				return nil, err
			}
		} else {
			result.Image = nil
		}

		result.Categories, err = categoriesOfFeed(result)
		if err != nil {
			appLog.DebugError("Unable to fetch feed categories")
			return nil, err
		}

		result.Items, err = itemsOfFeed(result)
		if err != nil {
			appLog.DebugError("Unable to fetch feeds")
			return nil, err
		}

		return result, nil
	} else {
		return nil, nil
	}
}

package dbfeed

import (
	"errors"
	"fmt"
	"strings"
	"time"

	appDatabase "github.com/dademo/rssreader/modules/database"
	appLog "github.com/dademo/rssreader/modules/log"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

type FeedItem struct {
	Id          uint64
	Author      *FeedAuthor
	Image       *FeedImage
	Categories  []*FeedCategory
	Enclosures  []*FeedEnclosure
	Feed        *Feed
	Title       string
	Description string
	Content     string
	Link        string
	Updated     *time.Time
	Published   *time.Time
	GUID        string
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

func mapItems(items []*gofeed.Item) []*FeedItem {

	feedItems := make([]*FeedItem, 0, len(items))

	for _, item := range items {
		feedItems = append(feedItems, FromFeedItem(item))
	}

	return feedItems
}

func (f *FeedItem) Save() error {

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

	if len(f.Enclosures) > 0 {
		for _, enclosure := range f.Enclosures {
			err := enclosure.Save()
			if err != nil {
				appLog.DebugError("Unable to save a feed enclosure")
				return err
			}
		}
	}

	err := f.Normalize()
	if err != nil {
		appLog.DebugError("Unable to normalize item")
		return err
	}

	existingFeedItem, err := feedItemByGUID(f.GUID)
	if err != nil {
		appLog.DebugError("Unable to check for feed item existance")
		return err
	}

	if existingFeedItem != nil {
		f.Id = existingFeedItem.Id
	}

	if f.Id == 0 {

		log.Debug("Adding a new feed item")

		sql := `
			INSERT INTO feed_item (id_feed, id_author, id_image, title, description, content, link, updated, published, guid)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			appLog.DebugError("Unable to create the statement for feed item creation")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		newId, err := appDatabase.SqlExec(stmt,
			appDatabase.EntityId(f.Feed),
			appDatabase.EntityId(f.Author),
			appDatabase.EntityId(f.Image),
			f.Title,
			f.Description,
			f.Content,
			f.Link,
			f.Updated,
			f.Published,
			f.GUID,
		)

		if err != nil {
			appLog.DebugError("An error occured while saving a feed item")
			return err
		} else {
			f.Id = appDatabase.PrimaryKey(newId)
		}

	} else {

		log.Debug("Updating a feed item")

		sql := `
			UPDATE feed_item SET
				id_feed = ?,
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
			appLog.DebugError("Unable to create the statement for feed item creation")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		_, err = appDatabase.SqlExec(stmt,
			appDatabase.EntityId(f.Feed),
			appDatabase.EntityId(f.Author),
			appDatabase.EntityId(f.Image),
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
			appLog.DebugError(fmt.Sprintf("An error occured while updating a feed item (%d)", f.Id))
			return err
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

func (f *FeedItem) Normalize() error {

	const trimValues = " \t"
	if strings.Trim(f.GUID, trimValues) == "" {
		if strings.Trim(f.Link, trimValues) != "" {
			f.GUID = f.Link
			return nil
		}
		if strings.Trim(f.Title, trimValues) != "" {
			f.GUID = f.Title
			return nil
		}
		return errors.New(fmt.Sprintf("Unable to find a replacement GUID for item [%s:%s]", f.Feed.Title, f.Title))
	}
	return nil
}

func feedItemByGUID(guid string) (*FeedItem, error) {

	sql := `
		SELECT
			id,
			id_author,
			id_image,
			title,
			description,
			content,
			link,
			updated,
			published,
			guid
		FROM feed_item
		WHERE guid = ?
	`

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare SELECT statement")
		return nil, err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	rows, err := stmt.Query(guid)
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
		var updatedRawValue, publishedRawValue interface{}
		result := new(FeedItem)

		err = rows.Scan(
			&result.Id,
			&authorId,
			&imageId,
			&result.Title,
			&result.Description,
			&result.Content,
			&result.Link,
			&updatedRawValue,
			&publishedRawValue,
			&result.GUID,
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

		if authorId != nil {
			result.Author, err = authorById(*authorId)
			if err != nil {
				appLog.DebugError("Unable to fetch feed item author")
				return nil, err
			}
		} else {
			result.Author = nil
		}

		if imageId != nil {
			result.Image, err = imageById(*imageId)
			if err != nil {
				appLog.DebugError("Unable to fetch feed item image")
				return nil, err
			}
		} else {
			result.Image = nil
		}

		result.Categories, err = categoriesOfFeedItem(result)
		if err != nil {
			appLog.DebugError("Unable to fetch feed item categories")
			return nil, err
		}

		result.Enclosures, err = enclosuresOfFeedItem(result)
		if err != nil {
			appLog.DebugError("Unable to fetch feed item enclosure")
			return nil, err
		}

		return result, nil
	} else {
		return nil, nil
	}
}

func itemsOfFeed(feed *Feed) ([]*FeedItem, error) {
	return GetFeedItems(feed.Id)
}

func GetFeedItems(feedId appDatabase.PrimaryKey) ([]*FeedItem, error) {

	sql := `
		SELECT
			id,
			id_author,
			id_image,
			title,
			description,
			content,
			link,
			updated,
			published,
			guid
		FROM feed_item
		WHERE id_feed = ?
	`

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare SELECT statement")
		return nil, err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	rows, err := stmt.Query(feedId)
	if err != nil {
		appLog.DebugError("Unable to get result row")
		return nil, err
	}
	if rows.Err() != nil {
		appLog.DebugError("Unable to get result row")
		return nil, err
	}
	defer appDatabase.DeferRowsCloseFct(rows)()

	allValues := make([]*FeedItem, 0)
	for rows.Next() {
		var authorId, imageId *appDatabase.PrimaryKey
		var updatedRawValue, publishedRawValue interface{}
		v := new(FeedItem)

		err = rows.Scan(
			&v.Id,
			&authorId,
			&imageId,
			&v.Title,
			&v.Description,
			&v.Content,
			&v.Link,
			&updatedRawValue,
			&publishedRawValue,
			&v.GUID,
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

		if authorId != nil {
			v.Author, err = authorById(*authorId)
			if err != nil {
				appLog.DebugError("Unable to fetch feed item author")
				return nil, err
			}
		} else {
			v.Author = nil
		}

		if imageId != nil {
			v.Image, err = imageById(*imageId)
			if err != nil {
				appLog.DebugError("Unable to fetch feed item image")
				return nil, err
			}
		} else {
			v.Image = nil
		}

		v.Categories, err = categoriesOfFeedItem(v)
		if err != nil {
			appLog.DebugError("Unable to fetch feed item categories")
			return nil, err
		}

		v.Enclosures, err = enclosuresOfFeedItem(v)
		if err != nil {
			appLog.DebugError("Unable to fetch feed item enclosure")
			return nil, err
		}

		allValues = append(allValues, v)
	}
	return allValues, nil
}

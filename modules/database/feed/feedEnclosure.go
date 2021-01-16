package feed

import (
	"fmt"

	appDatabase "github.com/dademo/rssreader/modules/database"
	appLog "github.com/dademo/rssreader/modules/log"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

type FeedEnclosure struct {
	Id     uint64
	URL    string
	Length string
	Type   string
}

func FromEnclosure(enclosure *gofeed.Enclosure) *FeedEnclosure {
	if enclosure != nil {
		return &FeedEnclosure{
			Id:     0,
			URL:    enclosure.URL,
			Length: enclosure.Length,
			Type:   enclosure.Type,
		}
	} else {
		return nil
	}
}

func mapEnclosures(enclosures []*gofeed.Enclosure) []*FeedEnclosure {

	feedEnclosures := make([]*FeedEnclosure, 0, len(enclosures))

	for _, enclosure := range enclosures {
		feedEnclosures = append(feedEnclosures, FromEnclosure(enclosure))
	}

	return feedEnclosures
}

func (f *FeedEnclosure) Save() error {

	if f.Id == 0 {

		existing, err := feedEnclosureByUrl(f.URL)
		if err != nil {
			appLog.DebugError("Unable to get enclosure by url")
			return err
		}

		if existing != nil {
			f.Id = existing.Id
			return nil
		} else {

			log.Debug("Adding a new feed enclosure")

			sql := `
			INSERT INTO feed_enclosure (url, length, type)
			VALUES (?, ?, ?)
		`

			stmt, err := database.Prepare(sql)

			if err != nil {
				appLog.DebugError("Unable to create the statement for feed enclosure update")
				return err
			}
			defer appDatabase.DeferStmtCloseFct(stmt)()

			newId, err := appDatabase.SqlExec(stmt,
				f.URL,
				f.Length,
				f.Type,
			)

			if err != nil {
				appLog.DebugError("An error occured while saving a feed enclosure")
				return err
			} else {
				f.Id = appDatabase.PrimaryKey(newId)
				return nil
			}
		}
	} else {

		log.Debug("Updating a feed enclosure")

		sql := `
			UPDATE feed_enclosure SET
				url = ?,
				length = ?,
				type = ?
			WHERE id = ?
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			appLog.DebugError("Unable to create the statement for feed enclosure update")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		_, err = appDatabase.SqlExec(stmt,
			f.URL,
			f.Length,
			f.Type,
			f.Id,
		)

		if err != nil {
			appLog.DebugError(fmt.Sprintf("An error occured while updating a feed enclosure (%d)", f.Id))
			return err
		} else {
			return nil
		}
	}
}

func enclosuresOfFeedItem(feedItem *FeedItem) ([]*FeedEnclosure, error) {

	sql := `
	SELECT
		id,
		url,
		length,
		type
	FROM feed_enclosure
	INNER JOIN feed_enclosure_item
		ON feed_enclosure_item.id_feed_enclosure = feed_enclosure.id
	WHERE id_feed_item = ?
	`

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare statement")
		return nil, err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	rows, err := stmt.Query(feedItem.Id)
	if err != nil {
		appLog.DebugError("Unable to get result row")
		return nil, err
	}
	if rows.Err() != nil {
		appLog.DebugError("Unable to get result row")
		return nil, rows.Err()
	}
	defer appDatabase.DeferRowsCloseFct(rows)()

	allValues := make([]*FeedEnclosure, 0)
	for rows.Next() {

		v := new(FeedEnclosure)
		err = rows.Scan(&v.Id, &v.URL, &v.Length, &v.Type)
		if err != nil {
			appLog.DebugError("Unable to affect results")
			return nil, err
		} else {
			allValues = append(allValues, v)
		}
	}
	return allValues, nil
}

func feedEnclosureByUrl(url string) (*FeedEnclosure, error) {

	sql := `
		SELECT
			id,
			url,
			length,
			type
		FROM feed_enclosure
		WHERE url = ?
	`

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare statement")
		return nil, err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	v := new(FeedEnclosure)

	rows, err := stmt.Query(url)
	if err != nil {
		appLog.DebugError("Unable to get result row")
		return nil, err
	}
	if rows.Err() != nil {
		appLog.DebugError("Unable to get result row")
		return nil, rows.Err()
	}
	defer appDatabase.DeferRowsCloseFct(rows)()

	if rows.Next() {
		err = rows.Scan(&v.Id, &v.URL, &v.Length, &v.Type)
		if err != nil {
			appLog.DebugError("Unable to affect results")
			return nil, err
		} else {
			return v, nil
		}
	} else {
		return nil, nil
	}
}

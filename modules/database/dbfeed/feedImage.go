package dbfeed

import (
	"fmt"

	appDatabase "github.com/dademo/rssreader/modules/database"
	appLog "github.com/dademo/rssreader/modules/log"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

type FeedImage struct {
	Id    uint64
	URL   string
	Title string
}

func FromImage(image *gofeed.Image) *FeedImage {
	if image != nil {
		return &FeedImage{
			Id:    0,
			URL:   image.URL,
			Title: image.Title,
		}
	} else {
		return nil
	}
}

func (f *FeedImage) Save() error {

	existingFeedImage, err := imageByUrl(f.URL)
	if err != nil {
		appLog.DebugError("Unable to check for feed image existance")
		return err
	}

	if existingFeedImage != nil {
		f.Id = existingFeedImage.Id
	}

	if f.Id == 0 {

		log.Debug("Adding a new feed image")

		sql, err := appDatabase.NormalizedSql(`
			INSERT INTO feed_image (url, title)
			VALUES (?, ?)
		`)
		if err != nil {
			appLog.DebugError(err)
			return err
		}

		stmt, err := database.Prepare(appDatabase.PrepareExecSQL(sql))

		if err != nil {
			appLog.DebugError("Unable to create the statement for feed image update")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		newId, err := appDatabase.SqlExecGetId(stmt,
			f.URL,
			f.Title,
		)

		if err != nil {
			appLog.DebugError("An error occured while saving a feed image")
			return err
		} else {
			f.Id = appDatabase.PrimaryKey(newId)
			return nil
		}

	} else {

		log.Debug("Updating a feed image")

		sql, err := appDatabase.NormalizedSql(`
			UPDATE feed_image SET
				url = ?,
				title = ?
			WHERE id = ?
		`)
		if err != nil {
			appLog.DebugError(err)
			return err
		}

		stmt, err := database.Prepare(appDatabase.PrepareExecSQL(sql))

		if err != nil {
			appLog.DebugError("Unable to create the statement for feed image update")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		_, err = appDatabase.SqlExecGetId(stmt,
			f.URL,
			f.Title,
			f.Id,
		)

		if err != nil {
			appLog.DebugError(fmt.Sprintf("An error occured while updating a feed image (%d)", f.Id))
			return err
		} else {
			return nil
		}
	}
}

func imageById(imageId appDatabase.PrimaryKey) (*FeedImage, error) {

	sql, err := appDatabase.NormalizedSql(`
		SELECT
			id,
			url,
			title
		FROM feed_image
		WHERE id = ?
	`)
	if err != nil {
		appLog.DebugError(err)
		return nil, err
	}

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare statement")
		return nil, err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	rows, err := stmt.Query(imageId)
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

		v := new(FeedImage)
		err = rows.Scan(&v.Id, &v.URL, &v.Title)
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

func imageByUrl(url string) (*FeedImage, error) {

	sql, err := appDatabase.NormalizedSql(`
		SELECT
			id,
			url,
			title
		FROM feed_image
		WHERE url = ?
	`)
	if err != nil {
		appLog.DebugError(err)
		return nil, err
	}

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare statement")
		return nil, err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

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

		v := new(FeedImage)
		err = rows.Scan(&v.Id, &v.URL, &v.Title)
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

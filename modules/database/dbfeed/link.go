package dbfeed

import (
	"errors"

	appDatabase "github.com/dademo/rssreader/modules/database"
	appLog "github.com/dademo/rssreader/modules/log"
)

func linkFeedCategoryToFeed(category *FeedCategory, feed *Feed) error {

	if category.Id == 0 {
		return errors.New("You must provide a saved feed category")
	}

	if feed.Id == 0 {
		return errors.New("You must provide a saved feed")
	}

	sql, err := appDatabase.NormalizedSql(`
		SELECT
			COUNT(*) AS cnt
		FROM feed_category_feed
		WHERE 
			    id_feed_category = ?
			AND id_feed = ?
	`)
	if err != nil {
		appLog.DebugError(err)
		return err
	}

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare SELECT statement")
		return err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	row := stmt.QueryRow(category.Id, feed.Id)
	if row.Err() != nil {
		appLog.DebugError("Unable to get result row")
		return err
	}

	var cnt int
	err = row.Scan(&cnt)
	if err != nil {
		appLog.DebugError("Unable to affect results")
		return err
	}

	if cnt == 0 {
		// Adding value
		sql, err = appDatabase.NormalizedSql(`
			INSERT INTO feed_category_feed(id_feed_category, id_feed)
			VALUES (?, ?)
		`)
		if err != nil {
			appLog.DebugError(err)
			return err
		}

		stmt, err = database.Prepare(sql)
		if err != nil {
			appLog.DebugError("Unable to prepare INSERT statement")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		err := appDatabase.SqlExec(stmt, category.Id, feed.Id)
		if err != nil {
			appLog.DebugError("An error occured while saving a feed")
			return err
		}
	}
	return nil
}

func linkFeedCategoryToFeedItem(category *FeedCategory, item *FeedItem) error {

	if category.Id == 0 {
		return errors.New("You must provide a saved feed category")
	}

	if item.Id == 0 {
		return errors.New("You must provide a saved feed item")
	}

	sql, err := appDatabase.NormalizedSql(`
		SELECT
			COUNT(*) AS cnt
		FROM feed_category_item
		WHERE 
			    id_feed_category = ?
			AND id_feed_item = ?
	`)
	if err != nil {
		appLog.DebugError(err)
		return err
	}

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare SELECT statement")
		return err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	row := stmt.QueryRow(category.Id, item.Id)
	if row.Err() != nil {
		appLog.DebugError("Unable to get result row")
		return err
	}

	var cnt int
	err = row.Scan(&cnt)
	if err != nil {
		appLog.DebugError("Unable to affect results")
		return err
	}

	if cnt == 0 {
		// Adding value
		sql, err = appDatabase.NormalizedSql(`
			INSERT INTO feed_category_item(id_feed_category, id_feed_item)
			VALUES (?, ?)
		`)
		if err != nil {
			appLog.DebugError(err)
			return err
		}

		stmt, err = database.Prepare(sql)
		if err != nil {
			appLog.DebugError("Unable to prepare INSERT statement")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		err := appDatabase.SqlExec(stmt, category.Id, item.Id)
		if err != nil {
			appLog.DebugError("An error occured while saving a feed")
			return err
		}
	}
	return nil
}

func linkFeedEnclosureToFeedItem(enclosure *FeedEnclosure, item *FeedItem) error {

	if enclosure.Id == 0 {
		return errors.New("You must provide a saved feed enclosure")
	}

	if item.Id == 0 {
		return errors.New("You must provide a saved feed item")
	}

	sql, err := appDatabase.NormalizedSql(`
		SELECT
			COUNT(*) AS cnt
		FROM feed_enclosure_item
		WHERE 
			    id_feed_enclosure = ?
			AND id_feed_item = ?
	`)
	if err != nil {
		appLog.DebugError(err)
		return err
	}

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare SELECT statement")
		return err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	row := stmt.QueryRow(enclosure.Id, item.Id)
	if row.Err() != nil {
		appLog.DebugError("Unable to get result row")
		return err
	}

	var cnt int
	err = row.Scan(&cnt)
	if err != nil {
		appLog.DebugError("Unable to affect results")
		return err
	}

	if cnt == 0 {
		// Adding value
		sql, err = appDatabase.NormalizedSql(`
			INSERT INTO feed_enclosure_item(id_feed_enclosure, id_feed_item)
			VALUES (?, ?)
		`)
		if err != nil {
			appLog.DebugError(err)
			return err
		}

		stmt, err = database.Prepare(sql)
		if err != nil {
			appLog.DebugError("Unable to prepare INSERT statement")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		err := appDatabase.SqlExec(stmt, enclosure.Id, item.Id)
		if err != nil {
			appLog.DebugError("An error occured while saving a feed")
			return err
		}
	}
	return nil
}

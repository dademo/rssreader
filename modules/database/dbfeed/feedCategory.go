package dbfeed

import (
	"fmt"

	appDatabase "github.com/dademo/rssreader/modules/database"
	appLog "github.com/dademo/rssreader/modules/log"

	log "github.com/sirupsen/logrus"
)

type FeedCategory struct {
	Id       uint64 `json:"id"`
	Category string `json:"category"`
}

func FromCategory(category string) *FeedCategory {
	return &FeedCategory{
		Id:       0,
		Category: category,
	}
}

func mapCategories(categories []string) []*FeedCategory {

	feedCategories := make([]*FeedCategory, 0, len(categories))

	for _, category := range categories {
		feedCategories = append(feedCategories, FromCategory(category))
	}

	return feedCategories
}

func (f *FeedCategory) Save() error {

	if f.Id == 0 {

		existing, err := feedCategoryByCategory(f.Category)
		if err != nil {
			appLog.DebugError("Unable to get category by name")
			return err
		}

		if existing != nil {
			f.Id = existing.Id
			return nil
		} else {

			log.Debug("Adding a new feed category")

			sql, err := appDatabase.NormalizedSql(`
				INSERT INTO feed_category (category)
				VALUES (?)
			`)
			if err != nil {
				appLog.DebugError(err)
				return err
			}

			stmt, err := database.Prepare(appDatabase.PrepareExecSQL(sql))

			if err != nil {
				appLog.DebugError("Unable to create the statement for feed category update")
				return err
			}
			defer appDatabase.DeferStmtCloseFct(stmt)()

			newId, err := appDatabase.SqlExecGetId(stmt,
				f.Category,
			)

			if err != nil {
				appLog.DebugError("An error occured while saving a feed category")
				return err
			} else {
				f.Id = appDatabase.PrimaryKey(newId)
				return nil
			}
		}

	} else {

		log.Debug("Updating a feed category")

		sql, err := appDatabase.NormalizedSql(`
			UPDATE feed_category SET
				category = ?
			WHERE id = ?
		`)
		if err != nil {
			appLog.DebugError(err)
			return err
		}

		stmt, err := database.Prepare(appDatabase.PrepareExecSQL(sql))

		if err != nil {
			appLog.DebugError("Unable to create the statement for feed category update")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		_, err = appDatabase.SqlExecGetId(stmt,
			f.Category,
			f.Id,
		)

		if err != nil {
			appLog.DebugError(fmt.Sprintf("An error occured while updating a feed category (%d)", f.Id))
			return err
		} else {
			return nil
		}
	}
}

func feedCategoryByCategory(category string) (*FeedCategory, error) {

	sql, err := appDatabase.NormalizedSql(`
		SELECT
			id,
			category
		FROM feed_category
		WHERE category = ?
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

	v := new(FeedCategory)

	rows, err := stmt.Query(category)
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
		err = rows.Scan(&v.Id, &v.Category)
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

func categoriesOfFeedItem(feedItem *FeedItem) ([]*FeedCategory, error) {

	sql, err := appDatabase.NormalizedSql(`
		SELECT
			id,
			category
		FROM feed_category
		INNER JOIN feed_category_item
			ON feed_category_item.id_feed_category = feed_category.id
		WHERE id_feed_item = ?
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

	allValues := make([]*FeedCategory, 0)
	for rows.Next() {

		v := new(FeedCategory)
		err = rows.Scan(&v.Id, &v.Category)
		if err != nil {
			appLog.DebugError("Unable to affect results")
			return nil, err
		} else {
			allValues = append(allValues, v)
		}
	}
	return allValues, nil
}

func categoriesOfFeed(feed *Feed) ([]*FeedCategory, error) {

	sql, err := appDatabase.NormalizedSql(`
		SELECT
			id,
			category
		FROM feed_category
		INNER JOIN feed_category_feed
			ON feed_category_feed.id_feed_category = feed_category.id
		WHERE id_feed = ?
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

	rows, err := stmt.Query(feed.Id)
	if err != nil {
		appLog.DebugError("Unable to get result row")
		return nil, err
	}
	if rows.Err() != nil {
		appLog.DebugError("Unable to get result row")
		return nil, rows.Err()
	}
	defer appDatabase.DeferRowsCloseFct(rows)()

	allValues := make([]*FeedCategory, 0)
	for rows.Next() {

		v := new(FeedCategory)
		err = rows.Scan(&v.Id, &v.Category)
		if err != nil {
			appLog.DebugError("Unable to affect results")
			return nil, err
		} else {
			allValues = append(allValues, v)
		}
	}
	return allValues, nil
}

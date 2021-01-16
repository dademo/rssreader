package dbfeed

import (
	"fmt"

	appDatabase "github.com/dademo/rssreader/modules/database"
	appLog "github.com/dademo/rssreader/modules/log"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

type FeedAuthor struct {
	Id    uint64
	Name  string
	Email string
}

func FromPerson(author *gofeed.Person) *FeedAuthor {
	if author != nil {
		return &FeedAuthor{
			Id:    0,
			Name:  author.Name,
			Email: author.Email,
		}
	} else {
		return nil
	}
}

func (f *FeedAuthor) Save() error {

	if f.Id == 0 {

		existing, err := feedAuthorByName(f.Name)
		if err != nil {
			appLog.DebugError("Unable to get author by name")
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
				appLog.DebugError("Unable to create the statement for feed author update")
				return err
			}
			defer appDatabase.DeferStmtCloseFct(stmt)()

			newId, err := appDatabase.SqlExec(stmt,
				f.Name,
				f.Email,
			)

			if err != nil {
				appLog.DebugError("An error occured while saving a feed author")
				return err
			} else {
				f.Id = appDatabase.PrimaryKey(newId)
				return nil
			}
		}

	} else {

		log.Debug("Updating a feed author")

		sql := `
			UPDATE feed_author SET
				name = ?,
				email = ?
			WHERE id = ?
		`

		stmt, err := database.Prepare(sql)

		if err != nil {
			appLog.DebugError("Unable to create the statement for feed author update")
			return err
		}
		defer appDatabase.DeferStmtCloseFct(stmt)()

		_, err = appDatabase.SqlExec(stmt,
			f.Name,
			f.Email,
			f.Id,
		)

		if err != nil {
			appLog.DebugError(fmt.Sprintf("An error occured while updating a feed author (%d)", f.Id))
			return err
		} else {
			return nil
		}
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
		appLog.DebugError("Unable to prepare statement")
		return nil, err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	rows, err := stmt.Query(name)
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

		v := new(FeedAuthor)
		err = rows.Scan(&v.Id, &v.Name, &v.Email)
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

func authorById(authorId appDatabase.PrimaryKey) (*FeedAuthor, error) {

	sql := `
	SELECT
		id,
		name,
		email
	FROM feed_author
	WHERE id = ?
	`

	stmt, err := database.Prepare(sql)
	if err != nil {
		appLog.DebugError("Unable to prepare statement")
		return nil, err
	}
	defer appDatabase.DeferStmtCloseFct(stmt)()

	rows, err := stmt.Query(authorId)
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

		v := new(FeedAuthor)
		err = rows.Scan(&v.Id, &v.Name, &v.Email)
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

package dbfeed

import (
	"database/sql"
	"fmt"

	appDatabase "github.com/dademo/rssreader/modules/database"
	appLog "github.com/dademo/rssreader/modules/log"

	log "github.com/sirupsen/logrus"
)

func getFeedSQL() []string {
	return []string{
		`
		CREATE TABLE feed_category (
			id			{{.SqlPrimaryKey}},
			category	TEXT
		);`,
		`
		CREATE TABLE feed_author (
			id			{{.SqlPrimaryKey}},
			name		TEXT,
			email		TEXT
		);`,
		`
		CREATE TABLE feed_image (
			id			{{.SqlPrimaryKey}},
			url			TEXT,
			title		TEXT
		);`,
		`
		CREATE TABLE feed_enclosure (
			id		{{.SqlPrimaryKey}},
			url		TEXT,
			length	TEXT,
			type	TEXT
		);`,
		`
		CREATE TABLE feed (
			id			{{.SqlPrimaryKey}},
			id_author	INTEGER REFERENCES feed_author(id),
			id_image	INTEGER REFERENCES feed_image(id),
			title		VARCHAR(200) NOT NULL UNIQUE,
			description	TEXT,
			link		TEXT,
			feed_link	TEXT,
			updated	 	{{.SqlTimestamp}},
			published	{{.SqlTimestamp}},
			language	TEXT,
			copyright	TEXT,
			generator	TEXT,
			last_update	{{.SqlTimestamp}}
		);`,
		`
		CREATE TABLE feed_item (
			id				{{.SqlPrimaryKey}},
			id_feed			INTEGER REFERENCES feed(id),
			id_author		INTEGER REFERENCES feed_author(id),
			id_image		INTEGER REFERENCES feed_image(id),
			title			TEXT,
			description		TEXT,
			content			TEXT,
			link			TEXT,
			updated			{{.SqlTimestamp}},
			published		{{.SqlTimestamp}},
			guid			TEXT
		);`,
		`
		CREATE TABLE feed_category_feed (
			id_feed_category	INTEGER NOT NULL REFERENCES feed_category(id),
			id_feed				INTEGER NOT NULL REFERENCES feed(id),
			UNIQUE(id_feed_category, id_feed)
		);`,
		`
		CREATE TABLE feed_category_item (
			id_feed_category	INTEGER NOT NULL REFERENCES feed_category(id),
			id_feed_item		INTEGER NOT NULL REFERENCES feed_item(id),
			UNIQUE(id_feed_category, id_feed_item)
		);`,
		`
		CREATE TABLE feed_enclosure_item (
			id_feed_enclosure	INTEGER NOT NULL REFERENCES feed_enclosure(id),
			id_feed_item		INTEGER NOT NULL REFERENCES feed_item(id),
			UNIQUE(id_feed_enclosure, id_feed_item)
		);`,
	}
}

var feedModuleDef = appDatabase.DatabaseModuleTableCreationDef{
	ModuleName:                 "Feed",
	Version:                    "0.0.1",
	DatabaseModuleTableCreator: databaseFeedModuleCreator,
	DatabaseModuleTableUpdater: databaseFeedModuleUpdater,
}

var database *sql.DB

func init() {
	appDatabase.RegisterDatabaseTableCreator(feedModuleDef)
	appDatabase.RegisterOnDatabaseSet(onDatabaseSet)
}

func databaseFeedModuleCreator(connection *sql.Tx) error {

	log.Debug("Creating feed tables")

	for _, row := range getFeedSQL() {
		sql, err := appDatabase.NormalizedSql(row)

		if err != nil {
			return err
		}

		log.Debug(fmt.Sprintf("Running command :\n%s", sql))

		_, err = connection.Exec(sql)
		if err != nil {
			appLog.DebugError(err, "Unable to create feed tables")
			return err
		}
	}

	log.Debug("Feed tables created")
	return nil
}

func databaseFeedModuleUpdater(connection *sql.Tx, oldVersion string) error {
	return nil
}

func onDatabaseSet(db *sql.DB) {
	database = db
}

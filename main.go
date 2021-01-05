package main

import (
	"os"
	"sort"

	"github.com/dademo/rssreader/cmd"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Test struct {
	A int
	b int
}

var (
	appVersion = "development"
)

func main() {

	app := cli.NewApp()
	app.Name = "RSS reader"
	app.Usage = "A rss stream sync service"
	app.Description = "An application syncing your rss streams"
	app.Version = appVersion

	app.Flags = []cli.Flag{
		cmd.FlagLogLevel,
		cmd.FlagLogFile,
		cmd.FlagConfig,
	}

	app.Commands = []cli.Command{
		cmd.CmdConfig,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	/*
		var b interface{}
		err := yaml.Unmarshal([]byte("a: 1\nb: 2"), &b)

		fmt.Println(b, err)

		parsed, ok := b.(map[string]interface{})
		fmt.Println(parsed, ok)
		fmt.Println(parsed["a"])
		fmt.Println(parsed["b"])

		var c Test
		err2 := yaml.Unmarshal([]byte("a: 1\nb: 2"), &c)
		fmt.Println(c, err2)

		fp := gofeed.NewParser()
		feed, _ := fp.ParseURL("http://feeds.twit.tv/twit.xml")
		fmt.Println(feed.Title)
	*/
}

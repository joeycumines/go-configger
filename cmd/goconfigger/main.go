package main

import (
	"os"
	"gopkg.in/urfave/cli.v1"
)

var (
	AppName      = `goconfigger`
	AppUsage     = `output a modified configuration file, allowing merging, modification, and conversion`
	AppUsageText = ``
	AppAction    = appAction
	AppArgs      = os.Args
	AppFlags     = appFlags
)

func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Usage = AppUsage
	app.Flags = AppFlags()
	app.Action = AppAction
	app.UsageText = AppUsageText
	app.Run(AppArgs)
}

func appAction(c *cli.Context) error {
	return nil
}

func appFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringSliceFlag{
			Name:  "lang",
			Usage: "language for the greeting",
		},
	}
}

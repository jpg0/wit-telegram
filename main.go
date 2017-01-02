package main

import (
	"os"
	"fmt"
	"strings"
	"github.com/Sirupsen/logrus"
	"github.com/juju/errors"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "wit-telegram"
	app.Usage = "Bridge Telegram & Wit.ai"
	app.Version = "1.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "wittoken",
			Usage: "Token to auth with Wit.ai",
		},
		cli.StringFlag{
			Name: "telegramtoken",
			Usage: "Token to auth with Telegram",
		},
		cli.StringFlag{
			Name: "actionserver",
			Usage: "URL to server executing actions",
		},
		cli.StringFlag{
			Name: "loglevel",
			Usage: "Logging level",
			Value: "info",
		},
	}
	app.Action = verbose(ConfigureAndStart)
	app.Run(os.Args)
}

func verbose(next func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		err := next(c)

		if err != nil {
			fmt.Println(errors.ErrorStack(err))
		}

		return err
	}
}

func initLogging(level string) error {
	switch strings.ToLower(level) {
	case "debug": logrus.SetLevel(logrus.DebugLevel)
	case "info": logrus.SetLevel(logrus.InfoLevel)
	case "warn": logrus.SetLevel(logrus.WarnLevel)
	case "error": logrus.SetLevel(logrus.ErrorLevel)
	case "fatal": logrus.SetLevel(logrus.FatalLevel)
	default:
		return errors.Errorf("Unknown logging level: %v", level)
	}

	return nil
}

func ConfigureAndStart(c *cli.Context) error {

	err := initLogging(c.String("loglevel"))

	if err != nil {
		return errors.Annotate(err, "Failed to initialize logging")
	}

	var ac ActionClient

	actionserver := c.String("actionserver")

	if (actionserver == "") {
		panic("No action server specified")
	} else {
		ac = NewRemoteActionClient(actionserver)
	}

	b, err := NewBridge(c.String("telegramtoken"), c.String("wittoken"), ac)

	if err != nil {
		return errors.Annotate(err, "Failed to create bridge")
	}

	err = b.Start()

	if err != nil {
		return errors.Annotate(err, "Failed to start bridge")
	}

	return nil
}

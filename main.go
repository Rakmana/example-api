package main

import (
	"log"
	"os"

	"github.com/RichardKnop/recall/commands"
	"github.com/RichardKnop/recall/config"
	"github.com/RichardKnop/recall/database"
	"github.com/codegangsta/cli"
)

var (
	cliApp *cli.App
)

func init() {
	// Initialise a CLI app
	cliApp = cli.NewApp()
	cliApp.Name = "recall"
	cliApp.Usage = "Recall"
	cliApp.Author = "Richard Knop"
	cliApp.Email = "risoknop@gmail.com"
	cliApp.Version = "0.0.0"
}

func main() {
	// Load the configuration, connect to the database
	cnf := config.NewConfig(
		true, // must load once
		true, // keep reloading
	)
	db, err := database.NewDatabase(cnf)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Disable logging
	db.LogMode(false)

	// Set the CLI app commands
	cliApp.Commands = []cli.Command{
		{
			Name:  "migrate",
			Usage: "run migrations",
			Action: func(c *cli.Context) {
				if err := commands.Migrate(db); err != nil {
					log.Fatal(err)
				}
			},
		},
		{
			Name:  "loaddata",
			Usage: "load data from fixture",
			Action: func(c *cli.Context) {
				if err := commands.LoadData(c.Args(), cnf, db); err != nil {
					log.Fatal(err)
				}
			},
		},
		{
			Name:  "runserver",
			Usage: "run web server",
			Action: func(c *cli.Context) {
				commands.RunServer(cnf, db)
			},
		},
	}

	// Run the CLI app
	cliApp.Run(os.Args)
}

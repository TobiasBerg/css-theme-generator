package main

import (
	"context"
	"log"
	"os"

	"github.com/TobiasBerg/theme-generator/cmd"
	"github.com/urfave/cli/v3"
)

func main() {
	generateCMD := &cli.Command{
		Name:   "generate-theme",
		Usage:  "Generate a theme from a color palette",
		Action: cmd.CreateTheneCMD(),
	}

	app := cli.Command{
		Name:           "theme-generator",
		Description:    "Generate a theme from a color palette",
		DefaultCommand: "generate-theme",
		Commands: []*cli.Command{
			generateCMD,
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

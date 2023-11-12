package main

import (
	"fmt"
	"log"
	"os"
	"github.com/gwpp/tinify-go/tinify"
	"github.com/urfave/cli/v2"
)

var (
	debugLevel int	// debug level, as set by the user with -d -d -d ...
)

func main() {
	// start app
	app := &cli.App{
		Name:		"tinify-go",
		Usage:		"Compresses/converts images using the TinyPNG API.",
		UsageText:	"See ",
		Version:	Tinify.VERSION,
		// DefaultCommand: "translate",	// to avoid brealing compatibility with earlier versions.
		EnableBashCompletion: true,
		// Compiled:
		Authors: []*cli.Author{
			{
				Name: "gwpp",
				Email: "ganwenpeng1993@163.com",
			},
			{
				Name: "Gwyneth Llewelyn",
				Email: "gwyneth.llewelyn@gwynethllewelyn.net",
			},
		},
		Copyright: "© 2017-2023 by gwpp. All rights reserved. Freely distributed under a MIT license.\nThis software is not affiliated nor endorsed by TinyPNG.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "source_lang",
				Aliases: []string{"s"},
				Usage:   "Set source language without using the settings file",
				Value:	 "EN",
//				Destination:	&setting.SourceLang,
			},
			&cli.StringFlag{
				Name:    "target_lang",
				Aliases: []string{"t"},
				Usage:   "Set target language without using the settings file",
				Value:	 "JA",
//				Destination:	&setting.TargetLang,
			},
			&cli.BoolFlag{
				Name:    "pro",
				Usage:   "Use Pro plan's endpoint?",
				Value:   false,
//				Destination: &setting.IsPro,
			},
			&cli.BoolFlag{
				Name:	"debug",
				Aliases: []string{"d"},
				Usage:	"Debugging; repeating the flag increases verbosity.",
				Count:	&debugLevel,
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "translate",
				Aliases:     []string{"trans"},
				Usage:       "Basic translation of a set of Unicode strings into another language",
				Description: "Text to be translated.\nOnly UTF-8-encoded plain text is supported. May contain multiple sentences, but the total request body size must not exceed 128 KiB (128 · 1024 bytes).\nPlease split up your text into multiple	calls if it exceeds this limit.",
				Category:	 "Translations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "tag_handling",
						Usage:       "Set to XML or HTML in order to do more advanced parsing (empty means just using the plain text variant)",
						Aliases:     []string{"tag"},
						Value:       "",
//						Destination:	&setting.TagHandling,
						Action: func(c *cli.Context, v string) error {
							switch v {
								case "xml", "html":
									return nil
								default:
									return fmt.Errorf("tag_handling must be either `xml` or `html` (got: %s)",
v)							}
						},
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
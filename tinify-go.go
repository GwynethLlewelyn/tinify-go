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
		UsageText:	"See https://tinypng.com/developers",
		Version:	Tinify.VERSION,
		DefaultCommand: "compress",
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
				Name:    "input",
				Aliases: []string{"i"},
				Usage:   "Input file name or URL",
				Value:	 "",
//				Destination:	&setting.SourceLang,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file name; if ommitted, writes to standard output",
				Value:	 "",
//				Destination:	&setting.TargetLang,
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
				Name:        "compress",
				Aliases:     []string{"c"},
				Usage:       "Compress images",
				Description: "You can upload any WebP, JPEG or PNG image to the Tinify API to compress it. We will automatically detect the type of image and optimise with the TinyPNG or TinyJPG engine accordingly. Compression will start as soon as you upload a file or provide the URL to the image.",
//				Category:	 "Translations",
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
			{
				Name:        "resize",
				Aliases:     []string{"r"},
				Usage:       "Resize images",
				Description: "Use the API to create resized versions of your uploaded images. By letting the API handle resizing you avoid having to write such code yourself and you will only have to upload your image once. The resized images will be optimally compressed with a nice and crisp appearance.\nYou can also take advantage of intelligent cropping to create thumbnails that focus on the most visually important areas of your image.\nResizing counts as one additional compression. For example, if you upload a single image and retrieve the optimized version plus 2 resized versions this will count as 3 compressions in total.",
//				Category:	 "Resizing",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "method",
						Usage:       fmt.Sprintf("Valid methods are: `%s`, `%s`, `%s`, or `%s`", Tinify.ResizeMethodScale, Tinify.ResizeMethodFit, Tinify.ResizeMethodCover, Tinify.ResizeMethodThumb),
						Aliases:     []string{"m"},
						Value:       Tinify.ResizeMethodScale,
//						Destination:	&setting.TagHandling,
						Action: func(c *cli.Context, method string) error {
							switch method {
							case Tinify.ResizeMethodScale, Tinify.ResizeMethodFit, Tinify.ResizeMethodCover, Tinify.ResizeMethodThumb:
								return nil
							default:
								return fmt.Errorf("method must be one of `%s`, `%s`, `%s`, or `%s` (got: %s)", Tinify.ResizeMethodScale, Tinify.ResizeMethodFit, Tinify.ResizeMethodCover, Tinify.ResizeMethodThumb, method)
							}
						},
					},
				},
			},
			{
				Name:        "convert",
				Aliases:     []string{"t"},
				Usage:       "Convert between image types",
				Description: "You can use the API to convert your images to your desired image type. Tinify currently supports converting between WebP, JPEG, and PNG. When you provide more than one image type in your convert request, the smallest version will be returned to you.\nImage converting will count as one additional compression.",
//				Category:	 "Conversion",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:        "image-type",
						Usage:       "",
						Aliases:     []string{"m"},
						Value:       Tinify.ExtensionWebP,
//						Destination:	&setting.TagHandling,
						Action: func(c *cli.Context, types []string) error {
							// check if we have gotten a valid selection of types
							return nil
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
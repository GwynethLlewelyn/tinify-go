package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"github.com/gwpp/tinify-go/tinify"
	"github.com/urfave/cli/v2"
)

var (
	debugLevel int			// debug level, as set by the user with -d -d -d ...
	imageName string		// filename or URL.
	outputFileName string	// if set, it's the output filename; if not, well...
)

func main() {
	// start app
	app := &cli.App{
		Name:		"tinify-go",
		Usage:		"Compresses/converts/resizes images using the TinyPNG API.",
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
		Copyright: "© 2017-2023 by gwpp. All rights reserved. Freely distributed under a MIT license.\nThis software is neither affiliated with, nor endorsed by Tinify B.V.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i", "url"},
				Usage:   "Input `file` name or URL; if omitted, reads from standard input",
				Value:	 "",
				Destination:	&imageName,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output `file` name; if ommitted, writes to standard output",
				Value:	 "",
				Destination:	&outputFileName,
			},
			&cli.BoolFlag{
				Name:	"debug",
				Aliases: []string{"d"},
				Usage:	"Debugging; repeating the flag increases verbosity.",
				Count:	&debugLevel,
			},
			&cli.StringFlag{
				Name:	"key",
				Aliases: []string{"k"},
				Usage:	"The `value` of your Tinify API key; you can also read it from the environment variable ",
				Value:	"",
				EnvVars: []string{"TINIFY_API_KEY"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "compress",
				Aliases:     []string{"c"},
				Usage:       "Compress images",
				Description: "You can upload any WebP, JPEG or PNG image to the Tinify API to compress it. We will automatically detect the type of image and optimise with the TinyPNG or TinyJPG engine accordingly. Compression will start as soon as you upload a file or provide the URL to the image.",
//				Category:	 "Translations",

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
						Usage:       fmt.Sprintf("Valid `method`s are: `%s`, `%s`, `%s`, or `%s`", Tinify.ResizeMethodScale, Tinify.ResizeMethodFit, Tinify.ResizeMethodCover, Tinify.ResizeMethodThumb),
						Aliases:     []string{"m"},
						Value:       Tinify.ResizeMethodScale,
//						Destination:	&setting.TagHandling,
						Action: func(c *cli.Context, method string) error {
							switch method {
							case Tinify.ResizeMethodScale, Tinify.ResizeMethodFit, Tinify.ResizeMethodCover, Tinify.ResizeMethodThumb:
								return nil
							default:
								return cli.Exit(fmt.Sprintf("method must be one of `%s`, `%s`, `%s`, or `%s` (got: %s)", Tinify.ResizeMethodScale, Tinify.ResizeMethodFit, Tinify.ResizeMethodCover, Tinify.ResizeMethodThumb, method), 4)
							}
						},
					},
					&cli.Int64Flag{
						Name:        "height",
						Usage:       "Image height",
						Aliases:     []string{"e"},
						Value:       100,
//						Destination:	&setting.height,
						Action: func(c *cli.Context, value int64) error {
							if value < 1 {
								return cli.Exit("height must be bigger than 1 pixel", 5)
							}
							return nil
						},
					},
					&cli.Int64Flag{
						Name:        "width",
						Usage:       "Image width",
						Aliases:     []string{"w"},
						Value:       100,
//						Destination:	&setting.width,
						Action: func(c *cli.Context, value int64) error {
							if value < 1 {
								return cli.Exit("width must be bigger than 1 pixel", 5)
							}
							return nil
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
						Usage:       "Valid image `type`s are: `webp`, `png`, `jpg`",
						Aliases:     []string{"g"},
//						Value:       "webp",	// cannot be just one!
//						DefaultText: "webp",
						Value: 		 cli.NewStringSlice("webp"),
//						Destination:	&setting.TagHandling,
						Action: func(c *cli.Context, types []string) error {
							// check if we have gotten a valid selection of types
							for _, str := range types {
								switch str {
									case "webp", "png", "jpg":
										// any of these are valid; continue looping
										continue
									default:
										return cli.Exit(fmt.Sprintf("invalid file format: %s", str), 3)
								}
							}
							return nil
						},
					},
				},
			},
		}, // end commands
		Action: func(c *cli.Context) error {
			// TODO(gwyneth): Create constants for debugging levels.
			if debugLevel > 1 {
				fmt.Fprintf(os.Stderr, "number of args (Narg): %d, c.Args.Len(): %d\n", c.NArg(), c.Args().Len())
			}
			// 0 arguments: ok, file comes from STDIN,
			// 1 argument:  ok, file comes either from local disk or is an URL to be sent to TinyPNG.
			// 2 or more:   invalid, we can only send one at the time. Maybe we'll loosen this at a later stage.
			if c.NArg() >= 2 {
				return cli.Exit("cannot specify multiple file paths or URLs", 2)
			}

			var (
				rawImage []byte			// raw image file, when loaded from disk.
				err error				// declared here due to scope issues.
				f = os.Stdin			// file handler; STDIN by default.
				isURL = false			// do we have an URL? (false means it's a file)
				source *Tinify.Source	// declared in advance to avoid scoping issues.
			)

			// check if we have a valid Tinify API key
			if len(c.String("key")) == 0 {
				return cli.Exit("the Tinify API key was not found anywhere (tried environment and CLI flags); cannot proceed", 2)
			}

			// Set the API key:
			Tinify.SetKey(c.String("key"))

			// theoretically, theoretically, one might do:
			//   `echo "https://example.com/myimage.png" | tinify-go compress`
			//  and expect it to work; we leave that for a future release. (gwyneth 20231130)
			if imageName == "" && c.NArg() == 1 {
				// check if it's URL or filename
				imageName = c.Args().First()
			}

			// `://` is hardly a valid filename, but a requirement for being an URL:
			isURL = strings.Contains(imageName, "://")
			// handle URL later
			if !isURL {
				f, err = os.Open(imageName)
				if err != nil {
					return err
				}

				rawImage, err = io.ReadAll(f)
				if err != nil {
					return err
				}

				if debugLevel > 1 {
					fmt.Fprintln(os.Stderr, "Arg: ", imageName, " Size: ", len(rawImage))
				}

				// Now call the TinyPNG API
				source, err = Tinify.FromBuffer(rawImage)
				if err != nil {
					return err
				}
			} else {
				source, err = Tinify.FromUrl(imageName)
				if err != nil {
					return err
				}
			}

			// commands are cumulative! Or ar least some are
			switch c.Command.Name {
				case "compress":
					fallthrough
				case "resize":
					err = source.Resize(&Tinify.ResizeOption{
						Method: Tinify.ResizeMethod(c.String("method")),
						Width:  c.Int64("width"), // replace by real value!
						Height: c.Int64("height"),
					})
					fallthrough
				case "convert":
					fallthrough
				default:
			}
			if err != nil {
				return err
			}

			// figure out the output filename, if any
			if len(c.String("output")) == 0 {
				rawImage, err := source.ToBuffer()
				if err != nil {
					return err
				}
				// rawImage contains the raw image data; we push it out to STDOUT
				n, err := io.WriteString(os.Stdout, string(rawImage))
				if err != nil {
					return err
				}
				if debugLevel > 1 {
					fmt.Fprintln(os.Stderr, "wrote ", n, " bytes to stdout")
				}
				return nil
			}

			outputFileName = c.String("output")
			if debugLevel > 1 {
				fmt.Fprintf(os.Stderr, "opening file %q for outputting image\n", outputFileName)
			}
			// write to file, we have a special function for that already defined:
			err = source.ToFile(outputFileName)
			if err != nil {
				return err
			}



			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
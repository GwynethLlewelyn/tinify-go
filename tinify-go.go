// Original Go Tinify library: Copyright (c) 2017 gwpp
// Distributed under a MIT licence.
// Additional coding and CLI example (c) 2025 by Gwyneth Llewelyn,
// also under a MIT licence.
package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/GwynethLlewelyn/justify"
	Tinify "github.com/gwpp/tinify-go/tinify"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

// No harm is done having just one context, which is simply the background.
var ctx = context.Background()

// Type to hold the global variables for all possible calls.
type Setting struct {
	LoggingLevel     string         `json:"log_level"`         // Debug/verbosity level, "error" by default
	ImageName        string         `json:"image_name"`        // Filename or URL.
	OutputFileName   string         `json:"output_file_name"`  // If set, it's the output filename; if not, well...
	FileType         string         `json:"file_type"`         // Any set of webp, png, jpg, avif.
	Key              string         `json:"key"`               // TinyPNG API key; can be on environment or read from `.env`.
	Logger           zerolog.Logger `json:"-"`                 // The main setting.Logger.
	Method           string         `json:"method"`            // Resizing method (scale, fit, cover, thumb).
	Width            int64          `json:"width"`             // Image width  (for resize operations).
	Height           int64          `json:"height"`            // Image height (  "   "      "    "  ).
	Transform        string         `json:"transform"`         // Transform the background to one of 'black', 'white', or hex value.
	TerminalWidth    int            `json:"terminal_width"`    // If we're on a TTY, stores the width; 80 is default.
	CompressionCount int64          `json:"compression_count"` // A measure of how many crdits are still left for further compression.
}

// Global settings for this CLI app.
var setting Setting

// Tinify API supported file types.
// Add more when TinyPNG supports additional types.
var types = []string{
	"png",
	"jpeg",
	"webp",
	"avif",
}

// Available image resizing methods.
// Add more when TinyPNG supports additional types.
var methods = []string{
	Tinify.ResizeMethodScale,
	Tinify.ResizeMethodFit,
	Tinify.ResizeMethodCover,
	Tinify.ResizeMethodThumb,
}

/*
				(_)
 _ __ ___   __ _ _ _ __
| '_ ` _ \ / _` | | '_ \
| | | | | | (_| | | | | |
|_| |_| |_|\__,_|_|_| |_|

*/
// Main starts here.
func main() {
	var err error // declared here due to scoping issues.

	// Set up the version/runtime/debug-related variables, and cache them.
	// `versionInfo` is a global which has very likely been already initialised.
	if versionInfo, err = initVersionInfo(); err != nil {
		panic(fmt.Sprintf("Failed to initialize version info: %v\n", err))
	}

	// Check if we have the API key on environment.
	// Note that we are using godotenv/autoload to automatically retrieve .env
	// and merge with the existing environment.
	setting.Key = os.Getenv("TINIFY_API_KEY")

	startLevel := zerolog.ErrorLevel
	// Debug override (for testing purposes): check environmnt, change debug level if present.
	if debugOverride := os.Getenv("TINIFY_API_DEBUG"); len(debugOverride) > 0 {
		if startLevel, err = zerolog.ParseLevel(debugOverride); err != nil {
			// when zerolog.ParseLevel() fails, it returns zerolog.NoLevel; we'll go back to default (zerolog.ErrorLevel).
			startLevel = zerolog.ErrorLevel
		}
	}

	// testing zerolog:
	setting.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.DateTime,
		PartsOrder: []string{
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
			zerolog.CallerFieldName,
		},
		FormatCaller: func(i any) string {
			return "(" + filepath.Base(fmt.Sprintf("%s", i)) + ")"
		},
	}).
		Level(startLevel).
		With().
		Caller().
		Logger()

	// while testing, temporarily override everything and put the loggr in trace mode.
	// tinifyLoggingLevel := zerolog.TraceLevel
	// setting.LoggingLevel = zerolog.LevelTraceValue

	// Note that the zerolog setting.Logger is *always* returned; if it cannot write to the log
	// for some reason, that error will be handled by the zerolog passage, thus
	// the simple `Debug()` call here: if this _fails_, we've not done anything yet with
	// the images, and can safely abort.
	setting.Logger.Debug().Msgf("setting.Logger started at logging level %q; tinify pkg version %s",
		setting.Logger.GetLevel(),
		Tinify.VERSION)

	// check for terminal width if we're on a TTY
	setting.TerminalWidth = 80
	if term.IsTerminal(int(os.Stdin.Fd())) {
		setting.Logger.Debug().Msgf("TTY detected on stdin")
		width, _, err := term.GetSize(int(os.Stdin.Fd()))
		if err == nil {
			setting.TerminalWidth = width
		} else {
			setting.Logger.Debug().Msgf("could not get the size of the TTY: %s", err)
		}
	}

	setting.Logger.Debug().Msgf("Terminal width set to %d", setting.TerminalWidth)

	// Contains information about the compiled code in a format that urfave/cli likes.
	metadata := map[string]any{
		"Version":      versionInfo.version,
		"Commit":       versionInfo.commit,
		"Date":         versionInfo.dateString,
		"Built by":     versionInfo.builtBy,
		"OS":           versionInfo.goOS,
		"Architecture": versionInfo.goARCH,
		"Go version":   versionInfo.goVersion,
	}

	// These are the special arguments for input and output. They're here to avoid constant repetition.
	var inputOutputFilenames = []cli.Argument{
		&cli.StringArg{
			Name:        "input file",
			UsageText:   "input `filename` (use '-' for STDIN)",
			Destination: &setting.ImageName,
		},
		&cli.StringArg{
			Name:        "output file",
			UsageText:   "output `filename` (use - for STDOUT)",
			Destination: &setting.OutputFileName,
		},
	}

	// these are flags

	// start CLI app
	cmd := &cli.Command{
		Name: filepath.Base(os.Args[0]),
		Usage: justify.Justify("Calls the Tinify API from TinyPNG "+func() string {
			if len(setting.Key) < 5 {
				return "(environment variable TINIFY_API_KEY not set or invalid key)"
			}
			return "(with key [..." + setting.Key[len(setting.Key)-4:] + "])"
		}(), setting.TerminalWidth),
		UsageText:             justify.Justify(os.Args[0]+" [COMMAND] [OPTIONS] [INPUT FILE] [OUTPUT FILE]\nWith no INPUT FILE, or when INPUT FILE is -, read from standard input.", setting.TerminalWidth),
		Version:               fmt.Sprint(versionInfo),
		DefaultCommand:        "compress",
		EnableShellCompletion: true,
		Suggest:               true, // see https://cli.urfave.org/v3/examples/help/suggestions/
		Metadata:              metadata,
		Authors: []any{
			&mail.Address{Name: "gwpp", Address: "ganwenpeng1993@163.com"},
			&mail.Address{Name: "Gwyneth Llewelyn", Address: "gwyneth.llewelyn@gwynethllewelyn.net"},
		},
		Copyright: justify.Justify(fmt.Sprintf("© 2017-%d by Ganwen Peng. All rights reserved. Freely distributed under an MIT license.\nThe authors are neither affiliated nor endorsed by Tinify B.V.", time.Now().Year()), setting.TerminalWidth),
		Arguments: inputOutputFilenames,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "input",
				Aliases:     []string{"i"},
				Usage:       "input `filename` (empty or '-' for STDIN)",
				Destination: &setting.ImageName,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Usage:       "output `filename` (empty or '-' for STDOUT)",
				Destination: &setting.OutputFileName,
			},
			&cli.StringFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "debug `level`; \"error\" means no logging",
				Value:       zerolog.LevelErrorValue,
				Destination: &setting.LoggingLevel,
				Action: func(ctx context.Context, c *cli.Command, s string) error {
					// Check if the debug level is valid: it must be one of the zerolog valid types.
					// NOTE: This will be set later on anyway...
					// NOTE: setting.LoggingLevel will *always* be set to the default value, no matter what.
					setting.Logger.Debug().Msgf("cli.StringFlag(): setting logging level to... %q", setting.LoggingLevel)
					return setLogLevel()
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "compress",
				Aliases:   []string{"comp"},
				Usage:     "compresses and optimises an image",
				UsageText: justify.Justify("You can upload any image to the Tinify API to compress it. We will automatically detect the type of image ("+strings.Join(types, ", ")+") and optimise with the TinyPNG or TinyJPG engine accordingly.\nCompression will start as soon as you upload a file or provide the URL to the image.", setting.TerminalWidth),
				Action:    compress,
				Arguments: inputOutputFilenames,
			},
			{
				Name:      "resize",
				Aliases:   []string{"r"},
				Usage:     "resizes the image to a new size, using one of the possible methods",
				UsageText: justify.Justify("Use the API to create resized versions of your uploaded images.\nBy letting the API handle resizing you avoid having to write such code yourself and you will only have to upload your image once. The resized images will be optimally compressed with a nice and crisp appearance.\nYou can also take advantage of intelligent cropping to create thumbnails that focus on the most visually important areas of your image.\nResizing counts as one additional compression. For example, if you upload a single image and retrieve the optimized version plus 2 resized versions this will count as 3 compressions in total.\nAvailable compression methods are: "+strings.Join(methods, ", "), setting.TerminalWidth),
				Action:    resize,
				Arguments: inputOutputFilenames,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "method",
						Aliases:     []string{"m"},
						Value:       Tinify.ResizeMethodScale,
						Usage:       "resizing method [" + strings.Join(methods, ", ") + "]",
						Destination: &setting.Method,
						Action: func(ctx context.Context, c *cli.Command, s string) error {
							// Check if the resizing method is a valid one.
							// First check if it's empty:
							if len(setting.Method) == 0 {
								setting.Method = Tinify.ResizeMethodScale // scale is default
							} else if !slices.Contains(methods, setting.Method) {
								// Checked if it's one of the valid methods; if not, abort.
								return fmt.Errorf("invalid resize method: %q", setting.Method)
							}
							return nil
						},
					},
					&cli.Int64Flag{
						Name:        "width",
						Aliases:     []string{"w"},
						Value:       0,
						Usage:       "destination image `width`",
						Destination: &setting.Width,
					},
					&cli.Int64Flag{
						Name:        "height",
						Aliases:     []string{"g"},
						Value:       0,
						Usage:       "destination image `height`",
						Destination: &setting.Height,
					},
				},
			},
			{
				Name:      "convert",
				Aliases:   []string{"conv"},
				Usage:     "converts from one file type to another (" + strings.Join(types, ", ") + " supported)",
				UsageText: justify.Justify("You can use the API to convert your images to your desired image type.\nTinify currently supports converting between: "+strings.Join(types, ", ")+".\nWhen you provide more than on image type in your convert request, the smallest version will be returned to you.\nImage converting will count as one additional compression.", setting.TerminalWidth),
				Action:    convert,
				Arguments: inputOutputFilenames,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "type",
						Aliases:     []string{"t"},
						Usage:       "file type [" + strings.Join(types, ", ") + "]",
						Value:       "webp",
						Destination: &setting.FileType,
						Action: func(ctx context.Context, c *cli.Command, s string) error {
							// Check if the type(s) are all valid:
							if setting.FileType != "" {
								typesFound := strings.Split(setting.FileType, ",")
								if typesFound == nil {
									return fmt.Errorf("convert: no valid file types found")
								}
								// A very inefficient way of checking if all file types are valid O(n).
								// TODO(gwyneth): See if there is already a library function for this,
								// or use a different, linear approach.
								for _, aFoundType := range typesFound {
									if !slices.Contains(types, aFoundType) {
										return fmt.Errorf("convert: invalid file format: %q", aFoundType)
									}
								}
								// if we're here, all file types are valid
								setting.Logger.Debug().Msg("convert: all file type parameters are valid")
							} else {
								setting.Logger.Debug().Msg("convert: no file type parameters found, trying to guess")
							}
							return nil
						},
					},
				},
			},
			{
				Name:      "transform",
				Aliases:   []string{"tr"},
				Usage:     "processes image further (currently only replaces the background with a solid colour)",
				UsageText: justify.Justify("If you wish to convert an image with a transparent background to one with a solid background, specify a background property in the transform object.\nIf this property is provided, the background of a transparent image will be filled (only \"white\", \"black\", or a hex value are allowed).", setting.TerminalWidth),
				Action:    transform,
				Arguments: inputOutputFilenames,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "background",
						Aliases:     []string{"bg"},
						Value:       "",
						Usage:       "only \"white\", \"black\", or a hex `value` are allowed",
						Destination: &setting.Transform,
						Action: func(ctx context.Context, c *cli.Command, s string) error {
							// Check if value passed is correct.
							setting.Transform = strings.ToLower(setting.Transform)
							if setting.Transform == "white" || setting.Transform == "black" {
								return nil
							}
							// Just check if the rmaining string is a valid hex string.
							// (gwyneth 20250713)
							if !isValidHex(setting.Transform) {
								return fmt.Errorf("background colour: invalid hex value %q", setting.Transform)
							}
							return nil
						},
					},
				},
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "show version and compilation data",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Printf("VERSION:\n%s\n", versionInfo)
					return nil
				},
			},
		},
		CommandNotFound: func(ctx context.Context, cmd *cli.Command, command string) {
			setting.Logger.Fatal().Msgf("Command %q not found.\nUsage: %s", command, cmd.UsageText)
		},
		OnUsageError: func(ctx context.Context, cmd *cli.Command, err error, isSubcommand bool) error {
			if isSubcommand {
				return err
			}

			setting.Logger.Error().Msgf("Wrong usage: %#v", err)
			return nil
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// Setup phase
			setting.Logger.Debug().Msgf("`Before` action inside loop: before calling setLogLevel(), logging level was: %q(%d)",
				setting.LoggingLevel,
				setting.Logger.GetLevel())

			// force new debugging level, if it was set (gwyneth 20251007)
			// NOTE: we can safely ignore the error here.
			setLogLevel()

			setting.Logger.Debug().Msgf("`Before` action inside loop: after calling setLogLevel(), log level is now set to: %q(%d)",
				setting.LoggingLevel,
				setting.Logger.GetLevel())

			// Check if key is somewhat valid, i.e. has a decent amount of chars:
			if len(setting.Key) < 5 {
				return ctx, fmt.Errorf("invalid Tinify API key %q; too short — please check your key and try again", setting.Key)
			}

			// Now safely set the API key
			Tinify.SetKey(setting.Key)
			setting.Logger.Debug().Msgf("`Before` action inside loop: a Tinify API key was found: [...%s]", setting.Key[len(setting.Key)-4:])

			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Everything not defined above happens here!

			setting.Logger.Debug().Msg("Reached empty Action block")
			return nil
		},
	}

	setting.Logger.Debug().Msgf("Log level before outside loop: %q(%d)",
		setting.LoggingLevel,
		setting.Logger.GetLevel(),
	)

	if err := cmd.Run(ctx, os.Args); err != nil {
		// setting.Logger.Fatal().Err(err)
		setting.Logger.Fatal().Msg(err.Error())
	}
} // main

// openStream attempts to open a file, stdin, or a URL, and passes the image along for
// processing by the API.
func openStream(ctx context.Context) (context.Context, *Tinify.Source, error) {
	// Input file may be either an image filename or an URL; TinyPNG will handle both.
	// Since `://` is hardly a valid filename, but a requirement for being an URL,
	// handle URL later.
	// Note that if setting.ImageName is unset, stdin is assumed, even if it might not yet work.

	var (
		err      error      // declared here due to scope issues.
		f        = os.Stdin // file handler; STDIN by default.
		rawImage []byte     // raw image file, when loaded from disk.
		source   *Tinify.Source
	)

	// Tweak command-line to deal with files as parameters without --input etc.

	// Now check for the special "-" which also denotes STDIN:
	if setting.ImageName == "-" {
		setting.ImageName = ""
		setting.Logger.Trace().Msg("openStream: input file was \"-\", setting to STDIN")
	}
	// At this stage, the ImageName can be an URL, which we cannot check; it's the
	// Tinify API that will deal with checking if it's valid or not.
	if strings.Contains(setting.ImageName, "://") {
		setting.Logger.Trace().Msgf("openStream: input file %q is an URL, can't check", setting.ImageName)
	} else {
		// Last chance: does this file exist on the local filesystem?
		if _, err := os.Stat(setting.ImageName); err != nil {
			setting.Logger.Error().Msgf("openStream: input file %q not found; reverting to STDIN", setting.ImageName)
			setting.ImageName = ""
		}
	}
	// At this point, the input is either a valid filename, an (unknown) URL, or STDIN.
	setting.Logger.Trace().Msgf("openStream: input file is %q", setting.ImageName)

	// Now do the same for the output filename, if it exists.
	// The code is similar to the above.
	// Check for the special "-" which in this case denotes STDOUT:
	if setting.OutputFileName == "-" {
		setting.OutputFileName = ""
		setting.Logger.Trace().Msg("openStream: output file was \"-\", setting to STDOUT")
	}
	// At this stage, either we write to STDOUT or to a file (which we don't know if we can do so or not).
	// We can check if the directory, at least, has write permissions, and fail otherwise, to avoid wasting
	// one token call in case of a mistake:
	if setting.OutputFileName != "" {
		writeDir := filepath.Dir(setting.OutputFileName)
		if b, dErr := isDirWritable(writeDir); !b {
			return ctx, nil, fmt.Errorf("cannot save to %q, error was: %q", setting.OutputFileName, dErr)
		}
	}
	// At this point, the output is either a valid filename or STDOUT. (see above)
	setting.Logger.Trace().Msgf("openStream: output file is %q", setting.OutputFileName)

	// And now we have already dealt with parsing potential arguments.
	// So, either setting.ImageName is a valid filename, or an URL, or empty (= STDIN).
	setting.Logger.Debug().Msgf("openStream: opening input file for reading: %q", setting.ImageName)
	if setting.ImageName == "" || !strings.Contains(setting.ImageName, "://") {
		if setting.ImageName == "" {
			// are we on a TTY, or getting content from a pipe?
			if term.IsTerminal(int(f.Fd())) {
				return ctx, nil, fmt.Errorf("openStream: cannot read interactively from a TTY; use --input or pipe a file to STDIN")
			}

			// Logging to console, so let the user knows that as well
			setting.Logger.Info().Msg("openStream: empty input filename; reading from console/stdin instead")
		} else {
			// check to see if we can open this file:
			f, err = os.Open(setting.ImageName)
			if err != nil {
				return ctx, nil, err
			}
			setting.Logger.Debug().Msgf("openStream: %q sucessfully opened", setting.ImageName)
		}
		// Get the image file from disk/stdin.
		rawImage, err = io.ReadAll(f)
		if err != nil {
			return ctx, nil, err
		}
		// check canonical mime type:
		mimeType := http.DetectContentType(rawImage)
		// Valid Media Types according to IANA (se https://www.iana.org/assignments/media-types/media-types.xhtml#image)
		switch mimeType {
		case "image/png", "image/apng", "image/vnd.mozilla.apng", "image/vnd.sealed.png", "image/jpeg", "image/webp", "image/avif":
			setting.Logger.Trace().Msgf("openStream: setting Media Type to (valid) %q", mimeType)
		default:
			return ctx, nil, fmt.Errorf("openStream: invalid or not recognised Media Type %q, aborting", mimeType)
		}

		setting.Logger.Debug().Msgf("openStream: arg: %q (empty means stdin), size %d, Media Type %q", setting.ImageName, len(rawImage), mimeType)

		// Now call the TinyPNG API
		source, err = Tinify.FromBuffer(rawImage)
		if err != nil {
			return ctx, nil, err
		}
	} else {
		// we're assuming that we've got a valid URL, which might *not* be the case!
		// TODO(Tasker): extra validation
		source, err = Tinify.FromUrl(setting.ImageName)
		if err != nil {
			return ctx, nil, err
		}
	}
	return ctx, source, nil
}

// All-purpose API call. Whatever is done, it happens on the globals.
func callAPI(_ context.Context, cmd *cli.Command, source *Tinify.Source) error {
	var (
		err      error  // declared here due to scope issues.
		rawImage []byte // raw image file, when loaded from disk.
	)

	if len(cmd.Name) == 0 {
		return fmt.Errorf("no command")
	}
	setting.Logger.Debug().Msgf("inside callAPI(), invoked by %q", cmd.Name)

	// If we have no explicit output filename, write directly to stdout.
	if len(setting.OutputFileName) == 0 {
		setting.Logger.Debug().Msg("callAPI: no output filename; writing to stdout instead")
		// Warning: `source` is a global variable in this context!.
		rawImage, setting.CompressionCount, err = source.ToBufferC()
		if err != nil {
			setting.Logger.Error().Err(err)
			return err
		}
		// rawImage contains the raw image data; we push it out to STDOUT.
		n, err := os.Stdout.Write(rawImage)
		if err != nil {
			setting.Logger.Error().Err(err)
			return err
		}

		setting.Logger.Debug().Msgf("callAPI: wrote %d byte(s) to stdout; compression count: %d", n, setting.CompressionCount)
		return nil
	}

	setting.Logger.Debug().Msgf("callAPI: opening file %q for outputting image", setting.OutputFileName)

	// write to file, we have a special function for that already defined:
	setting.CompressionCount, err = source.ToFileC(setting.OutputFileName)
	if err != nil {
		setting.Logger.Error().Err(err)
		return err
	}

	setting.Logger.Info().Msgf("Succesfully wrote to %q, compression count: %d", setting.OutputFileName, setting.CompressionCount)
	return nil
}

// Tries to get a list of types to convert to, and calls the API.
func convert(ctx context.Context, cmd *cli.Command) error {
	var (
		err    error // declared here due to scope issues.
		source *Tinify.Source
	)

	setting.Logger.Debug().Msgf("convert called, conversion type request was %q", setting.FileType)

	// If setting.FileType is empty, do a complex lookup to figure out what we ought to convert it to.
	// If all else fails, use WebP as the default (e.g., when stdout is selected, as there is no way to guess).
	if setting.FileType == "" {
		// check extension
		if setting.OutputFileName != "" {
			extension := strings.ToLower(filepath.Ext(setting.OutputFileName))
			if len(extension) > 1 { // skip dot before extension!
				// check if it's one that the Tinify API recognises as valid:
				if !slices.Contains(types, extension[1:]) {
					setting.FileType = extension[1:]
					setting.Logger.Trace().Msgf("convert: conversion type request was empty, now set to valid file extension %q", setting.FileType)
				}
			}
		}
	}
	// this could be written differently, avoiding the comparison, but this code
	// is cleaner: if we didn't manage to set it earlier, we set it now to the default, webp:
	if setting.FileType == "" {
		setting.FileType = "webp"
		setting.Logger.Trace().Msg("convert: conversion type request was not set, using \"webp\" as default")
	}

	if ctx, source, err = openStream(ctx); err != nil {
		setting.Logger.Error().Msgf("convert: invalid filenames, error was %q", err)
		return err
	}

	// user can request conversion to multiple file types, comma-separated; we need to split
	// these since our Convert logic presumes maps of strings, to properly JSONificta them,
	if err := source.Convert(strings.Split(strings.ToLower(setting.FileType), ",")); err != nil {
		return err
	}
	// again, note that `source` is a global.
	return callAPI(ctx, cmd, source)
}

// Resizes image, given a width and a height.
func resize(ctx context.Context, cmd *cli.Command) error {
	var (
		err    error // declared here due to scope issues.
		source *Tinify.Source
	)
	setting.Logger.Debug().Msgf("resize called; debug is %q, method is %q, width is %d px, height is %d px",
		setting.LoggingLevel, setting.Method, setting.Width, setting.Height)

	// width and height are globals.
	if setting.Width == 0 && setting.Height == 0 {
		setting.Logger.Error().Msg("resize: width and height cannot be simultaneously zero")
		return fmt.Errorf("resize: width and height cannot be simultaneously zero")
	}

	setting.Logger.Debug().Msg("resize: now calling openStream()")

	if ctx, source, err = openStream(ctx); err != nil {
		setting.Logger.Error().Msgf("resize: invalid filenames, error was %v", err)
		return err
	}

	setting.Logger.Debug().Msg("resize: now calling source.Resize()")

	// method is a global too.
	err = source.Resize(&Tinify.ResizeOption{
		Method: Tinify.ResizeMethod(setting.Method),
		Width:  setting.Width, // replace by real value!
		Height: setting.Height,
	})

	if err != nil {
		setting.Logger.Error().Err(err)
		return err
	}

	return callAPI(ctx, cmd, source)
}

// Compress is the default. It handles the special case when argument 0 (the command)
// has been omitted.
func compress(ctx context.Context, cmd *cli.Command) error {
	var (
		err    error // declared here due to scope issues.
		source *Tinify.Source
	)
	// Was this called as the default (i.e. empty) command? If so, adjust for 1 or 2 arguments.
	if len(os.Args) > 1 && os.Args[1] != "compress" && fs.ValidPath(os.Args[1]) {
		setting.Logger.Trace().Msgf("compress called by default, command was omitted; extracting filename(s) directly from arguments: %#v", os.Args)
		setting.ImageName = os.Args[1]
		if len(os.Args) > 2 && fs.ValidPath(os.Args[2]) {
			setting.OutputFileName = os.Args[2]
		}
	}

	setting.Logger.Debug().Msgf("compress called for %q -> %q", setting.ImageName, setting.OutputFileName)

	if ctx, source, err = openStream(ctx); err != nil {
		setting.Logger.Fatal().Msgf("compress: invalid filenames, error was: %v", err)
		// probably never reached:
		return err
	}

	return callAPI(ctx, cmd, source)
}

// Transform allows o remove the background (that's the only option in the Tinify API so far).
func transform(ctx context.Context, cmd *cli.Command) error {
	var (
		err    error // declared here due to scope issues.
		source *Tinify.Source
	)

	setting.Logger.Debug().Msg("transform called")
	if len(setting.Transform) == 0 {
		return fmt.Errorf("transform: empty transformation type passed")
	}

	if ctx, source, err = openStream(ctx); err != nil {
		setting.Logger.Error().Msgf("transform: invalid filenames, error was %v", err)
		return err
	}

	if err = source.Transform(&Tinify.TransformOptions{
		Background: setting.Transform,
	}); err != nil {
		return err
	}
	return callAPI(ctx, cmd, source)
}

// Aux functions

// setLogLevel is just a macro-style thing to force the logging level to be set.
func setLogLevel() error {
	setting.Logger.Debug().Msgf("setLogLevel(): log level to be set: %q", setting.LoggingLevel)
	if len(setting.LoggingLevel) > 0 {
		if tinifyLoggingLevel, err := zerolog.ParseLevel(setting.LoggingLevel); err == nil {
			// Ok, valid error level selected, set it:
			setting.Logger = setting.Logger.Level(tinifyLoggingLevel)
			setting.LoggingLevel = tinifyLoggingLevel.String()
			setting.Logger.Debug().Msgf("setLogLevel(): successfully set logging level to %q (%d)",
				setting.LoggingLevel,
				setting.Logger.GetLevel(),
			)
			return nil
		} else {
			setting.Logger.Debug().Msgf("setLogLevel(): error parsing logging level %q: %s",
				setting.LoggingLevel, err)
		}
	}
	// Unknown logging level, or empty logging level, so fall back to "error"
	setting.Logger.Debug().Msgf("setLogLevel(): empty or unknown logging level %q, falling back to \"error\"",
		setting.LoggingLevel)
	setting.Logger = setting.Logger.Level(zerolog.ErrorLevel)
	setting.LoggingLevel = zerolog.ErrorLevel.String()

	return fmt.Errorf("unknown logging type %q, setting to \"error\" (%d) by default",
		setting.LoggingLevel, setting.Logger.GetLevel())
}

// check if this is a valid Hex value for a colour or not.
// allow CSS RGB types of colours, with or without #
// 3 digits, 6 digits, or 8 digits (for transpareny) accepted.
// See https://stackoverflow.com/a/79589454/1035977
// (gwyneth 20250713)
func isValidHex(s string) bool {
	n := len(s)
	// empty string or string with just a '#'?
	if n == 0 {
		return false
	}

	start := 0
	if s[0] == '#' {
		n--
		start = 1
	}

	// check for valid ranges
	if n != 4 && n != 6 && n != 8 {
		return false
	}

	// must be "#xxx" or "#xxxxxx"
	// check each hex digit
	for i := start; i < n; i++ {
		b := s[i] | 0x20             // fold A–F into a–f, digits unaffected
		if b-'0' < 10 || b-'a' < 6 { // '0'–'9' or 'a'–'f' ?
			continue
		}
		return false
	}
	return true
}

// A stupid way (which works) to check if a directory is writable by this user or not.
// The advantage is that it works universally on all operating systems.
func isDirWritable(path string) (bool, error) {
	tmpFile := "tmpfile"

	file, err := os.CreateTemp(path, tmpFile)
	if err != nil {
		return false, err
	}

	defer os.Remove(file.Name())
	defer file.Close()

	return true, nil
}

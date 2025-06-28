package main

import (
	"context"
	"crypto/ecdh"
	"fmt"
	"io"
	"net/mail"
	"os"
	"slices"
	"strings"
	"time"

	Tinify "github.com/gwpp/tinify-go/tinify"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

// No harm is done having just one context, which is simoly the background.
var ctx = context.Background()

// Type to hold the global variables for all possible calls.
type Setting struct {
	DebugLevel     int8           `json:"debug_level"`      // Debug/verbosity level, 0 is no debugging, > 1 info only and above.
	ImageName      string         `json:"image_name"`       // Filename or URL.
	OutputFileName string         `json:"output_file_name"` // If set, it's the output filename; if not, well...
	FileType       string         `json:"file_type"`        // Any set of webp, png, jpg, avif.
	Key            string         `json:"key"`              // TinyPNG API key; can be on environment or read from .env.
	Logger         zerolog.Logger `json:"-"`                // The main setting.Logger. Probably not necessary.
	Method         string         `json:"method"`           // Resizing method (scale, fit, cover, thumb).
	Width          int64          `json:"width"`            // Image width  (for resize operations).
	Height         int64          `json:"height"`           // Image height (  "   "      "    "  ).
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

// Main starts here.
func main() {
	// Set up the version/runtime/debug-related variables, and cache them.
	if err := initVersionInfo(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize version info: %v\n", err)
	}

	// Check if we have the API key on environment.
	// Note that we are using godotenv/autoload to automatically retrieve .env
	// and merge with the existing environment.
	setting.Key = os.Getenv("TINIFY_API_KEY")

	// Grab flags

	// Start zerolog setting.Logger.
	// We're also using it for pretty-printing to the console.
	setting.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.Level(setting.DebugLevel)). // typecast from int8 to zerolog.Level
		With().
		Timestamp().
		Caller().
		//		Int("pid", os.Getpid()).
		//		Str("go_version", buildInfo.oVersion).
		Logger()

	// Note that the zerolog setting.Logger is *always* returned; if it cannot write to the log
	// for some reason, that error will be handled by the zerolog passage, thus
	// the simple `Debug()` call here: if this _fails_, we've not done anything yet with
	// the images, and can safely abort.
	setting.Logger.Debug().Msgf("setting.Logger started; tinify pkg version %s", Tinify.VERSION)

	// 4. Check if the type(s) are all valid:
	if setting.FileType != "" {
		typesFound := strings.Split(setting.FileType, ",")
		if typesFound == nil {
			setting.Logger.Fatal().Msg("no valid file types found")
		}
		// A very inefficient way of checking if all file types are valid O(n).
		// TODO(gwyneth): See if there is already a library function for this,
		// or use a different, linear approach.
		for _, aFoundType := range typesFound {
			if !slices.Contains(types, aFoundType) {
				setting.Logger.Fatal().Msgf("invalid file format: %s", aFoundType)
				os.Exit(3)
			}
		}
		// if we're here, all file types are valid
		setting.Logger.Debug().Msg("all file type parameters are valid")
	} else {
		setting.Logger.Debug().Msg("no file type parameters found")
	}

	// 5. Check if the resizing method is a valid one.
	// First check if it's empty:
	if len(setting.Method) == 0 {
		setting.Method = Tinify.ResizeMethodScale // scale is default
	} else if !slices.Contains(methods, setting.Method) {
		// Checked if it's one of the valid methods; if not, abort.
		setting.Logger.Fatal().Msgf("invalid resize method: %s", setting.Method)
		os.Exit(3)
	}

	// Input file may be either an image filename or an URL; TinyPNG will handle both-
	// Since `://` is hardly a valid filename, but a requirement for being an URL;
	// handle URL later.
	// Note that if setting.ImageName is unset, stdin is assumed, even if it might not yet work.
	setting.Logger.Debug().Msgf("opening input file for reading: %q", setting.ImageName)
	if setting.ImageName == "" || !strings.Contains(setting.ImageName, "://") {
		if setting.ImageName == "" {
			// empty filename; use stdin
			f = os.Stdin
			// Logging to console, so let the user knows that as well
			setting.Logger.Info().Msg("empty filename; reading from console/stdin instead")
		} else {
			// check to see if we can open this file:
			f, err = os.Open(setting.ImageName)
			if err != nil {
				setting.Logger.Fatal().Err(err)
			}
			setting.Logger.Debug().Msgf("%q sucessfully opened", setting.ImageName)
		}
		// Get the image file from disk/stdin.
		rawImage, err = io.ReadAll(f)
		if err != nil {
			setting.Logger.Fatal().Err(err)
		}

		setting.Logger.Debug().Msgf("arg: %q (empty means stdin), size %d", setting.ImageName, len(rawImage))

		// Now call the TinyPNG API
		source, err = Tinify.FromBuffer(rawImage)
		if err != nil {
			setting.Logger.Fatal().Err(err)
		}
	} else {
		// we're assuming that we've got a valid URL, which might *not* be the case!
		// TODO(Tasker): extra validation
		source, err = Tinify.FromUrl(setting.ImageName)
		if err != nil {
			setting.Logger.Fatal().Err(err)
		}
	}

	// start CLI app
	cmd := &cli.Command{
		Name:      "tinify-go",
		Usage:     "Calls the Tinify API from TinyPNG. Make sure you have TINIFY_API_KEY set!",
		UsageText: os.Args[0] + " [OPTION] [FLAGS] [INPUT FILE] [OUTPUT FILE]\nWith no INPUT FILE, or when INPUT FILE is -, read from standard input.",
		Version: fmt.Sprintf(
			"%s (rev %s)\n[%s %s %s]\n[build at %s by %s]",
			versionInfo.version,
			versionInfo.commit,
			versionInfo.goOS,
			versionInfo.goARCH,
			versionInfo.goVersion,
			versionInfo.dateString, // Date as string in RFC3339 notation.
			versionInfo.builtBy,    // `go build -ldflags "-X main.TheBuilder=[insertname here]"`
		),
		EnableShellCompletion: true,
		//		Compiled: versionInfo.date,		// Converted from RFC333
		Authors: []any{
			&mail.Address{Name: "gwpp", Address: "ganwenpeng1993@163.com"},
			&mail.Address{Name: "Gwyneth Llewelyn", Address: "gwyneth.llewelyn@gwynethllewelyn.net"},
		},
		Copyright: fmt.Sprintf("© 2017-%d by Ganwen Peng. All rights reserved. Freely distributed under an MIT license.", time.Now().Year()),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "input",
				Aliases:     []string{"i"},
				Usage:       "input filename (empty=STDIN)",
				Destination: &setting.ImageName,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Usage:       "output filename (empty=STDOUT)",
				Destination: &setting.OutputFileName,
			},
			&cli.StringFlag{
				Name:        "type",
				Aliases:     []string{"t"},
				Usage:       "file type [" + strings.Join(types, "    , ") + "]",
				DefaultText: "webp",
				Destination: &setting.FileType,
			},
			&cli.Int8Flag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "debug level; zero means no debug",
				Value:       4,
				Destination: &setting.DebugLevel,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "compress",
				Aliases: []string{"comp"},
				Usage:   "You can upload any image to the Tinify API to com press it. We will automatically detect the type of image (" + strings.Join(types, ", ") + ") and optimise with the TinyPNG or TinyJPG engine accordingly.\nCompression will start as soon as you upload a file or provide the URL to the image.",
				Action:  compress,
			},
			{
				Name:    "resize",
				Aliases: []string{"r"},
				Usage:   "Use the API to create resized versions of your uploaded images. By letting the API handle resizing you avoid having to write such code yourself and you will only have to upload your image once. The resized images will be optimally compressed with a nice and crisp appearance.\nYou can also take advantage of intelligent cropping to create thumbnails that focus on the most visually important areas of your image.\nResizing counts as one additional compression. For example, if you upload a single image and retrieve the optimized version plus 2 resized versions this will count as 3 compressions in total.\nAvailable compression methods are: " + strings.Join(methods, ", "),
				Action:  resize,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "input",
						Aliases: []string{"i"},
						Usage:   "input filename (empty=STDIN)",
					},
				},
				/*
					flag.StringVarP(&setting.Method, "method", "m", Tinify.ResizeMethodScale, "resizing method ["+strings.Join(methods, ", ")+"]")
					flag.Int64VarP(&setting.Width, "width", "w", 0, "destination image width")
					flag.Int64VarP(&setting.Height, "height", "g", 0, "destination image height")

				*/
			},
			{
				Name:    "convert",
				Aliases: []string{"conv"},
				Usage:   "You can use the API to convert your images to your desired image type. Tinify currently supports converting between: " + strings.Join(types, ",") + ".\n When you provide more than on image type i   n your convert request, the smallest version will be   returned to you.\nImage converting will count as one additional compression.",
				Action:  convert,
			},
			{
				Name:      "help",
				Aliases:   []string{"h"},
				Usage:     "Shows command help",
				ArgsUsage: "[command]",
				HideHelp:  false,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Everything happens here!

			// Check if key is somewhat valid, i.e. has a decent amount of chars:
			if len(setting.Key) < 5 {
				setting.Logger.Fatal().Msgf("invalid Tinify API %q; too short — please check your key and try again\n", setting.Key)
				os.Exit(2)
			}

			// Set the API key (we've already checked if it is valid & exists):
			Tinify.SetKey(setting.Key)
			setting.Logger.Debug().Msgf("a TinyPNG key was found: [...%s]", setting.Key[len(setting.Key)-4:])

			fmt.Println("to-do")
			return nil
		},
	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "Shows command help",
	}

	//	cli.CommandHelpTemplate = commandHelpTemplate

	if err := cmd.Run(ctx, os.Args); err != nil {
		setting.Logger.Fatal().Err(err)
	}
} // main

var (
	rawImage []byte         // raw image file, when loaded from disk.
	err      error          // declared here due to scope issues.
	f        = os.Stdin     // file handler; STDIN by default.
	source   *Tinify.Source // declared in advance to avoid scoping issues.
)

// All-purpose API call. Whatever is done, it happens on the globals.
func callAPI(ctx context.Context, cmd *cli.Command) error {
	setting.Logger.Debug().Msg("inside callAPI()")

	// If we have no explicit output filename, write directly to stdout
	if len(setting.OutputFileName) == 0 {
		setting.Logger.Debug().Msg("no output filename; writing to stdout instead")
		rawImage, err := source.ToBuffer()
		if err != nil {
			setting.Logger.Error().Err(err)
			return err
		}
		// rawImage contains the raw image data; we push it out to STDOUT
		n, err := os.Stdout.Write(rawImage)
		if err != nil {
			setting.Logger.Error().Err(err)
			return err
		}

		setting.Logger.Debug().Msgf("wrote %d byte(s) to stdout", n)
		return nil
	}

	setting.Logger.Debug().Msgf("opening file %q for outputting image", setting.OutputFileName)

	// write to file, we have a special function for that already defined:
	err = source.ToFile(setting.OutputFileName)
	if err != nil {
		setting.Logger.Error().Err(err)
		return err
	}

	return nil
}

// Tries to get a list
func convert(ctx context.Context, cmd *cli.Command) error {
	setting.Logger.Debug().Msg("convert called")
	return callAPI(ctx, cmd)
}

// Resizes image, given a width and a height.
func resize(ctx context.Context, cmd *cli.Command) error {
	// width and height are globals.
	setting.Logger.Debug().Msgf("resize called with width %d px and height %d px", setting.Width, setting.Height)
	if setting.Width == 0 && setting.Height == 0 {
		setting.Logger.Error().Msg("width and height cannot be simultaneously zero")
		return fmt.Errorf("width and height cannot be simultaneously zero")
	}

	setting.Logger.Debug().Msg("now calling source.Resize()")
	// method is a global too.
	err := source.Resize(&Tinify.ResizeOption{
		Method: Tinify.ResizeMethod(setting.Method),
		Width:  setting.Width, // replace by real value!
		Height: setting.Height,
	})

	if err != nil {
		setting.Logger.Error().Err(err)
		return err
	}

	return callAPI(ctx, cmd)
}

// Compress is the default.
func compress(ctx context.Context, cmd *cli.Command) error {
	setting.Logger.Debug().Msg("compress called")
	return callAPI(ctx, cmd)
}

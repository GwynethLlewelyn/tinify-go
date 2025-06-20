package main

import (
	"context"
	"fmt"
	"io"
	"net/mail"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gwpp/tinify-go/tinify"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

// Global variables for all possible calls.

var (
	debugLevel     int8           // Debug level, as set by the user; 0 = debug, > 1 info only and above.
	imageName      string         // Filename or URL.
	outputFileName string         // If set, it's the output filename; if not, well...
	fileType       string         // Any set of webp, png, jpg, avif.
	help           = false        // Fake flag-like construct to write usage etc.
	key            string         // TinyPNG API key; can be on environment or read from .env.
	logger         zerolog.Logger // The main logger. Probably not necessary.
	method         string         // Resizing method (scale, fit, cover, thumb).
	width          int64          // Image width  (for resize operations).
	height         int64          // Image height (  "   "      "    "  ).
)

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
	key = os.Getenv("TINIFY_API_KEY")

	// Grab flags
	flag.Int8VarP(&debugLevel, "debug", "d", 4, "debug level; non-zero means no debug")
	flag.StringVarP(&imageName, "input", "i", "", "input filename (empty=STDIN)")
	flag.StringVarP(&outputFileName, "output", "o", "", "output filename (empty=STDOUT)")
	flag.StringVarP(&fileType, "type", "t", "webp", "file type ["+strings.Join(types, ", ")+"]")
	flag.StringVarP(&method, "method", "m", Tinify.ResizeMethodScale, "resizing method ["+strings.Join(methods, ", ")+"]")
	flag.Int64VarP(&width, "width", "w", 0, "destination image width")
	flag.Int64VarP(&height, "height", "g", 0, "destination image height")
	flag.BoolVarP(&help, "help", "h", false, "show usage")

	flag.Parse()

	// Help message and default usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprint(os.Stderr, "Compresses/resizes/converts images using the Tinify API.\n"+
			"See https://tinypng.com/developers\n"+
			"This software is neither affiliated with, nor endorsed by Tinify B.V.\n",
			"Tinify Go Package version "+Tinify.VERSION+"\n")
		fmt.Fprintf(os.Stderr, "Built with %v\n", buildInfo.GoVersion)

		if len(key) < 5 {
			fmt.Fprintln(os.Stderr,
				"TINIFY_API_KEY not set in environment/command-line arguments, or key is invalid")
		} else {
			fmt.Fprintf(os.Stderr, "TINIFY_API_KEY found with last digits %q\n", key[len(key)-4:])
		}

		fmt.Fprintln(os.Stderr, "\nCOMMANDS:")
		for _, cmdHelp := range commands {
			fmt.Fprintf(os.Stderr, "- %s:\t%s\n\n", cmdHelp.Name, cmdHelp.Usage)
		}
		fmt.Fprintln(os.Stderr, "FLAGS:")
		flag.PrintDefaults()
	}

	// If no args, print help.
	if len(os.Args) < 1 {
		// this will not give us help for commands
		flag.Usage()
		return
	}

	// Start zerolog logger.
	// We're also using it for pretty-printing to the console.
	logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.Level(debugLevel)). // typecast from int8 to zerolog.Level
		With().
		Timestamp().
		Caller().
		//		Int("pid", os.Getpid()).
		//		Str("go_version", buildInfo.GoVersion).
		Logger()

	// Note that the zerolog logger is *always* returned; if it cannot write to the log
	// for some reason, that error will be handled by the zerolog passage, thus
	// the simple `Debug()` call here: if this _fails_, we've not done anything yet with
	// the images, and can safely abort.
	logger.Debug().Msgf("logger started; tinify pkg version %s", Tinify.VERSION)

	// Extract command from command line; it must be the first parameter.
	command := flag.Arg(0)

	// Do some basic error-checking on flag parameters.
	// TODO(gwyneth): Check if parameters are appropriate for this specific command,
	//   since not all commands take all parameters.

	// 1. Note that "help" is a special command as well.
	if help {
		command = "help"
	}

	// 2. Default command, if not found on the command line, is "compress".
	// This is to conform to TinyPNG specifications.
	if command == "" {
		command = "compress"
	}

	// 3. Every other command is/was garbage and needs to be flagged as error.
	if !checkCommand(command) {
		logger.Error().Msgf("invalid command %q", command)
		os.Exit(1)
	}
	// 4. Check if the type(s) are all valid:
	if fileType != "" {
		typesFound := strings.Split(fileType, ",")
		if typesFound == nil {
			logger.Fatal().Msg("no valid file types found")
		}
		// A very inefficient way of checking if all file types are valid O(n).
		// TODO(gwyneth): See if there is already a library function for this,
		// or use a different, linear approach.
		for _, aFoundType := range typesFound {
			if !slices.Contains(types, aFoundType) {
				logger.Fatal().Msgf("invalid file format: %s", aFoundType)
				os.Exit(3)
			}
		}
		// if we're here, all file types are valid
		logger.Debug().Msg("all file type parameters are valid")
	} else {
		logger.Debug().Msg("no file type parameters found")
	}

	// 5. Check if the resizing method is a valid one.
	// First check if it's empty:
	if len(method) == 0 {
		method = Tinify.ResizeMethodScale // scale is default
	} else if !slices.Contains(methods, method) {
		// Checked if it's one of the valid methods; if not, abort.
		logger.Fatal().Msgf("invalid resize method: %s", method)
		os.Exit(3)
	}

	// Prepare in advance some variables.

	// Last chance to get a valid API key! See if it was passed via flags (not recommended)
	if key == "" {
		flag.StringVarP(&key, "key", "k", "", "Tinify API key (ideally read from environment TINIFY_API_KEY)")
		if key == "" {
			// No key found anywhere, abort.
			logger.Fatal().Msg("the Tinify API key was not found anywhere (tried environment TINIFY_API_KEY and CLI flags); cannot proceed, aborting")
			// The best we can do at this stage is call help
			if execStatus := executeCommand("help"); execStatus != 0 {
				logger.Error().Msgf("help returned with error code %d", execStatus)
				os.Exit(execStatus)
			}
			os.Exit(2)
		}
	}

	// Check if key is somewhat valid, i.e. has a decent amount of chars:
	if len(key) < 5 {
		logger.Fatal().Msgf("invalid Tinify API %q; too short — please check your key and try again\n", key)
		os.Exit(2)
	}

	// Set the API key (we've already checked if it is valid & exists):
	Tinify.SetKey(key)
	logger.Debug().Msgf("a TinyPNG key was found: [...%s]", key[len(key)-4:])

	// Input file may be either an image filename or an URL; TinyPNG will handle both-
	// Since `://` is hardly a valid filename, but a requirement for being an URL;
	// handle URL later.
	// Note that if imageName is unset, stdin is assumed, even if it might not yet work.
	logger.Debug().Msgf("opening input file for reading: %q", imageName)
	if imageName == "" || !strings.Contains(imageName, "://") {
		if imageName == "" {
			// empty filename; use stdin
			f = os.Stdin
			// Logging to console, so let the user knows that as well
			logger.Info().Msg("empty filename; reading from console/stdin instead")
		} else {
			// check to see if we can open this file:
			f, err = os.Open(imageName)
			if err != nil {
				logger.Fatal().Err(err)
			}
			logger.Debug().Msgf("%q sucessfully opened", imageName)
		}
		// Get the image file from disk/stdin.
		rawImage, err = io.ReadAll(f)
		if err != nil {
			logger.Fatal().Err(err)
		}

		logger.Debug().Msgf("arg: %q (empty means stdin), size %d", imageName, len(rawImage))

		// Now call the TinyPNG API
		source, err = Tinify.FromBuffer(rawImage)
		if err != nil {
			logger.Fatal().Err(err)
		}
	} else {
		// we're assuming that we've got a valid URL, which might *not* be the case!
		// TODO(Tasker): extra validation
		source, err = Tinify.FromUrl(imageName)
		if err != nil {
			logger.Fatal().Err(err)
		}
	}

	// start app
	app := &cli.Command{
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
			&cli.BoolFlag{
				Name:        "binary",
				Aliases:     []string{"d"},
				Usage:       "read in binary mode (ignored)",
				Value:       false,
				Destination: &setting.BinaryOrText,
			},
			&cli.BoolFlag{
				Name:        "check",
				Aliases:     []string{"c"},
				Usage:       "read checksums from the FILEs and check them",
				Value:       false,
				Destination: &setting.Check,
			},
			&cli.BoolFlag{
				Name:        "tag",
				Usage:       "create a BSD-style checksum",
				Value:       false,
				Destination: &setting.Tag,
			},
			&cli.BoolFlag{
				Name:        "text",
				Aliases:     []string{"t"},
				Usage:       "read in text mode (ignored)",
				Value:       false,
				Destination: &setting.BinaryOrText,
			},
			&cli.BoolFlag{
				Name:        "zero",
				Aliases:     []string{"z"},
				Usage:       "end each output line with NUL, not newline, and disable file name escaping",
				Value:       false,
				Destination: &setting.Zero,
			},
			&cli.BoolFlag{
				Name:        "quiet",
				Aliases:     []string{"q"},
				Usage:       "don't print OK for each successfully verified file",
				Value:       false,
				Destination: &setting.Quiet,
				Action: func(ctx context.Context, cmd *cli.Command, flag bool) error {
					if !setting.Check {
						return cli.Exit(os.Args[0]+"the --quiet option is meaningful only when verifying checksums", 2)
					}
					return nil
				},
			},
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "shows additional debugging information",
				Value:       false,
				Destination: &setting.Debug,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "compress",
				Aliases: []string{"comp"},
				Usage:   "You can upload any image to the Tinify API to compress it. We will automatically detect the type of image (" + strings.Join(types, ", ") + ") and optimise with the TinyPNG or TinyJPG engine accordingly.\nCompression will start as soon as you upload a file or provide the URL to the image.",
				Action:  compress,
			},
			{
				Name:    "resize",
				Aliases: []string{"r"},
				Usage:   "Use the API to create resized versions of your uploaded images. By letting the API handle resizing you avoid having to write such code yourself and you will only have to upload your image once. The resized images will be optimally compressed with a nice and crisp appearance.\nYou can also take advantage of intelligent cropping to create thumbnails that focus on the most visually important areas of your image.\nResizing counts as one additional compression. For example, if you upload a single image and retrieve the optimized version plus 2 resized versions this will count as 3 compressions in total.\nAvailable compression methods are: " + strings.Join(methods, ", "),
				Action:  resize,
			},
			{
				Name:    "convert",
				Aliases: []string{"conv"},
				Usage:   "You can use the API to convert your images to your desired image type. Tinify currently supports converting between: " + strings.Join(types, ",") + ".\n When you provide more than one image type in your convert request, the smallest version will be returned to you.\nImage converting will count as one additional compression.",
				Action:  convert,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Everything happens here!

			fmt.Println("to-do")
		},
	}
} // main

var (
	rawImage []byte         // raw image file, when loaded from disk.
	err      error          // declared here due to scope issues.
	f        = os.Stdin     // file handler; STDIN by default.
	source   *Tinify.Source // declared in advance to avoid scoping issues.
)

// All-purpose API call. Whatever is done, it happens on the globals.
func callAPI() int {
	logger.Debug().Msg("inside callAPI()")

	// If we have no explicit output filename, write directly to stdout
	if len(outputFileName) == 0 {
		logger.Debug().Msg("no output filename; writing to stdout instead")
		rawImage, err := source.ToBuffer()
		if err != nil {
			logger.Error().Err(err)
			return 1
		}
		// rawImage contains the raw image data; we push it out to STDOUT
		n, err := os.Stdout.Write(rawImage)
		if err != nil {
			logger.Error().Err(err)
			return 1
		}

		logger.Debug().Msgf("wrote %d byte(s) to stdout", n)
		return 0
	}

	logger.Debug().Msgf("opening file %q for outputting image", outputFileName)

	// write to file, we have a special function for that already defined:
	err = source.ToFile(outputFileName)
	if err != nil {
		logger.Error().Err(err)
		return 1
	}

	return 0
}

// Tries to get a list
func convert() int {
	logger.Debug().Msg("convert called")
	return callAPI()
}

// Resizes image, given a width and a height.
func resize() int {
	// width and height are globals.
	logger.Debug().Msgf("resize called with width %d px and height %d px", width, height)
	if width == 0 && height == 0 {
		logger.Error().Msg("width and height cannot be simultaneously zero")
		return 2
	}

	logger.Debug().Msg("now calling source.Resize()")
	// method is a global too.
	err := source.Resize(&Tinify.ResizeOption{
		Method: Tinify.ResizeMethod(method),
		Width:  width, // replace by real value!
		Height: height,
	})

	if err != nil {
		logger.Error().Err(err)
		return 1
	}

	return callAPI()
}

// Compress is the default.
func compress() int {
	logger.Debug().Msg("compress called")
	return callAPI()
}

// Temporary function to display the usage of a specific command.
//
// Deprecated: usage will be displayed directly inside the urfave/cli object.
func helpUsage() int {
	explainCommand := flag.Arg(1)
	logger.Debug().Msgf("explaining command %q:", explainCommand)
	if _, ok := commands[explainCommand]; ok {
		fmt.Fprintf(os.Stderr, "- %s:\t%s\n", commands[explainCommand].Name, commands[explainCommand].Usage)
		return 0
	}
	// command unknown or empty, so print the usage.
	flag.Usage()
	return 1
}

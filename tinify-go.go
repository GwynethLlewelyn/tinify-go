package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/gwpp/tinify-go/tinify"
	"github.com/rs/zerolog"
	_ "github.com/joho/godotenv/autoload"
	flag "github.com/spf13/pflag"
)

// Global variables for all possible calls.

var (
	debugLevel int8			// Debug level, as set by the user; 0 = debug, > 1 info only and above.
	imageName string		// Filename or URL.
	outputFileName string	// If set, it's the output filename; if not, well...
	fileType string			// Any set of webp, png, jpg, avif.
	help = false			// Fake flag-like construct to write usage etc.
	key string				// TinyPNG API key; can be on environment or read from .env.
	logger zerolog.Logger	// The main logger. Probably not necessary.
	method string			// Resizing method (scale, fit, cover, thumb).
	width int64				// Image width  (for resize operations).
	height int64			// Image height (  "   "      "    "  ).
)

// Tinify API supported file types.
// Add more when TinyPNG supports additional types.
var types = []string {
	"png",
	"jpeg",
	"webp",
	"avif",
}

// Available image resizing methods.
// Add more when TinyPNG supports additional types.
var methods = []string {
	Tinify.ResizeMethodScale,
	Tinify.ResizeMethodFit,
	Tinify.ResizeMethodCover,
	Tinify.ResizeMethodThumb,
}

// Data interpreter structure.
// Minimalist scenario assuming just one command at the time, calling one action.
// Flags are globals, parsed as a side-effect.

// Struct to hold command data.
type Command struct {
	Name  string		// Command name.
	Usage string		// Usage/help for this command.
	Action func()int	// Function to call for this command.
}

// Type for all commands.
// Not strictly necessary, but useful later on.
type allCommands map[string]Command

// Map of all commands; global due to scoping issues.
// Further info for each command could be added here, e.g. aliases, etc.
var commands allCommands


// Define a function to execute a command.
func executeCommand(command string) int {
	// Check if command exists in map
	if _, ok := commands[command]; !ok {
		flag.Usage()
		return 1
	}

	// execute function for this command.
	return commands[command].Action()
}

//
// Main starts here.
//
func main() {
	// Init commands here, to avoid recursivity.
	commands = allCommands{
		"compress": {
			Name:	"compress",
			Usage:	"You can upload any image to the Tinify API to compress it. We will automatically detect the type of image (" + strings.Join(types, ", ") + ") and optimise with the TinyPNG or TinyJPG engine accordingly.\nCompression will start as soon as you upload a file or provide the URL to the image.",
			Action: compress,
		},
		"resize": {
			Name:	"resize",
			Usage:	"Use the API to create resized versions of your uploaded images. By letting the API handle resizing you avoid having to write such code yourself and you will only have to upload your image once. The resized images will be optimally compressed with a nice and crisp appearance.\nYou can also take advantage of intelligent cropping to create thumbnails that focus on the most visually important areas of your image.\nResizing counts as one additional compression. For example, if you upload a single image and retrieve the optimized version plus 2 resized versions this will count as 3 compressions in total.\nAvailable compression methods are: " + strings.Join(methods, ", "),
			Action: resize,
		},
		"convert": {
			Name:	"convert",
			Usage:	"You can use the API to convert your images to your desired image type. Tinify currently supports converting between: " + strings.Join(types, ",") + ".\n When you provide more than one image type in your convert request, the smallest version will be returned to you.\nImage converting will count as one additional compression.",
			Action: convert,
		},
		"help": {
			Name:	"help",
			Usage:	"Briefly explains command usage.",
			Action:	helpUsage,
		},
	}

	buildInfo, _ := debug.ReadBuildInfo()

	// Check if we have the API key on environment.
	// Note that we are using godotenv/autoload to automatically retrieve .env
	// and merge with the existing environment.
	key = os.Getenv("TINIFY_API_KEY")

	// Grab flags
	flag.Int8VarP(&debugLevel, "debug", "d", 4, "debug level; non-zero means no debug")
	flag.StringVarP(&imageName, "input", "i", "test.jpg", "input filename")
	flag.StringVarP(&outputFileName, "output", "o", "test.webp", "output filename")
	flag.StringVarP(&fileType, "type", "t", "webp", "file type [" + strings.Join(types, ", ") + "]")
	flag.StringVarP(&method, "method", "m", Tinify.ResizeMethodScale, "resizing method [" + strings.Join(methods, ", ") + "]")
	flag.Int64VarP(&width, "width", "w", 0, "destination image width")
	flag.Int64VarP(&height, "height", "g", 0, "destination image height")
	flag.BoolVarP(&help, "help", "h", false, "show usage")

	// Last chance to get a valid API key! See if it was passed via flags (not recommended)
	if key == "" {
		flag.StringVarP(&key, "key", "k", "", "Tinify API key (ideally read from environment)")
		if key == "" {
			// No key found anywhere, abort.
			fmt.Fprintln(os.Stderr, "the Tinify API key was not found anywhere (tried environment and CLI flags); cannot proceed")
			os.Exit(2)
		}
	}

	flag.Parse()

	// Help message and default usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,	"Usage of %s:\n", os.Args[0])
		fmt.Fprint(os.Stderr,	"Compresses/resizes/converts images using the Tinify API.\n" +
								"See https://tinypng.com/developers\n" +
								"This software is neither affiliated with, nor endorsed by Tinify B.V.\n",
								"Tinify Go Package version " + Tinify.VERSION + "\n")
		fmt.Fprintf(os.Stderr,	"Built with %v\n", buildInfo.GoVersion)

		fmt.Fprintf(os.Stderr, 	"TINIFY_API_KEY found? %t\n", key != "")

		fmt.Fprintln(os.Stderr,	"\nCOMMANDS:")
		for _, cmdHelp := range commands {
			fmt.Fprintf(os.Stderr, "- %s:\t%s\n\n", cmdHelp.Name, cmdHelp.Usage)
		}
		fmt.Fprintln(os.Stderr,	"FLAGS:");
		flag.PrintDefaults()
	}

	// If no args, print help.
	if len(os.Args) < 1 {
		// this will not give us help for commands
		flag.Usage()
		return
	}

	// Start zerolog logger, since we'll need it later.
	// We're also using it for pretty-printing to the console.
	logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.Level(debugLevel)).	// typecast from int8 to zerolog.Level
		With().
		Timestamp().
		Caller().
//		Int("pid", os.Getpid()).
//		Str("go_version", buildInfo.GoVersion).
		Logger()

	logger.Info().Msgf("Start debugging; tinify pkg version %s", Tinify.VERSION)

	// Extract command from command line; it must be the first parameter.
	command := flag.Arg(0)

	// note that "help" is a special command as well.
	if help {
		command = "help"
	}

	// Do some basic error-checking on flag parameters.
	// 1. Check if the type(s) are all valid:
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
				logger.Error().Msgf("invalid file format: %s", aFoundType)
				os.Exit(3)
			}
		}
		// if we're here, all file types are valid
		logger.Debug().Msg("all file type parameters are valid")
	} else {
		logger.Debug().Msg("no file type parameters found")
	}

	// 2. Check if the resizing method is a valid one.
	// First check if it's empty:
	if len(method) == 0 {
		method = Tinify.ResizeMethodScale	// scale is default
	} else if !slices.Contains(methods, method) {
		// Checked if it's one of the valid methods; if not, abort.
		logger.Error().Msgf("invalid resize method: %s", method)
		os.Exit(3)
	}

	// Prepare in advance some variables.
	// Set the API key:
	Tinify.SetKey(key)

	// Input file may be either an image filename or an URL; TinyPNG will handle both-
	// Since `://` is hardly a valid filename, but a requirement for being an URL;
	// handle URL later.
	// Note that if imageName is unset, stdin is assumed, even if it might not yet work.
	if !strings.Contains(imageName, "://") {
		f, err = os.Open(imageName)
		if err != nil {
			logger.Fatal().Err(err)
		}

		// Get the image file from disk/stdin.
		rawImage, err = io.ReadAll(f)
		if err != nil {
			logger.Fatal().Err(err)
		}

		logger.Debug().Msgf("Arg: %s, Size %d\n", imageName, len(rawImage))

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

	// Pass the non-flag arguments to the command;
	execStatus := executeCommand(command)
	if execStatus != 0 {
		logger.Error().Msgf("%s returned with error code %d\n", os.Args[0], execStatus)
		os.Exit(execStatus)
	}
} // main

var (
	rawImage []byte			// raw image file, when loaded from disk.
	err error				// declared here due to scope issues.
	f = os.Stdin			// file handler; STDIN by default.
	source *Tinify.Source	// declared in advance to avoid scoping issues.
)

// All-purpose API call. Whatever is done, it happens on the globals.
func callAPI() int {
	// If we have no explicit output filename, write directly to stdout
	if len(outputFileName) == 0 {
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

		logger.Debug().Msgf("wrote %d byte(s) to stdout\n", n)
		return 0
	}

	logger.Debug().Msgf("opening file %q for outputting image\n", outputFileName)

	// write to file, we have a special function for that already defined:
	err = source.ToFile(outputFileName)
	if err != nil {
		logger.Error().Err(err)
		return 1
	}

	logger.Debug().Msgf("opening file %q for outputting image\n", outputFileName)

	return 0
}

// Tries to get a list
func convert() int {
	return callAPI()
}

// Resizes image, given a width and a height.
func resize() int {
	// width and height are globals.
	if width == 0 && height == 0 {
		logger.Error().Msg("width and height cannot be simultaneously zero")
		return 2
	}

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
	return callAPI()
}

func helpUsage() int {
	explainCommand := flag.Arg(1)
	logger.Debug().Msgf("explaining command %q:\n", explainCommand)
	if _, ok := commands[explainCommand]; ok {
		fmt.Fprintf(os.Stderr, "- %s:\t%s\n", commands[explainCommand].Name, commands[explainCommand].Usage)
		return 0
	}
	// command unknown or empty, so print the usage.
	flag.Usage()
	return 1
}
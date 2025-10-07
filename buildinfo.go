// Auxiliary functions to return information from the built executable,
// such as version, Git commit, architcture, etc.
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

// versionInfoType holds the relevant information for this build.
// It is meant to be used as a cache.
type versionInfoType struct {
	version    string    // Runtime version.
	commit     string    // Commit revision number.
	dateString string    // Commit revision time (as a RFC3339 string).
	date       time.Time // Same as before, converted to a time.Time, because that's what the cli package uses.
	builtBy    string    // User who built this (see note).
	goOS       string    // Operating system for this build (from runtime).
	goARCH     string    // Architecture, i.e., CPU type (from runtime).
	goVersion  string    // Go version used to compile this build (from runtime).
	init       bool      // Have we already initialised the cache object?
}

// NOTE: I don't know where the "builtBy" information comes from, so, right now, it gets injected
// during build time, e.g. `go build -ldflags "-X main.TheBuilder=gwyneth"` (gwyneth 20231103)
// NOTE: debugLevel is set in main.

var (
	versionInfo *versionInfoType // cached values for this build.
	TheBuilder  string           // to be overwritten via the linker command `go build -ldflags "-X main.TheBuilder=gwyneth"`.
	TheVersion  string           // to be overwritten with -X main.TheVersion=X.Y.Z, as above.
)

// Initialises a versionInfo variable.
func initVersionInfo() (vI *versionInfoType, err error) {
	vI = new(versionInfoType)
	if vI.init {
		// already initialised, no need to do anything else!
		return vI, nil
	}
	// get the following entries from the runtime:
	vI.goOS = runtime.GOOS
	vI.goARCH = runtime.GOARCH
	vI.goVersion = runtime.Version()

	// attempt to get some build info as well:
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, fmt.Errorf("no valid build information found")
	}
	// use our supplied version instead of the long, useless, default Go version string.
	if TheVersion == "" {
		vI.version = buildInfo.Main.Version
	} else {
		vI.version = TheVersion
	}

	// Now dig through settings and extract what we can...

	var vcs, rev string // Name of the version control system name (very likely Git) and the revision.
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs":
			vcs = setting.Value
		case "vcs.revision":
			rev = setting.Value
		case "vcs.time":
			vI.dateString = setting.Value
		}
	}
	vI.commit = "unknown"
	if vcs != "" {
		vI.commit = vcs
	}
	if rev != "" {
		vI.commit += " [" + rev + "]"
	}
	// attempt to parse the date, which comes as a string in RFC3339 format, into a date.Time:
	var parseErr error
	if vI.date, parseErr = time.Parse(vI.dateString, time.RFC3339); parseErr != nil {
		// Note: we can safely ignore the parsing error: either the conversion works, or it doesn't, and we
		// cannot do anything about it... (gwyneth 20231103)
		// However, the AI revision bots dislike this, so we'll assign the current date instead.
		vI.date = time.Now()

		if setting.DebugLevel == "debug" || setting.DebugLevel == "info" || setting.DebugLevel == "trace" || setting.DebugLevel == "error" {
			fmt.Fprintf(os.Stderr, "date parse error: %v", parseErr)
		}
	}

	// see comment above
	vI.builtBy = TheBuilder
	// Mark initialisation as complete before returning.
	vI.init = true
	return vI, nil
}

// Returns a pretty-printed version of versionInfo, respecting the String() syntax.
func (vI *versionInfoType) String() string {
	return fmt.Sprintf(
		"\t%s\n\t(rev %s)\n\t[%s %s %s]\n\tBuilt on %s by %s",
		vI.version,
		vI.commit,
		vI.goOS,
		vI.goARCH,
		vI.goVersion,
		vI.dateString, // Date as string in RFC3339 notation.
		vI.builtBy,    // see note at the top...
	)
}

// Initialises a global, pre-defined versionInfo variable (we might just need one).
// Panics if allocation failed!
func init() {
	var err error
	if versionInfo, err = initVersionInfo(); err != nil {
		panic(err)
	}
}

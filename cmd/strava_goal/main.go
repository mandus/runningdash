// Custom Goal interpreter for the Strava Training Dashboard.
//
// This builds a Goal binary with the project's custom extensions linked in.
// As of EXT-2, http and sqlite extensions are registered. Additional extensions
// (json — EXT-3, httpserver — EXT-4) will be wired in here as those tasks complete.
//
// To build (from the repository root):
//
//	go build -o strava_goal ./cmd/strava_goal
//
// Usage:
//
//	./strava_goal script.goal   # run a Goal script
//	./strava_goal               # start REPL
package main

import (
	"os"

	"codeberg.org/anaseto/goal"
	"codeberg.org/anaseto/goal/cmd"
	"codeberg.org/anaseto/goal/help"

	// Standard Goal extensions (same set as Goal's "full" build).
	"codeberg.org/anaseto/goal/archive/zip"
	"codeberg.org/anaseto/goal/encoding/base64"
	"codeberg.org/anaseto/goal/math"
	gos "codeberg.org/anaseto/goal/os"

	// Project extensions.
	"github.com/mandus/runningdash/extensions/http"
	"github.com/mandus/runningdash/extensions/sqlite"
)

func main() {
	ctx := goal.NewContext()
	ctx.Log = os.Stderr

	// Standard Goal extensions: gives us say/print/read/open (os),
	// math.*, zip.*, base64.*.
	gos.Import(ctx, "")
	math.Import(ctx, "math")
	zip.Import(ctx, "")
	base64.Import(ctx, "")

	// Project extensions.
	http.Import(ctx, "")
	sqlite.Import(ctx, "")

	// Help function that includes our extensions on top of Goal's built-ins.
	hf := help.Wrap(
		help.HelpFunc(),
		math.HelpFunc(),
		zip.HelpFunc(),
		base64.HelpFunc(),
		http.HelpFunc(),
		sqlite.HelpFunc(),
	)

	cmd.Exit(cmd.Run(ctx, cmd.Config{
		Help: hf,
		Man:  "strava_goal",
	}))
}

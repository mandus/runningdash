// Custom Goal interpreter for Strava Training Dashboard
// This builds Goal with all required extensions for the project
//
// To build:
//   cd cmd/strava_goal
//   go mod init github.com/mandus/runningdash/cmd/strava_goal
//   go mod edit -replace codeberg.org/anaseto/goal=../goal
//   go mod tidy
//   go build
//
// The resulting binary will have all extensions loaded.
//
// Usage:
//   ./strava_goal [script.goal]  # Run a Goal script
//   ./strava_goal              # Start REPL
package main

import (
	"os"

	"codeberg.org/anaseto/goal"
	"codeberg.org/anaseto/goal/cmd"
	"codeberg.org/anaseto/goal/help"

	// Import our custom extensions
	"github.com/mandus/runningdash/extensions/http"
	"github.com/mandus/runningdash/extensions/json"
	"github.com/mandus/runningdash/extensions/sqlite"
)

func main() {
	ctx := goal.NewContext()
	ctx.Log = os.Stderr

	// Import standard extensions (from Goal's full.go)
	// Note: These require the Goal source to be available
	// For production, you may need to vendor these or use a different approach
	
	// Import our custom extensions
	http.Import(ctx, "")
	json.Import(ctx, "")
	sqlite.Import(ctx, "")

	// Create help function that includes our extensions
	hf := help.Wrap(
		help.HelpFunc(),
		http.HelpFunc(),
		json.HelpFunc(),
		sqlite.HelpFunc(),
	)

	// Run Goal with our context
	cmd.Exit(cmd.Run(ctx, cmd.Config{
		Help: hf,
		Man:  "strava_goal",
	}))
}

package main

import (
	"github.com/BuddhiLW/arara/internal/app"
)

// Binary-commands tree-branches will grow from the Root.
func main() {
	// Remove welcome message to avoid interfering with help output
	app.Cmd.Exec()
}

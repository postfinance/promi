// Package client represents the discovery client command.
package cmd

import (
	"github.com/zbindenren/king"
)

// CLI is the client command.
type CLI struct {
	Globals
	Alerts  alertCmd  `cmd:"" help:"Register and unregister servers." aliases:"srv"`
	Targets targetCmd `cmd:"" help:"Perform OIDC login."`
}

// Globals are the global client flags.
type Globals struct {
	PrometheusURLs []string         `short:"u" help:"A comma separated list of prometheus servers." default:"localhost:9090"`
	ShowConfig     king.ShowConfig  `help:"Show used config files"`
	Version        king.VersionFlag `help:"Show version information"`
	Debug          bool             `short:"d" help:"Show debug output." `
}

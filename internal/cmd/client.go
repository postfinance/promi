// Package cmd represents the discovery client command.
package cmd

import (
	"time"

	"github.com/postfinance/promcli/internal/prometheus"
	"github.com/zbindenren/king"
)

// CLI is the client command.
type CLI struct {
	Globals
	Alerts  alertCmd  `cmd:"" help:"Show alert status." aliases:"a"`
	Targets targetCmd `cmd:"" help:"Show target status."`
}

// Globals are the global client flags.
type Globals struct {
	PrometheusURLs []string         `short:"u" name:"prometheus-urls" help:"A comma separated list of prometheus base URLs." default:"http://localhost:9090"`
	ShowConfig     king.ShowConfig  `help:"Show used config files"`
	Version        king.VersionFlag `help:"Show version information"`
	Debug          bool             `short:"d" help:"Show debug output." `
	Timeout        time.Duration    `help:"The http request timeout." default:"10s"`
}

func (g Globals) client() (*prometheus.Client, error) {
	return prometheus.New(g.PrometheusURLs...)
}

package cmd

import (
	"github.com/alecthomas/kong"
	"go.uber.org/zap"
)

type alertCmd struct {
	Selector string `short:"s" help:"Kubernetes style selectors (key=value) to select servers with specific labels."`
}

func (a alertCmd) Run(g *Globals, l *zap.SugaredLogger, c *kong.Context) error {
	return nil
}

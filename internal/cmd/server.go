package cmd

import (
	"errors"
	"regexp"

	"github.com/alecthomas/kong"
	"github.com/postfinance/promi/internal/web"
	"github.com/zbindenren/king"
	"go.uber.org/zap"
)

type serverCmd struct {
	ListerAddr  string `default:":8080" help:"The TCP address for the server to listen on"`
	Deduplicate bool   `help:"Deduplicate targets by scrape url."`
}

func (s serverCmd) Run(g *Globals, l *zap.SugaredLogger, app *kong.Context) error {
	l.Infow("starting http server",
		king.FlagMap(app, regexp.MustCompile("key"), regexp.MustCompile("password"), regexp.MustCompile("secret")).
			Rm("help", "env-help", "version", "show-config", "etcd-ca", "etcd-cert").
			List()...)

	if len(g.PrometheusURLs) == 0 {
		return errors.New("no prometheus server configured")
	}

	a, err := web.New(l, g.Timeout, s.Deduplicate, g.PrometheusURLs...)
	if err != nil {
		return err
	}

	return a.Start()
}

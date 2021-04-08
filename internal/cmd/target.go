package cmd

import (
	"context"
	"os"
	"regexp"

	"github.com/alecthomas/kong"
	"github.com/postfinance/promcli/internal/prometheus"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/zbindenren/sfmt"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"
)

type targetCmd struct {
	Output       string `short:"o" default:"table" enum:"json,yaml,table" help:"Output format (table|json|yaml)."`
	Compact      bool   `short:"c" help:"Do not display labels and last error."`
	NoHeaders    bool   `short:"n" help:"Do not display headers in table output."`
	targetFilter `prefix:"filter-"`
}

func (t targetCmd) Run(g *Globals, l *zap.SugaredLogger, app *kong.Context) error {
	c, err := g.client()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.Timeout)
	defer cancel()

	targets, err := c.Targets(ctx)
	if err != nil {
		return err
	}

	filters, err := t.targetFilter.filters()
	if err != nil {
		return err
	}

	targets = targets.Filter(filters...)
	targets.Sort()

	if t.Compact {
		targets.Compact()
	}

	s := sfmt.SliceWriter{
		Writer:    os.Stdout,
		NoHeaders: t.NoHeaders,
	}

	format := sfmt.ParseFormat(t.Output)

	s.Write(format, targets)

	return nil
}

type targetFilter struct {
	Name      string          `short:"N" help:"Filter targets by job name (regular expression)."`
	Server    string          `short:"S" help:"Filter targets by promehteus server name (regular expression)."`
	ScrapeURL string          `short:"u" help:"Filter targets by scrape url (regular expression)."`
	Health    v1.HealthStatus `short:"H" help:"Filter targets by health (up|down)" enum:"up,down,"`
	Selector  string          `short:"s" help:"Filter services by (k8s style) selector."`
}

func (t targetFilter) filters() ([]prometheus.TargetFilterFunc, error) {
	filters := []prometheus.TargetFilterFunc{}

	if t.Name != "" {
		r, err := regexp.Compile(t.Name)
		if err != nil {
			return nil, err
		}

		filters = append(filters, prometheus.TargetByJob(r))
	}

	if t.ScrapeURL != "" {
		r, err := regexp.Compile(t.ScrapeURL)
		if err != nil {
			return nil, err
		}

		filters = append(filters, prometheus.TargetByScrapeURL(r))
	}

	if t.Server != "" {
		r, err := regexp.Compile(t.Server)
		if err != nil {
			return nil, err
		}

		filters = append(filters, prometheus.TargetByServer(r))
	}

	if t.Health != "" {
		filters = append(filters, prometheus.TargetByHealth(t.Health))
	}

	if t.Selector != "" {
		sel, err := labels.Parse(t.Selector)
		if err != nil {
			return nil, err
		}

		filters = append(filters, prometheus.TargetBySelector(sel))
	}

	return filters, nil
}

package cmd

import (
	"context"
	"os"
	"regexp"

	"github.com/alecthomas/kong"
	"github.com/postfinance/promi/internal/prometheus"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/zbindenren/sfmt"
	"go.uber.org/zap"
)

type alertCmd struct {
	Output      string `short:"o" default:"table" enum:"json,yaml,table" help:"Output format (table|json|yaml)."`
	NoHeaders   bool   `short:"n" help:"Do not display headers in table output."`
	alertFilter `prefix:"filter-"`
}

func (a alertCmd) Run(g *Globals, l *zap.SugaredLogger, app *kong.Context) error {
	c, err := g.client()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.Timeout)
	defer cancel()

	alerts, err := c.Alerts(ctx)
	if err != nil {
		return err
	}

	filters, err := a.alertFilter.filters()
	if err != nil {
		return err
	}

	alerts = alerts.Filter(filters...)

	s := sfmt.SliceWriter{
		Writer:    os.Stdout,
		NoHeaders: a.NoHeaders,
	}

	format := sfmt.ParseFormat(a.Output)

	s.Write(format, alerts)

	return nil
}

type alertFilter struct {
	Name   string        `short:"N" help:"Filter alerts by job name (regular expression)."`
	Alert  string        `short:"a" help:"Filter alerts by alert name (regular expression)."`
	Server string        `short:"S" help:"Filter alerts by prometheus server name (regular expression)."`
	State  v1.AlertState `short:"s" help:"Filter alerts by state (pending|firing)" enum:"pending,firing,"`
}

func (a alertFilter) filters() ([]prometheus.AlertFilterFunc, error) {
	filters := []prometheus.AlertFilterFunc{}

	if a.Name != "" {
		r, err := regexp.Compile(a.Name)
		if err != nil {
			return nil, err
		}

		filters = append(filters, prometheus.AlertByJob(r))
	}

	if a.Alert != "" {
		r, err := regexp.Compile(a.Alert)
		if err != nil {
			return nil, err
		}

		filters = append(filters, prometheus.AlertByName(r))
	}

	if a.Server != "" {
		r, err := regexp.Compile(a.Server)
		if err != nil {
			return nil, err
		}

		filters = append(filters, prometheus.AlertByServer(r))
	}

	if a.State != "" {
		filters = append(filters, prometheus.AlertByState(a.State))
	}

	return filters, nil
}

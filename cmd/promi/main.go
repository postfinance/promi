package main

import (
	"github.com/alecthomas/kong"
	"github.com/postfinance/flash"
	"github.com/postfinance/promi/internal/cmd"
	"github.com/zbindenren/king"
)

//nolint:gochecknoglobals //this vars are set on build by goreleaser
var (
	version = "0.0.0"
	commit  = "12345678"
	date    = "2020-09-23T07:03:55+02:00"
)

func main() {
	cli := cmd.CLI{}
	l := flash.New(flash.WithColor())

	b, err := king.NewBuildInfo(version,
		king.WithDateString(date),
		king.WithRevision(commit),
		king.WithLocation("Europe/Zurich"),
	)
	if err != nil {
		l.Fatal(err)
	}

	app := kong.Parse(&cli, king.DefaultOptions(
		king.Config{
			Name:        "promi",
			Description: "CLI to query targets and alerts of multiple prometheus servers.",
			BuildInfo:   b,
		},
	)...)

	l.SetDebug(cli.Debug)

	if err := app.Run(&cli.Globals, l.Get()); err != nil {
		l.Fatal(err)
	}
}

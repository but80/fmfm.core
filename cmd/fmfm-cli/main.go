package main

import (
	"os"

	"github.com/but80/fmfm.core/cmd/fmfm-cli/internal/subcmd"
	"github.com/urfave/cli"
)

var version string

func init() {
	if version == "" {
		version = "unknown"
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "fmfm-cli"
	app.Version = version
	app.Usage = "YAMAHA MA-5/YMF825 clone synthesizer"
	app.Authors = []cli.Author{
		{
			Name:  "but80",
			Email: "mersenne.sister@gmail.com",
		},
	}
	app.HelpName = "fmfm-cli"
	app.Commands = []cli.Command{
		subcmd.MIDI,
		subcmd.List,
	}
	app.Action = func(ctx *cli.Context) error {
		cli.ShowAppHelp(ctx)
		return nil
	}
	app.Run(os.Args)
}

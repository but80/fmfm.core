package subcmd

import (
	"fmt"

	"github.com/but80/fmfm.core/cmd/fmfm-cli/internal/player"
	"github.com/urfave/cli"
)

var List = cli.Command{
	Name:      "list",
	Aliases:   []string{"l"},
	Usage:     "List MIDI devices",
	ArgsUsage: " ",
	Flags:     []cli.Flag{},
	Action: func(ctx *cli.Context) error {
		devices := player.ListMIDIDeivces()
		for _, dev := range devices {
			fmt.Println(dev)
		}
		return nil
	},
}

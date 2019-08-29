package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	fmfm "github.com/but80/fmfm.core"
	"github.com/but80/fmfm.core/cmd/fmfm-cli/internal/player"
	"github.com/but80/fmfm.core/sim"
	"github.com/urfave/cli"
	"gopkg.in/but80/go-smaf.v1/pb/smaf"
)

var version string

func init() {
	if version == "" {
		version = "unknown"
	}
}

var listCmd = cli.Command{
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

var midiCmd = cli.Command{
	Name:      "midi",
	Aliases:   []string{"m"},
	Usage:     "Listen MIDI events",
	ArgsUsage: "[<Input MIDI device>]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "mono, m",
			Usage: `Force mono mode in all MIDI channels except drum PC`,
		},
		cli.BoolFlag{
			Name:  "mute-nopc, z",
			Usage: `Mute if program change is not found`,
		},
		cli.Float64Flag{
			Name:  "level, l",
			Usage: `Total level in dB`,
			Value: -12.0,
		},
		cli.Float64Flag{
			Name:  "limiter, c",
			Usage: `Limiter threshold in dB`,
			Value: -6.0,
		},
		cli.IntFlag{
			Name:  "ignore, n",
			Usage: `Ignore specified MIDI channel`,
		},
		cli.IntFlag{
			Name:  "solo, s",
			Usage: `Accept only specified MIDI channel`,
		},
		cli.IntFlag{
			Name:  "dump, d",
			Usage: `Dump MIDI channel`,
		},
		cli.BoolFlag{
			Name:  "print, p",
			Usage: `Print status`,
		},
	},
	Action: func(ctx *cli.Context) error {
		args := ctx.Args()
		midiDevice := ""
		if 1 <= ctx.NArg() {
			midiDevice = args[0]
		}

		info, err := ioutil.ReadDir("voice")
		if err != nil {
			panic(err)
		}
		var lib smaf.VM5VoiceLib
		for _, i := range info {
			if i.IsDir() || !strings.HasSuffix(i.Name(), ".vm5.pb") {
				continue
			}
			err := lib.LoadFile("voice/" + i.Name())
			if err != nil {
				panic(err)
			}
		}

		dumpMIDIChannel := -1
		if 0 < ctx.Int("dump") {
			dumpMIDIChannel = ctx.Int("dump") - 1
		}

		renderer := player.NewRenderer()
		limiter := player.NewLimiter(renderer.Parameters.SampleRate)
		limiter.SetThreshold(ctx.Float64("limiter"))
		renderer.Insert(limiter)
		chip := sim.NewChip(
			renderer.Parameters.SampleRate,
			ctx.Float64("level"),
			dumpMIDIChannel,
		)
		regs := sim.NewRegisters(chip)
		opts := &fmfm.ControllerOpts{
			Registers:          regs,
			Library:            &lib,
			MuteIfPCNotFound:   ctx.Bool("mute-nopc"),
			ForceMono:          ctx.Bool("mono"),
			PrintStatus:        ctx.Bool("print"),
			IgnoreMIDIChannels: []int{},
			SoloMIDIChannel:    dumpMIDIChannel,
		}
		if 0 < ctx.Int("ignore") {
			opts.IgnoreMIDIChannels = append(opts.IgnoreMIDIChannels, ctx.Int("ignore")-1)
		}
		if 0 < ctx.Int("solo") {
			for i := 0; i < 16; i++ {
				if i == ctx.Int("solo")-1 {
					continue
				}
				opts.IgnoreMIDIChannels = append(opts.IgnoreMIDIChannels, i)
			}
		}
		seq := player.NewSequencer(midiDevice, opts)
		defer seq.Close()
		renderer.Start(chip.Next, seq.FlushMIDIMessages)
		time.Sleep(24 * time.Hour)
		return nil
	},
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
		midiCmd,
		listCmd,
	}
	app.Action = func(ctx *cli.Context) error {
		cli.ShowAppHelp(ctx)
		return nil
	}
	app.Run(os.Args)
}

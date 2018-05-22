package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/but80/fmfm.core"
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

func main() {
	app := cli.NewApp()
	//app.EnableBashCompletion = true
	app.Name = "fmfm-cli"
	app.Version = version
	app.Description = "YAMAHA MA-5/YMF825 clone synthesizer"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "mono, m",
			Usage: `Force mono mode in all MIDI channel except drum note`,
		},
		cli.IntFlag{
			Name:  "ignore, n",
			Usage: `Ignore MIDI channel`,
		},
		cli.IntFlag{
			Name:  "solo, s",
			Usage: `Play only specified MIDI channel`,
		},
		cli.IntFlag{
			Name:  "dump, d",
			Usage: `Dump MIDI channel`,
		},
		cli.BoolFlag{
			Name:  "print, p",
			Usage: `Print status`,
		},
	}
	app.Authors = []cli.Author{
		{
			Name:  "but80",
			Email: "mersenne.sister@gmail.com",
		},
	}
	app.HelpName = "fmfm-cli"

	app.Action = func(ctx *cli.Context) error {
		info, err := ioutil.ReadDir("voice")
		if err != nil {
			panic(err)
		}
		var lib smaf.VM5VoiceLib
		for _, i := range info {
			if i.IsDir() || !strings.HasSuffix(i.Name(), ".vm5.pb") {
				continue
			}
			err := lib.LoadFile(i.Name())
			if err != nil {
				panic(err)
			}
		}

		dumpMIDIChannel := -1
		if 0 < ctx.Int("dump") {
			dumpMIDIChannel = ctx.Int("dump") - 1
		}

		renderer := player.NewRenderer()
		chip := sim.NewChip(renderer.Parameters.SampleRate, -15.0, dumpMIDIChannel)
		regs := sim.NewRegisters(chip)
		opts := &fmfm.ControllerOpts{
			Registers:          regs,
			Library:            &lib,
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
		seq := player.NewSequencer(opts)
		defer seq.Close()
		renderer.Start(chip.Next, seq.FlushMIDIMessages)
		time.Sleep(24 * time.Hour)
		return nil
	}

	app.Run(os.Args)
}

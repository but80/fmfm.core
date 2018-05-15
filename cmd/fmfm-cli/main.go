package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/but80/fmfm.core"
	"github.com/but80/fmfm.core/cmd/fmfm-cli/internal/player"
	"github.com/but80/fmfm.core/sim"
	"github.com/but80/smaf825/pb/smaf"
	"github.com/golang/protobuf/proto"
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
	//app.EnableBashCompletion = true
	app.Name = "fmfm-cli"
	app.Version = version
	app.Description = "YAMAHA MA-5/YMF825 clone synthesizer"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "ignore, n",
			Usage: `Ignore MIDI channel`,
		},
		cli.IntFlag{
			Name:  "dump, d",
			Usage: `Dump MIDI channel`,
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
		libs := []*smaf.VM5VoiceLib{}
		for _, i := range info {
			if i.IsDir() || !strings.HasSuffix(i.Name(), ".vm5.pb") {
				continue
			}
			b, err := ioutil.ReadFile("voice/" + i.Name())
			if err != nil {
				panic(err)
			}
			var lib smaf.VM5VoiceLib
			err = proto.Unmarshal(b, &lib)
			if err != nil {
				panic(err)
			}
			libs = append(libs, &lib)
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
			Libraries:          libs,
			IgnoreMIDIChannels: []int{},
			SoloMIDIChannel:    dumpMIDIChannel,
		}
		if 0 < ctx.Int("ignore") {
			opts.IgnoreMIDIChannels = append(opts.IgnoreMIDIChannels, ctx.Int("ignore")-1)
		}
		seq := player.NewSequencer(opts)
		defer seq.Close()
		renderer.Start(chip.Next, seq.FlushMIDIMessages)
		time.Sleep(24 * time.Hour)
		return nil
	}

	app.Run(os.Args)
}

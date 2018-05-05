package main

import (
	"time"

	"github.com/but80/fmfm/cmd/fmfm-cli/player"
	"github.com/but80/fmfm/ymf"
	"github.com/but80/smaf825/smaf/voice"
)

func main() {
	lib, err := voice.NewVM5VoiceLib("voice/default.vm5")
	if err != nil {
		panic(err)
	}

	renderer := player.NewRenderer()
	chip := ymf.NewChip(renderer.Parameters.SampleRate)
	seq := player.NewSequencer(chip, lib)
	seq.Reset()
	renderer.Start(chip.Next)
	time.Sleep(24 * time.Hour)
}

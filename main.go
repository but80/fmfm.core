package main

import (
	"time"

	"github.com/but80/smaf825/smaf/voice"
	"github.com/but80/fmfm/player"
	"github.com/but80/fmfm/ymf"
)

func main() {
	lib, err := voice.NewVM5VoiceLib("voice/default.vm5")
	if err != nil {
		panic(err)
	}

	renderer := player.NewRenderer()
	chip := ymf.NewChip(renderer.Parameters.SampleRate)
	seq := player.NewSequencer(chip, lib)
	seq.Load()
	renderer.Start(chip.Next)
	time.Sleep(24 * time.Hour)
}

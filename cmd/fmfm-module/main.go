package main

import (
	"C"

	"github.com/but80/fmfm/ymf"
)
import (
	"time"

	"github.com/but80/smaf825/smaf/voice"
)

func main() {
	// noop
}

var chip *ymf.Chip

func Init(string voicePath) {
	lib, err := voice.NewVM5VoiceLib(voicePath)
	if err != nil {
		panic(err)
	}

	chip := ymf.NewChip(renderer.Parameters.SampleRate)
	seq.Reset()
	renderer.Start(chip.Next)
	time.Sleep(24 * time.Hour)
}

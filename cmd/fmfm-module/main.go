package main

import (
	"C"
	"time"

	"github.com/but80/fmfm"
	"github.com/but80/fmfm/ymf"
	"github.com/but80/smaf825/smaf/voice"
)

func main() {
	// noop
}

var chip *ymf.Chip

//export Init
func Init(sampleRate float64, voicePath string) {
	lib, err := voice.NewVM5VoiceLib(voicePath)
	if err != nil {
		panic(err)
	}

	chip := ymf.NewChip(sampleRate)
	ctrl := fmfm.NewController(chip, lib)
	ctrl.Reset()
	time.Sleep(24 * time.Hour)
}

package main

import (
	"io/ioutil"
	"time"

	"github.com/but80/fmfm/cmd/fmfm-cli/internal/player"
	"github.com/but80/fmfm/ymf"
	"github.com/but80/smaf825/pb/smaf"
	"github.com/golang/protobuf/proto"
)

func main() {
	b, err := ioutil.ReadFile("voice/default.vm5.pb")
	if err != nil {
		panic(err)
	}
	var lib smaf.VM5VoiceLib
	err = proto.Unmarshal(b, &lib)
	if err != nil {
		panic(err)
	}

	renderer := player.NewRenderer()
	chip := ymf.NewChip(renderer.Parameters.SampleRate)
	seq := player.NewSequencer(chip, &lib)
	seq.Reset()
	renderer.Start(chip.Next)
	time.Sleep(24 * time.Hour)
}

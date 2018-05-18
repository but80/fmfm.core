package sim_test

import (
	"testing"

	"github.com/but80/fmfm.core"
	"github.com/but80/fmfm.core/sim"
	"github.com/but80/smaf825/pb/smaf"
	"github.com/google/gofuzz"
)

func TestNewChip(t *testing.T) {
	sampleRate := 44100.0

	for i := 0; i < 1000; i++ {
		var lib smaf.VM5VoiceLib
		fuzz.New().Fuzz(&lib)
		lib.Normalize()
		func() {
			defer func() {
				err := recover()
				if err != nil {
					t.Fatal(err)
				}
			}()

			chip := sim.NewChip(sampleRate, -15.0, -1)
			regs := sim.NewRegisters(chip)
			opts := &fmfm.ControllerOpts{
				Registers: regs,
				Libraries: []*smaf.VM5VoiceLib{&lib},
			}
			seq := fmfm.NewController(opts)
			chip.Next()

			seq.PushMIDIMessage(fmfm.MIDIProgramChange, 0, 0, 0, 0)
			seq.FlushMIDIMessages(0)
			chip.Next()

			seq.PushMIDIMessage(fmfm.MIDINoteOn, 0, 0, 60, 0)
			seq.FlushMIDIMessages(0)
			chip.Next()

			seq.PushMIDIMessage(fmfm.MIDINoteOn, 0, 0, 60, 0)
			seq.FlushMIDIMessages(0)
			chip.Next()
		}()
	}
}

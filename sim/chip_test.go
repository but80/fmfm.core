package sim_test

import (
	"testing"

	"github.com/google/gofuzz"
	fmfm "gopkg.in/but80/fmfm.core.v1"
	"gopkg.in/but80/fmfm.core.v1/sim"
	"gopkg.in/but80/go-smaf.v1/pb/smaf"
)

func TestNewChip(t *testing.T) {
	sampleRate := 44100.0
	f := fuzz.New()

	for i := 0; i < 1000; i++ {
		pc := smaf.VM35VoicePC{}
		f.Fuzz(&pc)
		pc.BankMsb = 0
		pc.BankLsb = 0
		pc.Pc = 0
		pc.DrumNote = 0
		pc.VoiceType = smaf.VoiceType_FM
		pc.FmVoice = &smaf.VM35FMVoice{}
		f.Fuzz(&pc.FmVoice)

		lib := &smaf.VM5VoiceLib{Programs: []*smaf.VM35VoicePC{&pc}}
		lib.Normalize()

		func() {
			chip := sim.NewChip(sampleRate, -15.0, -1)
			regs := sim.NewRegisters(chip)
			opts := &fmfm.ControllerOpts{
				Registers: regs,
				Library:   lib,
			}
			seq := fmfm.NewController(opts)
			chip.Next()

			seq.PushMIDIMessage(fmfm.MIDIControlChange, 1, 0, 0, 0)
			seq.PushMIDIMessage(fmfm.MIDIControlChange, 1, 0, 32, 0)
			seq.PushMIDIMessage(fmfm.MIDIProgramChange, 1, 0, 0, 0)
			seq.FlushMIDIMessages(2)
			chip.Next()

			seq.PushMIDIMessage(fmfm.MIDINoteOn, 3, 0, 60, 127)
			seq.FlushMIDIMessages(4)
			for j := 0; j < 100; j++ {
				chip.Next()
			}

			seq.PushMIDIMessage(fmfm.MIDINoteOff, 5, 0, 60, 0)
			seq.FlushMIDIMessages(6)
			for j := 0; j < 100; j++ {
				chip.Next()
			}
		}()
	}
}

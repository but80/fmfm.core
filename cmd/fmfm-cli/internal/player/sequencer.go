package player

import (
	"fmt"
	"sync"

	"github.com/but80/fmfm"
	"github.com/but80/fmfm/ymf"
	"github.com/but80/smaf825/pb/smaf"
	"github.com/xlab/portmidi"
)

const defaultMIDIDeviceName = "IAC YAMAHA Virtual MIDI Device 0"

// Sequencer は、PortMIDI により MIDIメッセージを受信して Chip のレジスタをコントロールします。
// TODO: rename
type Sequencer struct {
	*fmfm.Controller
}

var newSequencerOnce = sync.Once{}

// NewSequencer は、新しい Sequencer を作成します。
func NewSequencer(chip *ymf.Chip, libraries []*smaf.VM5VoiceLib) *Sequencer {
	newSequencerOnce.Do(func() {
		portmidi.Initialize()
		if portmidi.CountDevices() < 1 {
			panic("no midi device")
		}
	})

	selectedMIDIDeviceID, _ := portmidi.DefaultInputDeviceID()

	for i := 0; i < portmidi.CountDevices(); i++ {
		deviceID := portmidi.DeviceID(i)
		info := portmidi.GetDeviceInfo(deviceID)
		if info.IsInputAvailable && info.Name == defaultMIDIDeviceName {
			selectedMIDIDeviceID = deviceID
		}
	}

	in, err := portmidi.NewInputStream(selectedMIDIDeviceID, 512, 0)
	if err != nil {
		panic(err)
	}
	// defer in.Close()

	seq := &Sequencer{
		Controller: fmfm.NewController(chip, libraries),
	}

	go func() {
		for e := range in.Source() {
			if e.Timestamp < 0 {
				continue
			}
			msg := portmidi.Message(e.Message)
			status := int(msg.Status())
			channel := int(status & 15)
			switch status & 0xf0 {
			case 0x90:
				seq.NoteOn(channel, int(msg.Data1()), int(msg.Data2()))
			case 0x80:
				seq.NoteOff(channel, int(msg.Data1()))
			case 0xb0:
				seq.ControlChange(channel, int(msg.Data1()), int(msg.Data2()))
			case 0xc0:
				seq.ProgramChange(channel, int(msg.Data1()))
			case 0xe0:
				seq.PitchBend(channel, int(msg.Data1()), int(msg.Data2()))
			default:
				fmt.Printf("%x\n", status)
			}
		}
	}()

	return seq
}

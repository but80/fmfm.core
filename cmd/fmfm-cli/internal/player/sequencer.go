package player

import (
	"fmt"
	"sync"

	"github.com/xlab/portmidi"
	fmfm "gopkg.in/but80/fmfm.core.v1"
)

const defaultMIDIDeviceName = "IAC YAMAHA Virtual MIDI Device 0"

// Sequencer は、PortMIDI により MIDIメッセージを受信して Chip のレジスタをコントロールします。
// TODO: rename
type Sequencer struct {
	*fmfm.Controller
	input *portmidi.Stream
}

var newSequencerOnce = sync.Once{}

// NewSequencer は、新しい Sequencer を作成します。
func NewSequencer(midiDevice string, opts *fmfm.ControllerOpts) *Sequencer {
	if midiDevice == "@" {
		midiDevice = defaultMIDIDeviceName
	}

	newSequencerOnce.Do(func() {
		portmidi.Initialize()
		if portmidi.CountDevices() < 1 {
			panic("no midi device")
		}
	})

	var selectedMIDIDeviceID portmidi.DeviceID

	if midiDevice == "" {
		var found bool
		selectedMIDIDeviceID, found = portmidi.DefaultInputDeviceID()
		if !found {
			panic("No default MIDI device found")
		}
	} else {
		var found bool
		for i := 0; i < portmidi.CountDevices(); i++ {
			deviceID := portmidi.DeviceID(i)
			info := portmidi.GetDeviceInfo(deviceID)
			if info.IsInputAvailable && info.Name == midiDevice {
				selectedMIDIDeviceID = deviceID
				found = true
				break
			}
		}
		if !found {
			panic("No such MIDI device found: " + midiDevice)
		}
	}

	info := portmidi.GetDeviceInfo(selectedMIDIDeviceID)
	fmt.Printf("MIDI device: %s > %s\n", info.Interface, info.Name)

	input, err := portmidi.NewInputStream(selectedMIDIDeviceID, 512, 0)
	if err != nil {
		panic(err)
	}

	seq := &Sequencer{
		Controller: fmfm.NewController(opts),
		input:      input,
	}

	go func() {
		for e := range seq.input.Source() {
			if e.Timestamp < 0 {
				continue
			}
			msg := portmidi.Message(e.Message)
			status := int(msg.Status())
			channel := int(status & 15)
			var typ fmfm.MIDIMessage
			switch status & 0xf0 {
			case 0x90:
				typ = fmfm.MIDINoteOn
			case 0x80:
				typ = fmfm.MIDINoteOff
			case 0xb0:
				typ = fmfm.MIDIControlChange
			case 0xc0:
				typ = fmfm.MIDIProgramChange
			case 0xe0:
				typ = fmfm.MIDIPitchBend
			}
			seq.PushMIDIMessage(typ, int(e.Timestamp), channel, int(msg.Data1()), int(msg.Data2()))
		}
	}()

	return seq
}

// Close は、MIDIメッセージの受信を終了します。
func (seq *Sequencer) Close() {
	seq.input.Close()
}

// ListMIDIDeivces は、入力として選択可能なMIDIデバイスの一覧を取得します。
func ListMIDIDeivces() []string {
	result := []string{}
	for i := 0; i < portmidi.CountDevices(); i++ {
		deviceID := portmidi.DeviceID(i)
		info := portmidi.GetDeviceInfo(deviceID)
		if info.IsInputAvailable && info.Name != "" {
			result = append(result, info.Name)
		}
	}
	return result
}

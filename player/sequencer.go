package player

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/but80/smaf825/smaf/voice"
	"github.com/xlab/portmidi"
	"github.com/but80/fmfm/ymf"
	"github.com/but80/fmfm/ymf/ymfdata"
)

const defaultMIDIDeviceName = "IAC YAMAHA Virtual MIDI Device 0"

const (
	CH_SUSTAIN   = 0x02
	CH_VIBRATO   = 0x04
	CH_FREE      = 0x80
	HIGHEST_NOTE = 127
	MOD_MIN      = 40
)

type MIDIControl int

const (
	MIDIControl_bank         MIDIControl = 0
	MIDIControl_modulation   MIDIControl = 1
	MIDIControl_dataEntryHi  MIDIControl = 6
	MIDIControl_volume       MIDIControl = 7
	MIDIControl_pan          MIDIControl = 10
	MIDIControl_expression   MIDIControl = 11
	MIDIControl_dataEntryLo  MIDIControl = 38
	MIDIControl_sustainPedal MIDIControl = 64
	MIDIControl_softPedal    MIDIControl = 67
	MIDIControl_reverb       MIDIControl = 91
	MIDIControl_chorus       MIDIControl = 93
	MIDIControl_nRPNLo       MIDIControl = 98
	MIDIControl_nRPNHi       MIDIControl = 99
	MIDIControl_rpnLo        MIDIControl = 100
	MIDIControl_rpnHi        MIDIControl = 101
	MIDIControl_soundsOff    MIDIControl = 120
	MIDIControl_notesOff     MIDIControl = 123
	MIDIControl_mono         MIDIControl = 126
	MIDIControl_poly         MIDIControl = 127
)

type Slot struct {
	channel    int
	note       int
	realnote   int
	flags      int
	finetune   int
	pitch      int
	velocity   int
	instrument *voice.VM35VoicePC
	time       time.Time
}

type ChannelState struct {
	instr      uint32
	volume     uint8
	expression uint8
	pan        uint8
	pitch      int8
	sustain    uint8
	modulation uint8
	pitchSens  uint16
	rpn        uint16
}

type Sequencer struct {
	chip    *ymf.Chip
	library *voice.VM5VoiceLib

	channelStates [16]*ChannelState
	slots         [ymfdata.CHANNEL_COUNT]*Slot
}

var newSequencerOnce = sync.Once{}

func NewSequencer(chip *ymf.Chip, library *voice.VM5VoiceLib) *Sequencer {
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

	seq := &Sequencer{
		chip:    chip,
		library: library,
	}
	for i := range seq.slots {
		seq.slots[i] = &Slot{}
	}
	for i := range seq.channelStates {
		seq.channelStates[i] = &ChannelState{}
	}

	in, err := portmidi.NewInputStream(selectedMIDIDeviceID, 512, 0)
	if err != nil {
		panic(err)
	}
	// defer in.Close()

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
				seq.onNoteOn(channel, int(msg.Data1()), int(msg.Data2()))
			case 0x80:
				seq.onNoteOff(channel, int(msg.Data1()))
			case 0xb0:
				seq.onControlChange(channel, int(msg.Data1()), int(msg.Data2()))
			case 0xc0:
				seq.onProgramChange(channel, int(msg.Data1()))
			case 0xe0:
				seq.onPitchBend(channel, int(msg.Data1()), int(msg.Data2()))
			default:
				fmt.Printf("%x\n", status)
			}
		}
	}()

	return seq
}

func (seq *Sequencer) onNoteOn(ch, note, velocity int) {
	// TODO: remove
	if ch == 9 {
		return
	}
	if velocity == 0 {
		seq.onNoteOff(ch, note)
		return
	}

	instr, ok := seq.getInstrument(ch, note)
	if !ok {
		// TODO: warning
		return
	}

	slotID := seq.findFreeSlot(ch, note)
	if 0 <= slotID {
		seq.occupySlot(slotID, ch, note, velocity, instr)
	} else {
		// TODO: warning
	}
}

func (seq *Sequencer) onNoteOff(ch, note int) {
	sus := seq.channelStates[ch].sustain
	for slotID, slot := range seq.slots {
		if slot.channel == ch && slot.note == note {
			if sus < 0x40 {
				seq.releaseSlot(slotID, false)
			} else {
				slot.flags |= CH_SUSTAIN
			}
		}
	}
}

func (seq *Sequencer) onControlChange(ch, cc, value int) {
	seq.ymfChangeControl(ch, MIDIControl(cc), value)
}

func (seq *Sequencer) onProgramChange(ch, value int) {
	seq.ymfProgramChange(ch, value)
}

func (seq *Sequencer) onPitchBend(ch, l, h int) {
	seq.ymfPitchWheel(ch, h*128+l)
}

func (seq *Sequencer) Load() {
	for _, slot := range seq.slots {
		slot.channel = -1
		slot.note = 0
		slot.flags = 0
		slot.realnote = 0
		slot.finetune = 0
		slot.pitch = 0
		slot.velocity = 0
		slot.instrument = nil
		slot.time = time.Time{}
	}
	for _, state := range seq.channelStates {
		state.volume = 100
		state.pan = 64
	}
	seq.rewind()
}

func (seq *Sequencer) rewind() {
	seq.ymfShutup()
	seq.ymfStopMusic()
	seq.ymfPlayMusic()
}

func (seq *Sequencer) writeFrequency(slot, note, pitch int, keyon bool) {
	seq.ymfWriteFreq(slot, note, pitch, keyon)
}

func (seq *Sequencer) writeModulation(slot int, instr *voice.VM35VoicePC, state bool) {
	fmvoice, ok := instr.Voice.(*voice.VM35FMVoice)
	if !ok {
		// TODO: warning
		return
	}

	// TODO: モジュレータではevbだけを見る(stateは無視)？
	o := fmvoice.Operators
	seq.ymfWriteSlotEachOps(
		ymf.OpRegister_EVB,
		slot,
		bool2int(o[0].EVB || state),
		bool2int(o[1].EVB || state),
		bool2int(o[2].EVB || state),
		bool2int(o[3].EVB || state),
	)
}

func (seq *Sequencer) occupySlot(slotID, channel, note, velocity int, instr *voice.VM35VoicePC) {
	state := seq.channelStates[channel]
	slot := seq.slots[slotID]
	slot.channel = channel
	slot.note = note
	slot.flags = 0
	if MOD_MIN <= state.modulation {
		slot.flags |= CH_VIBRATO
	}
	slot.time = time.Now()

	fmvoice, ok := instr.Voice.(*voice.VM35FMVoice)
	if !ok {
		// TODO: warning
		return
	}

	slot.velocity = velocity
	if instr.DrumNote != 0 {
		note = int(fmvoice.DrumKey)
	} else {
		slot.finetune = 0
	}
	slot.pitch = slot.finetune + int(state.pitch)
	slot.instrument = instr
	if instr.DrumNote == 0 {
		// for note < 0 {
		// 	note += 12
		// }
		// for HIGHEST_NOTE < note {
		// 	note -= 12
		// }
	}
	note += 2 - 12
	slot.realnote = note

	seq.ymfWriteInstrument(slotID, instr)
	if slot.flags&CH_VIBRATO != 0 {
		seq.writeModulation(slotID, instr, true)
	}
	seq.chip.WriteChannel(ymf.ChRegister_CHPAN, slotID, int(seq.channelStates[channel].pan))
	seq.chip.WriteChannel(ymf.ChRegister_VOLUME, slotID, int(seq.channelStates[channel].volume))
	seq.chip.WriteChannel(ymf.ChRegister_EXPRESSION, slotID, int(seq.channelStates[channel].expression))
	seq.chip.WriteChannel(ymf.ChRegister_BO, slotID, int(fmvoice.BO))
	seq.ymfWriteVelocity(slotID, slot.velocity, instr)
	seq.writeFrequency(slotID, note, slot.pitch, true)
}

func (seq *Sequencer) releaseSlot(slotID int, killed bool) {
	slot := seq.slots[slotID]
	seq.writeFrequency(slotID, slot.realnote, slot.pitch, false)
	slot.channel = -1
	slot.time = time.Now()
	slot.flags = CH_FREE
	if killed {
		seq.ymfWriteSlotAllOps(ymf.OpRegister_SL, slotID, 0)
		seq.ymfWriteSlotAllOps(ymf.OpRegister_RR, slotID, 15) // release rate - fastest
		seq.ymfWriteSlotAllOps(ymf.OpRegister_KSL, slotID, 0)
		seq.ymfWriteSlotAllOps(ymf.OpRegister_TL, slotID, 0x3f) // no volume
	}
}

func (seq *Sequencer) releaseSustain(channel int) {
	for i, slot := range seq.slots {
		if slot.channel == channel && slot.flags&CH_SUSTAIN != 0 {
			seq.releaseSlot(i, false)
		}
	}
}

func (seq *Sequencer) findFreeSlot(channel, note int) int {
	for i := 0; i < len(seq.slots); i++ {
		if seq.slots[i].flags&CH_FREE != 0 {
			return i
		}
	}

	oldest := -1
	oldesttime := time.Now()

	// find some 2nd-voice channel and determine the oldest
	for i := 0; i < len(seq.slots); i++ {
		if seq.slots[i].time.Before(oldesttime) {
			oldesttime = seq.slots[i].time
			oldest = i
		}
	}

	// if possible, kill the oldest channel
	if 0 <= oldest {
		seq.releaseSlot(oldest, true)
		return oldest
	}

	// can't find any free channel
	return -1
}

func (seq *Sequencer) getInstrument(channel, note int) (*voice.VM35VoicePC, bool) {
	// TODO: smaf825側で検索
	// TODO: ドラム音色
	n := int(seq.channelStates[channel].instr)
	for _, p := range seq.library.Programs {
		if p.PC == n {
			return p, true
		}
	}
	fmt.Printf("voice not found: @%d\n", n)
	return seq.library.Programs[0], false
}

func (seq *Sequencer) ymfPitchWheel(channel, pitch int) {
	// Convert pitch from 14-bit to 7-bit, then scale it, since the player
	// code only understands sensitivities of 2 semitones.
	pitch = int(float64(pitch-8192)*float64(seq.channelStates[channel].pitchSens)/(200*128) + 64)
	seq.channelStates[channel].pitch = int8(pitch)
	for i, slot := range seq.slots {
		if slot.channel == channel {
			slot.time = time.Now()
			slot.pitch = slot.finetune + pitch
			seq.writeFrequency(i, slot.realnote, slot.pitch, true)
		}
	}
}

func (seq *Sequencer) ymfChangeControl(channel int, controller MIDIControl, value int) {
	switch controller {
	case MIDIControl_modulation:
		seq.channelStates[channel].modulation = uint8(value)
		for i, slot := range seq.slots {
			if slot.channel == channel {
				flags := slot.flags
				slot.time = time.Now()
				if MOD_MIN <= value {
					slot.flags |= CH_VIBRATO
					if slot.flags != flags {
						seq.writeModulation(i, slot.instrument, true)
					}
				} else {
					slot.flags &= ^CH_VIBRATO
					if slot.flags != flags {
						seq.writeModulation(i, slot.instrument, false)
					}
				}
			}
		}

	case MIDIControl_volume: // change volume
		seq.channelStates[channel].volume = uint8(value)
		for i, slot := range seq.slots {
			if slot.channel == channel {
				slot.time = time.Now()
				seq.chip.WriteChannel(ymf.ChRegister_VOLUME, i, value)
			}
		}

	case MIDIControl_expression: // change expression
		seq.channelStates[channel].expression = uint8(value)
		for i, slot := range seq.slots {
			if slot.channel == channel {
				slot.time = time.Now()
				seq.chip.WriteChannel(ymf.ChRegister_EXPRESSION, i, value)
			}
		}

	case MIDIControl_pan: // change pan (balance)
		seq.channelStates[channel].pan = uint8(value)
		for i, slot := range seq.slots {
			if slot.channel == channel {
				slot.time = time.Now()
				seq.chip.WriteChannel(ymf.ChRegister_CHPAN, i, value)
			}
		}

	case MIDIControl_sustainPedal: // change sustain pedal (hold)
		seq.channelStates[channel].sustain = uint8(value)
		if value < 0x40 {
			seq.releaseSustain(channel)
		}

	case MIDIControl_notesOff: // turn off all notes that are not sustained
		for i, slot := range seq.slots {
			if slot.channel == channel {
				if seq.channelStates[channel].sustain < 0x40 {
					seq.releaseSlot(i, false)
				} else {
					slot.flags |= CH_SUSTAIN
				}
			}
		}

	case MIDIControl_soundsOff: // release all notes for this channel
		for i, slot := range seq.slots {
			if slot.channel == channel {
				seq.releaseSlot(i, false)
			}
		}

	case MIDIControl_rpnHi:
		seq.channelStates[channel].rpn = (seq.channelStates[channel].rpn & 0x007f) | (uint16(value) << 7)

	case MIDIControl_rpnLo:
		seq.channelStates[channel].rpn = (seq.channelStates[channel].rpn & 0x3f80) | uint16(value)

	case MIDIControl_nRPNLo, MIDIControl_nRPNHi:
		seq.channelStates[channel].rpn = 0x3fff

	case MIDIControl_dataEntryHi:
		if seq.channelStates[channel].rpn == 0 {
			seq.channelStates[channel].pitchSens = uint16(value)*100 + (seq.channelStates[channel].pitchSens % 100)
		}

	case MIDIControl_dataEntryLo:
		if seq.channelStates[channel].rpn == 0 {
			seq.channelStates[channel].pitchSens = uint16(value) + uint16(seq.channelStates[channel].pitchSens/100)*100
		}
	}
}

func (seq *Sequencer) ymfProgramChange(channel, value int) {
	seq.channelStates[channel].instr = uint32(value)
}

func (seq *Sequencer) ymfResetControllers(channel int) {
	seq.channelStates[channel].volume = 100
	seq.channelStates[channel].expression = 127
	seq.channelStates[channel].sustain = 0
	seq.channelStates[channel].pitch = 64
	seq.channelStates[channel].rpn = 0x3fff
	seq.channelStates[channel].pitchSens = 200
}

func (seq *Sequencer) ymfPlayMusic() {
	for i := range seq.slots {
		seq.ymfResetControllers(i)
	}
}

func (seq *Sequencer) ymfStopMusic() {
	for i := range seq.slots {
		if seq.slots[i].flags&CH_FREE == 0 {
			seq.releaseSlot(i, true)
		}
	}
}

func (seq *Sequencer) ymfWriteSlotAllOps(regbase ymf.OpRegister, slotID, data int) {
	seq.chip.WriteOperator(regbase, slotID, 0, data)
	seq.chip.WriteOperator(regbase, slotID, 1, data)
	seq.chip.WriteOperator(regbase, slotID, 2, data)
	seq.chip.WriteOperator(regbase, slotID, 3, data)
}

func (seq *Sequencer) ymfWriteSlotEachOps(regbase ymf.OpRegister, slotID, data1, data2, data3, data4 int) {
	seq.chip.WriteOperator(regbase, slotID, 0, data1)
	seq.chip.WriteOperator(regbase, slotID, 1, data2)
	seq.chip.WriteOperator(regbase, slotID, 2, data3)
	seq.chip.WriteOperator(regbase, slotID, 3, data4)
}

func (seq *Sequencer) ymfWriteFreq(slotID, note, pitch int, keyon bool) {
	n := float64(note-ymfdata.A3Note) + float64(pitch-64)/32.0
	freq := ymfdata.A3Freq * math.Pow(2.0, n/12.0)

	block := note / 12
	if 7 < block {
		block = 7
	}

	fnum := int(freq*ymfdata.FnumK) >> uint(block-1)
	if fnum < 0 {
		fnum = 0
	} else {
		for 1024 < fnum {
			block++
			fnum >>= 1
		}
	}
	if block < 0 {
		block = 0
	} else if 7 < block {
		block = 7
	}

	seq.chip.WriteChannel(ymf.ChRegister_FNUM, slotID, fnum)
	seq.chip.WriteChannel(ymf.ChRegister_BLOCK, slotID, block)
	k := 0
	if keyon {
		k = 1
	}
	seq.chip.WriteChannel(ymf.ChRegister_KON, slotID, k)
}

func ymfConvertVelocity(data, velocity int) int {
	r := int(velocitytable[velocity])
	return 0x3f - ((0x3f - data) * r >> 7)
}

func (seq *Sequencer) ymfWriteVelocity(slotID, velocity int, instr *voice.VM35VoicePC) {
	v, ok := instr.Voice.(*voice.VM35FMVoice)
	if !ok {
		// TODO: warning
		return
	}
	ops := seq.chip.Channels[slotID].Operators
	for i, op := range v.Operators {
		v := op.TL
		if !ops[i].IsModulator {
			v = ymfConvertVelocity(op.TL, velocity)
		}
		seq.chip.WriteOperator(ymf.OpRegister_TL, slotID, i, v)
	}
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (seq *Sequencer) ymfWriteInstrument(slotID int, instr *voice.VM35VoicePC) {
	v, ok := instr.Voice.(*voice.VM35FMVoice)
	if !ok {
		// TODO: warning
		return
	}
	seq.ymfWriteSlotAllOps(ymf.OpRegister_TL, slotID, 0x3f) // no volume

	for i, op := range v.Operators {
		seq.chip.WriteOperator(ymf.OpRegister_EAM, slotID, i, bool2int(op.EAM))
		seq.chip.WriteOperator(ymf.OpRegister_EVB, slotID, i, bool2int(op.EVB))
		seq.chip.WriteOperator(ymf.OpRegister_DAM, slotID, i, op.DAM)
		seq.chip.WriteOperator(ymf.OpRegister_DVB, slotID, i, op.DVB)
		seq.chip.WriteOperator(ymf.OpRegister_DT, slotID, i, op.DT)
		seq.chip.WriteOperator(ymf.OpRegister_KSL, slotID, i, op.KSL)
		seq.chip.WriteOperator(ymf.OpRegister_KSR, slotID, i, bool2int(op.KSR))
		seq.chip.WriteOperator(ymf.OpRegister_WS, slotID, i, op.WS)
		seq.chip.WriteOperator(ymf.OpRegister_MULT, slotID, i, int(op.MULTI))
		seq.chip.WriteOperator(ymf.OpRegister_FB, slotID, i, op.FB)
		seq.chip.WriteOperator(ymf.OpRegister_AR, slotID, i, op.AR)
		seq.chip.WriteOperator(ymf.OpRegister_DR, slotID, i, op.DR)
		seq.chip.WriteOperator(ymf.OpRegister_SL, slotID, i, op.SL)
		seq.chip.WriteOperator(ymf.OpRegister_SR, slotID, i, op.SR)
		seq.chip.WriteOperator(ymf.OpRegister_RR, slotID, i, op.RR)
		seq.chip.WriteOperator(ymf.OpRegister_TL, slotID, i, op.TL)
		seq.chip.WriteOperator(ymf.OpRegister_XOF, slotID, i, bool2int(op.XOF))
	}

	seq.chip.WriteChannel(ymf.ChRegister_ALG, slotID, int(v.ALG))
	seq.chip.WriteChannel(ymf.ChRegister_LFO, slotID, v.LFO)
	seq.chip.WriteChannel(ymf.ChRegister_PANPOT, slotID, int(v.PANPOT))
	seq.chip.WriteChannel(ymf.ChRegister_BO, slotID, int(v.BO))
}

func (seq *Sequencer) ymfShutup() {
	for i := range seq.slots {
		seq.ymfWriteSlotAllOps(ymf.OpRegister_KSL, i, 0)
		seq.ymfWriteSlotAllOps(ymf.OpRegister_TL, i, 0x3f) // turn off volume
		seq.ymfWriteSlotAllOps(ymf.OpRegister_AR, i, 15)   // the fastest attack,
		seq.ymfWriteSlotAllOps(ymf.OpRegister_DR, i, 15)   // decay
		seq.ymfWriteSlotAllOps(ymf.OpRegister_SL, i, 0)    //
		seq.ymfWriteSlotAllOps(ymf.OpRegister_RR, i, 15)   // ... and release
		seq.chip.WriteChannel(ymf.ChRegister_KON, i, 0)    // KEY-OFF
	}
}

var velocitytable = [...]uint8{
	0, 1, 3, 5, 6, 8, 10, 11,
	13, 14, 16, 17, 19, 20, 22, 23,
	25, 26, 27, 29, 30, 32, 33, 34,
	36, 37, 39, 41, 43, 45, 47, 49,
	50, 52, 54, 55, 57, 59, 60, 61,
	63, 64, 66, 67, 68, 69, 71, 72,
	73, 74, 75, 76, 77, 79, 80, 81,
	82, 83, 84, 84, 85, 86, 87, 88,
	89, 90, 91, 92, 92, 93, 94, 95,
	96, 96, 97, 98, 99, 99, 100, 101,
	101, 102, 103, 103, 104, 105, 105, 106,
	107, 107, 108, 109, 109, 110, 110, 111,
	112, 112, 113, 113, 114, 114, 115, 115,
	116, 117, 117, 118, 118, 119, 119, 120,
	120, 121, 121, 122, 122, 123, 123, 123,
	124, 124, 125, 125, 126, 126, 127, 127,
}

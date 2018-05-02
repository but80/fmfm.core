package player

import (
	"fmt"
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
	LATENCY      = 2
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
	MIDIControl_rPNLo        MIDIControl = 100
	MIDIControl_rPNHi        MIDIControl = 101
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

	channelStates [ymfdata.CHANNEL_COUNT]*ChannelState
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
	for i := 0; i < ymfdata.CHANNEL_COUNT; i++ {
		seq.channelStates[i] = &ChannelState{}
		seq.slots[i] = &Slot{}
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
		seq.occupyChannel(slotID, ch, note, velocity, instr)
	} else {
		// TODO: warning
	}
}

func (seq *Sequencer) onNoteOff(ch, note int) {
	sus := seq.channelStates[ch].sustain
	for slotID, slot := range seq.slots {
		if slot.channel == ch && slot.note == note {
			if sus < 0x40 {
				seq.releaseChannel(slotID, false)
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

func (seq *Sequencer) occupyChannel(slotID, channel, note, velocity int, instr *voice.VM35VoicePC) {
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
		note += 2 - 12
	}
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

func (seq *Sequencer) releaseChannel(slotID int, killed bool) {
	slot := seq.slots[slotID]
	seq.writeFrequency(slotID, slot.realnote, slot.pitch, false)
	slot.channel |= CH_FREE
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
			seq.releaseChannel(i, false)
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
		seq.releaseChannel(oldest, true)
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
					seq.releaseChannel(i, false)
				} else {
					slot.flags |= CH_SUSTAIN
				}
			}
		}

	case MIDIControl_soundsOff: // release all notes for this channel
		for i, slot := range seq.slots {
			if slot.channel == channel {
				seq.releaseChannel(i, false)
			}
		}

	case MIDIControl_rPNHi:
		seq.channelStates[channel].rpn = (seq.channelStates[channel].rpn & 0x007f) | (uint16(value) << 7)

	case MIDIControl_rPNLo:
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
			seq.releaseChannel(i, true)
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
	octave := 0
	j := (note << 5) + pitch

	if j < 0 {
		j = 0
	} else if 284 <= j {
		j -= 284
		octave = j / (32 * 12)
		j = j%(32*12) + 284
		for 7 < octave {
			octave--
			if j+32*12 < len(frequencies) {
				j += 32 * 12
			}
		}
	}

	seq.chip.WriteChannel(ymf.ChRegister_FNUM, slotID, int(frequencies[j]))
	seq.chip.WriteChannel(ymf.ChRegister_BLOCK, slotID, octave)
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

var frequencies = [...]uint16{
	0x133, 0x133, 0x134, 0x134, 0x135, 0x136, 0x136, 0x137, // -1
	0x137, 0x138, 0x138, 0x139, 0x139, 0x13a, 0x13b, 0x13b,
	0x13c, 0x13c, 0x13d, 0x13d, 0x13e, 0x13f, 0x13f, 0x140,
	0x140, 0x141, 0x142, 0x142, 0x143, 0x143, 0x144, 0x144,

	0x145, 0x146, 0x146, 0x147, 0x147, 0x148, 0x149, 0x149, // -2
	0x14a, 0x14a, 0x14b, 0x14c, 0x14c, 0x14d, 0x14d, 0x14e,
	0x14f, 0x14f, 0x150, 0x150, 0x151, 0x152, 0x152, 0x153,
	0x153, 0x154, 0x155, 0x155, 0x156, 0x157, 0x157, 0x158,

	// These are used for the first seven MIDI note values:

	0x158, 0x159, 0x15a, 0x15a, 0x15b, 0x15b, 0x15c, 0x15d, // 0
	0x15d, 0x15e, 0x15f, 0x15f, 0x160, 0x161, 0x161, 0x162,
	0x162, 0x163, 0x164, 0x164, 0x165, 0x166, 0x166, 0x167,
	0x168, 0x168, 0x169, 0x16a, 0x16a, 0x16b, 0x16c, 0x16c,

	0x16d, 0x16e, 0x16e, 0x16f, 0x170, 0x170, 0x171, 0x172, // 1
	0x172, 0x173, 0x174, 0x174, 0x175, 0x176, 0x176, 0x177,
	0x178, 0x178, 0x179, 0x17a, 0x17a, 0x17b, 0x17c, 0x17c,
	0x17d, 0x17e, 0x17e, 0x17f, 0x180, 0x181, 0x181, 0x182,

	0x183, 0x183, 0x184, 0x185, 0x185, 0x186, 0x187, 0x188, // 2
	0x188, 0x189, 0x18a, 0x18a, 0x18b, 0x18c, 0x18d, 0x18d,
	0x18e, 0x18f, 0x18f, 0x190, 0x191, 0x192, 0x192, 0x193,
	0x194, 0x194, 0x195, 0x196, 0x197, 0x197, 0x198, 0x199,

	0x19a, 0x19a, 0x19b, 0x19c, 0x19d, 0x19d, 0x19e, 0x19f, // 3
	0x1a0, 0x1a0, 0x1a1, 0x1a2, 0x1a3, 0x1a3, 0x1a4, 0x1a5,
	0x1a6, 0x1a6, 0x1a7, 0x1a8, 0x1a9, 0x1a9, 0x1aa, 0x1ab,
	0x1ac, 0x1ad, 0x1ad, 0x1ae, 0x1af, 0x1b0, 0x1b0, 0x1b1,

	0x1b2, 0x1b3, 0x1b4, 0x1b4, 0x1b5, 0x1b6, 0x1b7, 0x1b8, // 4
	0x1b8, 0x1b9, 0x1ba, 0x1bb, 0x1bc, 0x1bc, 0x1bd, 0x1be,
	0x1bf, 0x1c0, 0x1c0, 0x1c1, 0x1c2, 0x1c3, 0x1c4, 0x1c4,
	0x1c5, 0x1c6, 0x1c7, 0x1c8, 0x1c9, 0x1c9, 0x1ca, 0x1cb,

	0x1cc, 0x1cd, 0x1ce, 0x1ce, 0x1cf, 0x1d0, 0x1d1, 0x1d2, // 5
	0x1d3, 0x1d3, 0x1d4, 0x1d5, 0x1d6, 0x1d7, 0x1d8, 0x1d8,
	0x1d9, 0x1da, 0x1db, 0x1dc, 0x1dd, 0x1de, 0x1de, 0x1df,
	0x1e0, 0x1e1, 0x1e2, 0x1e3, 0x1e4, 0x1e5, 0x1e5, 0x1e6,

	0x1e7, 0x1e8, 0x1e9, 0x1ea, 0x1eb, 0x1ec, 0x1ed, 0x1ed, // 6
	0x1ee, 0x1ef, 0x1f0, 0x1f1, 0x1f2, 0x1f3, 0x1f4, 0x1f5,
	0x1f6, 0x1f6, 0x1f7, 0x1f8, 0x1f9, 0x1fa, 0x1fb, 0x1fc,
	0x1fd, 0x1fe, 0x1ff, 0x200, 0x201, 0x201, 0x202, 0x203,

	// First note of looped range used for all octaves:

	0x204, 0x205, 0x206, 0x207, 0x208, 0x209, 0x20a, 0x20b, // 7
	0x20c, 0x20d, 0x20e, 0x20f, 0x210, 0x210, 0x211, 0x212,
	0x213, 0x214, 0x215, 0x216, 0x217, 0x218, 0x219, 0x21a,
	0x21b, 0x21c, 0x21d, 0x21e, 0x21f, 0x220, 0x221, 0x222,

	0x223, 0x224, 0x225, 0x226, 0x227, 0x228, 0x229, 0x22a, // 8
	0x22b, 0x22c, 0x22d, 0x22e, 0x22f, 0x230, 0x231, 0x232,
	0x233, 0x234, 0x235, 0x236, 0x237, 0x238, 0x239, 0x23a,
	0x23b, 0x23c, 0x23d, 0x23e, 0x23f, 0x240, 0x241, 0x242,

	0x244, 0x245, 0x246, 0x247, 0x248, 0x249, 0x24a, 0x24b, // 9
	0x24c, 0x24d, 0x24e, 0x24f, 0x250, 0x251, 0x252, 0x253,
	0x254, 0x256, 0x257, 0x258, 0x259, 0x25a, 0x25b, 0x25c,
	0x25d, 0x25e, 0x25f, 0x260, 0x262, 0x263, 0x264, 0x265,

	0x266, 0x267, 0x268, 0x269, 0x26a, 0x26c, 0x26d, 0x26e, // 10
	0x26f, 0x270, 0x271, 0x272, 0x273, 0x275, 0x276, 0x277,
	0x278, 0x279, 0x27a, 0x27b, 0x27d, 0x27e, 0x27f, 0x280,
	0x281, 0x282, 0x284, 0x285, 0x286, 0x287, 0x288, 0x289,

	0x28b, 0x28c, 0x28d, 0x28e, 0x28f, 0x290, 0x292, 0x293, // 11
	0x294, 0x295, 0x296, 0x298, 0x299, 0x29a, 0x29b, 0x29c,
	0x29e, 0x29f, 0x2a0, 0x2a1, 0x2a2, 0x2a4, 0x2a5, 0x2a6,
	0x2a7, 0x2a9, 0x2aa, 0x2ab, 0x2ac, 0x2ae, 0x2af, 0x2b0,

	0x2b1, 0x2b2, 0x2b4, 0x2b5, 0x2b6, 0x2b7, 0x2b9, 0x2ba, // 12
	0x2bb, 0x2bd, 0x2be, 0x2bf, 0x2c0, 0x2c2, 0x2c3, 0x2c4,
	0x2c5, 0x2c7, 0x2c8, 0x2c9, 0x2cb, 0x2cc, 0x2cd, 0x2ce,
	0x2d0, 0x2d1, 0x2d2, 0x2d4, 0x2d5, 0x2d6, 0x2d8, 0x2d9,

	0x2da, 0x2dc, 0x2dd, 0x2de, 0x2e0, 0x2e1, 0x2e2, 0x2e4, // 13
	0x2e5, 0x2e6, 0x2e8, 0x2e9, 0x2ea, 0x2ec, 0x2ed, 0x2ee,
	0x2f0, 0x2f1, 0x2f2, 0x2f4, 0x2f5, 0x2f6, 0x2f8, 0x2f9,
	0x2fb, 0x2fc, 0x2fd, 0x2ff, 0x300, 0x302, 0x303, 0x304,

	0x306, 0x307, 0x309, 0x30a, 0x30b, 0x30d, 0x30e, 0x310, // 14
	0x311, 0x312, 0x314, 0x315, 0x317, 0x318, 0x31a, 0x31b,
	0x31c, 0x31e, 0x31f, 0x321, 0x322, 0x324, 0x325, 0x327,
	0x328, 0x329, 0x32b, 0x32c, 0x32e, 0x32f, 0x331, 0x332,

	0x334, 0x335, 0x337, 0x338, 0x33a, 0x33b, 0x33d, 0x33e, // 15
	0x340, 0x341, 0x343, 0x344, 0x346, 0x347, 0x349, 0x34a,
	0x34c, 0x34d, 0x34f, 0x350, 0x352, 0x353, 0x355, 0x357,
	0x358, 0x35a, 0x35b, 0x35d, 0x35e, 0x360, 0x361, 0x363,

	0x365, 0x366, 0x368, 0x369, 0x36b, 0x36c, 0x36e, 0x370, // 16
	0x371, 0x373, 0x374, 0x376, 0x378, 0x379, 0x37b, 0x37c,
	0x37e, 0x380, 0x381, 0x383, 0x384, 0x386, 0x388, 0x389,
	0x38b, 0x38d, 0x38e, 0x390, 0x392, 0x393, 0x395, 0x397,

	0x398, 0x39a, 0x39c, 0x39d, 0x39f, 0x3a1, 0x3a2, 0x3a4, // 17
	0x3a6, 0x3a7, 0x3a9, 0x3ab, 0x3ac, 0x3ae, 0x3b0, 0x3b1,
	0x3b3, 0x3b5, 0x3b7, 0x3b8, 0x3ba, 0x3bc, 0x3bd, 0x3bf,
	0x3c1, 0x3c3, 0x3c4, 0x3c6, 0x3c8, 0x3ca, 0x3cb, 0x3cd,

	// The last note has an incomplete range, and loops round back to
	// the start.  Note that the last value is actually a buffer overrun
	// and does not fit with the other values.

	0x3cf, 0x3d1, 0x3d2, 0x3d4, 0x3d6, 0x3d8, 0x3da, 0x3db, // 18
	0x3dd, 0x3df, 0x3e1, 0x3e3, 0x3e4, 0x3e6, 0x3e8, 0x3ea,
	0x3ec, 0x3ed, 0x3ef, 0x3f1, 0x3f3, 0x3f5, 0x3f6, 0x3f8,
	0x3fa, 0x3fc, 0x3fe, 0x36c,
}

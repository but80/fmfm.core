package fmfm

import (
	"fmt"
	"math"
	"time"

	"github.com/but80/fmfm/ymf"
	"github.com/but80/fmfm/ymf/ymfdata"
	"github.com/but80/smaf825/pb/smaf"
)

const (
	flagSustain = 0x02
	flagVibrato = 0x04
	flagFree    = 0x80
)

const modThresh = 40

const (
	ccBankMSB      = 0
	ccModulation   = 1
	ccDataEntryHi  = 6
	ccVolume       = 7
	ccPan          = 10
	ccExpression   = 11
	ccBankLSB      = 32
	ccDataEntryLo  = 38
	ccSustainPedal = 64
	ccSoftPedal    = 67
	ccReverb       = 91
	ccChorus       = 93
	ccNRPNLo       = 98
	ccNRPNHi       = 99
	ccRPNLo        = 100
	ccRPNHi        = 101
	ccSoundsOff    = 120
	ccNotesOff     = 123
	ccMono         = 126
	ccPoly         = 127
)

type slot struct {
	channel    int
	note       int
	realnote   int
	flags      int
	finetune   int
	pitch      int
	velocity   int
	instrument *smaf.VM35VoicePC
	time       time.Time
}

type channelState struct {
	bankLSB    uint8
	bankMSB    uint8
	pc         uint8
	volume     uint8
	expression uint8
	pan        uint8
	pitch      int8
	sustain    uint8
	modulation uint8
	pitchSens  uint16
	rpn        uint16
}

// Controller は、MIDIに類似するインタフェースで Chip のレジスタをコントロールします。
type Controller struct {
	chip    *ymf.Chip
	libraries []*smaf.VM5VoiceLib

	channelStates [16]*channelState
	slots         [ymfdata.ChannelCount]*slot
}

// NewController は、新しい Controller を作成します。
func NewController(chip *ymf.Chip, libraries []*smaf.VM5VoiceLib) *Controller {
	ctrl := &Controller{
		chip:    chip,
		libraries: libraries,
	}
	for i := range ctrl.slots {
		ctrl.slots[i] = &slot{}
	}
	for i := range ctrl.channelStates {
		ctrl.channelStates[i] = &channelState{}
	}
	return ctrl
}

// NoteOn は、MIDIノートオン受信時の音源の振る舞いを再現します。
func (ctrl *Controller) NoteOn(ch, note, velocity int) {
	if velocity == 0 {
		ctrl.NoteOff(ch, note)
		return
	}

	instr, ok := ctrl.getInstrument(ch, note)
	if !ok {
		// TODO: warning
		return
	}

	if instr.VoiceType != smaf.VoiceType_FM {
		fmt.Printf("unsupported voice type: @%d-%d-%d note=%d type=%s\n", instr.BankMsb, instr.BankLsb, instr.Pc, note, instr.VoiceType)
		return
	}

	slotID := ctrl.findFreeSlot(ch, note)
	if 0 <= slotID {
		ctrl.occupySlot(slotID, ch, note, velocity, instr)
	} else {
		// TODO: warning
	}
}

// NoteOff は、MIDIノートオフ受信時の音源の振る舞いを再現します。
func (ctrl *Controller) NoteOff(ch, note int) {
	sus := ctrl.channelStates[ch].sustain
	for slotID, slot := range ctrl.slots {
		if slot.channel == ch && slot.note == note {
			if sus < 0x40 {
				ctrl.releaseSlot(slotID, false)
			} else {
				slot.flags |= flagSustain
			}
		}
	}
}

// ControlChange は、MIDIコントロールチェンジ受信時の音源の振る舞いを再現します。
func (ctrl *Controller) ControlChange(ch, cc, value int) {
	ctrl.ymfChangeControl(ch, cc, value)
}

// ProgramChange は、MIDIプログラムチェンジ受信時の音源の振る舞いを再現します。
func (ctrl *Controller) ProgramChange(ch, value int) {
	ctrl.ymfProgramChange(ch, value)
}

// PitchBend は、MIDIピッチベンド受信時の音源の振る舞いを再現します。
func (ctrl *Controller) PitchBend(ch, l, h int) {
	ctrl.ymfPitchWheel(ch, h*128+l)
}

// Reset は、音源の状態をリセットします。
func (ctrl *Controller) Reset() {
	for _, slot := range ctrl.slots {
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
	for _, state := range ctrl.channelStates {
		state.volume = 100
		state.pan = 64
	}
	ctrl.ymfShutup()
	ctrl.ymfStopMusic()
	ctrl.ymfPlayMusic()
}

func (ctrl *Controller) writeModulation(slot int, instr *smaf.VM35VoicePC, state bool) {
	fmvoice := instr.FmVoice

	// TODO: モジュレータではevbだけを見る(stateは無視)？
	o := fmvoice.Operators
	ctrl.ymfWriteSlotEachOps(
		ymf.OpRegisters.EVB,
		slot,
		bool2int(o[0].Evb || state),
		bool2int(o[1].Evb || state),
		bool2int(o[2].Evb || state),
		bool2int(o[3].Evb || state),
	)
}

func (ctrl *Controller) occupySlot(slotID, channel, note, velocity int, instr *smaf.VM35VoicePC) {
	state := ctrl.channelStates[channel]
	slot := ctrl.slots[slotID]
	slot.channel = channel
	slot.note = note
	slot.flags = 0
	if modThresh <= state.modulation {
		slot.flags |= flagVibrato
	}
	slot.time = time.Now()

	slot.velocity = velocity
	slot.finetune = 0
	if instr.DrumNote != 0 {
		note = int(instr.FmVoice.DrumKey)
	}
	slot.pitch = slot.finetune + int(state.pitch)
	slot.instrument = instr
	if instr.DrumNote == 0 {
		// for note < 0 {
		// 	note += 12
		// }
		// for 127 < note {
		// 	note -= 12
		// }
	}
	note += 2 - 12
	slot.realnote = note

	ctrl.ymfWriteInstrument(slotID, instr)
	ctrl.writeModulation(slotID, instr, slot.flags&flagVibrato != 0)
	ctrl.chip.WriteChannel(ymf.ChRegisters.CHPAN, slotID, int(ctrl.channelStates[channel].pan))
	ctrl.chip.WriteChannel(ymf.ChRegisters.VOLUME, slotID, int(ctrl.channelStates[channel].volume))
	ctrl.chip.WriteChannel(ymf.ChRegisters.EXPRESSION, slotID, int(ctrl.channelStates[channel].expression))
	ctrl.ymfWriteVelocity(slotID, slot.velocity, instr)
	ctrl.writeFrequency(slotID, note, slot.pitch, true)
}

func (ctrl *Controller) releaseSlot(slotID int, killed bool) {
	slot := ctrl.slots[slotID]
	ctrl.writeFrequency(slotID, slot.realnote, slot.pitch, false)
	slot.channel = -1
	slot.time = time.Now()
	slot.flags = flagFree
	if killed {
		ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.SL, slotID, 0)
		ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.RR, slotID, 15) // release rate - fastest
		ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.KSL, slotID, 0)
		ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.TL, slotID, 0x3f) // no volume
	}
}

func (ctrl *Controller) releaseSustain(channel int) {
	for i, slot := range ctrl.slots {
		if slot.channel == channel && slot.flags&flagSustain != 0 {
			ctrl.releaseSlot(i, false)
		}
	}
}

func (ctrl *Controller) findFreeSlot(channel, note int) int {
	for i := 0; i < len(ctrl.slots); i++ {
		if ctrl.slots[i].flags&flagFree != 0 {
			return i
		}
	}

	oldest := -1
	oldesttime := time.Now()

	// find some 2nd-voice channel and determine the oldest
	for i := 0; i < len(ctrl.slots); i++ {
		if ctrl.slots[i].time.Before(oldesttime) {
			oldesttime = ctrl.slots[i].time
			oldest = i
		}
	}

	// if possible, kill the oldest channel
	if 0 <= oldest {
		ctrl.releaseSlot(oldest, true)
		return oldest
	}

	// can't find any free channel
	return -1
}

func (ctrl *Controller) getInstrument(channel, note int) (*smaf.VM35VoicePC, bool) {
	// TODO: smaf825側で検索
	// TODO: ドラム音色
	s := ctrl.channelStates[channel]
	for _, lib := range ctrl.libraries {
		for _, p := range lib.Programs {
			if !(p.Pc == uint32(s.pc) && p.BankLsb == uint32(s.bankLSB) && p.BankMsb == uint32(s.bankMSB)) {
				continue
			}
			if p.DrumNote != 0 && int(p.DrumNote) != note {
				continue
			}
			return p, true
		}
	}
	fmt.Printf("voice not found: @%d-%d-%d note=%d\n", s.bankMSB, s.bankLSB, s.pc, note)

	// TODO: Remove
	if s.bankMSB == 125 && s.pc != 1 {
		s.pc = 1
		return ctrl.getInstrument(channel, note)
	}

	return ctrl.libraries[0].Programs[0], false
}

func (ctrl *Controller) ymfPitchWheel(channel, pitch int) {
	pitch = int(float64(pitch-8192)*float64(ctrl.channelStates[channel].pitchSens)/(200*128) + 64)
	ctrl.channelStates[channel].pitch = int8(pitch)
	for i, slot := range ctrl.slots {
		if slot.channel == channel {
			slot.time = time.Now()
			slot.pitch = slot.finetune + pitch
			ctrl.writeFrequency(i, slot.realnote, slot.pitch, true)
		}
	}
}

func (ctrl *Controller) ymfChangeControl(channel int, controller int, value int) {
	switch controller {
	case ccBankMSB:
		ctrl.channelStates[channel].bankMSB = uint8(value)
	case ccBankLSB:
		ctrl.channelStates[channel].bankLSB = uint8(value)
	case ccModulation:
		ctrl.channelStates[channel].modulation = uint8(value)
		for i, slot := range ctrl.slots {
			if slot.channel == channel {
				flags := slot.flags
				slot.time = time.Now()
				if modThresh <= value {
					slot.flags |= flagVibrato
					if slot.flags != flags {
						ctrl.writeModulation(i, slot.instrument, true)
					}
				} else {
					slot.flags &= ^flagVibrato
					if slot.flags != flags {
						ctrl.writeModulation(i, slot.instrument, false)
					}
				}
			}
		}

	case ccVolume: // change volume
		ctrl.channelStates[channel].volume = uint8(value)
		for i, slot := range ctrl.slots {
			if slot.channel == channel {
				slot.time = time.Now()
				ctrl.chip.WriteChannel(ymf.ChRegisters.VOLUME, i, value)
			}
		}

	case ccExpression: // change expression
		ctrl.channelStates[channel].expression = uint8(value)
		for i, slot := range ctrl.slots {
			if slot.channel == channel {
				slot.time = time.Now()
				ctrl.chip.WriteChannel(ymf.ChRegisters.EXPRESSION, i, value)
			}
		}

	case ccPan: // change pan (balance)
		ctrl.channelStates[channel].pan = uint8(value)
		for i, slot := range ctrl.slots {
			if slot.channel == channel {
				slot.time = time.Now()
				ctrl.chip.WriteChannel(ymf.ChRegisters.CHPAN, i, value)
			}
		}

	case ccSustainPedal: // change sustain pedal (hold)
		ctrl.channelStates[channel].sustain = uint8(value)
		if value < 0x40 {
			ctrl.releaseSustain(channel)
		}

	case ccNotesOff: // turn off all notes that are not sustained
		for i, slot := range ctrl.slots {
			if slot.channel == channel {
				if ctrl.channelStates[channel].sustain < 0x40 {
					ctrl.releaseSlot(i, false)
				} else {
					slot.flags |= flagSustain
				}
			}
		}

	case ccSoundsOff: // release all notes for this channel
		for i, slot := range ctrl.slots {
			if slot.channel == channel {
				ctrl.releaseSlot(i, false)
			}
		}

	case ccRPNHi:
		ctrl.channelStates[channel].rpn = (ctrl.channelStates[channel].rpn & 0x007f) | (uint16(value) << 7)

	case ccRPNLo:
		ctrl.channelStates[channel].rpn = (ctrl.channelStates[channel].rpn & 0x3f80) | uint16(value)

	case ccNRPNLo, ccNRPNHi:
		ctrl.channelStates[channel].rpn = 0x3fff

	case ccDataEntryHi:
		if ctrl.channelStates[channel].rpn == 0 {
			ctrl.channelStates[channel].pitchSens = uint16(value)*100 + (ctrl.channelStates[channel].pitchSens % 100)
		}

	case ccDataEntryLo:
		if ctrl.channelStates[channel].rpn == 0 {
			ctrl.channelStates[channel].pitchSens = uint16(value) + uint16(ctrl.channelStates[channel].pitchSens/100)*100
		}
	}
}

func (ctrl *Controller) ymfProgramChange(channel, value int) {
	ctrl.channelStates[channel].pc = uint8(value)
}

func (ctrl *Controller) ymfResetControllers(channel int) {
	ctrl.channelStates[channel].volume = 100
	ctrl.channelStates[channel].expression = 127
	ctrl.channelStates[channel].sustain = 0
	ctrl.channelStates[channel].pitch = 64
	ctrl.channelStates[channel].rpn = 0x3fff
	ctrl.channelStates[channel].pitchSens = 200
}

func (ctrl *Controller) ymfPlayMusic() {
	for i := range ctrl.slots {
		ctrl.ymfResetControllers(i)
	}
}

func (ctrl *Controller) ymfStopMusic() {
	for i := range ctrl.slots {
		if ctrl.slots[i].flags&flagFree == 0 {
			ctrl.releaseSlot(i, true)
		}
	}
}

func (ctrl *Controller) ymfWriteSlotAllOps(regbase ymf.OpRegister, slotID, data int) {
	ctrl.chip.WriteOperator(regbase, slotID, 0, data)
	ctrl.chip.WriteOperator(regbase, slotID, 1, data)
	ctrl.chip.WriteOperator(regbase, slotID, 2, data)
	ctrl.chip.WriteOperator(regbase, slotID, 3, data)
}

func (ctrl *Controller) ymfWriteSlotEachOps(regbase ymf.OpRegister, slotID, data1, data2, data3, data4 int) {
	ctrl.chip.WriteOperator(regbase, slotID, 0, data1)
	ctrl.chip.WriteOperator(regbase, slotID, 1, data2)
	ctrl.chip.WriteOperator(regbase, slotID, 2, data3)
	ctrl.chip.WriteOperator(regbase, slotID, 3, data4)
}

func (ctrl *Controller) writeFrequency(slotID, note, pitch int, keyon bool) {
	n := float64(note-ymfdata.A3Note) + float64(pitch-64)/32.0
	freq := ymfdata.A3Freq * math.Pow(2.0, n/12.0)

	block := note / 12
	if 7 < block {
		block = 7
	}

	fnum := int(freq*ymfdata.FNUMCoef) >> uint(block-1)
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

	ctrl.chip.WriteChannel(ymf.ChRegisters.FNUM, slotID, fnum)
	ctrl.chip.WriteChannel(ymf.ChRegisters.BLOCK, slotID, block)
	k := 0
	if keyon {
		k = 1
	}
	ctrl.chip.WriteChannel(ymf.ChRegisters.KON, slotID, k)
}

func ymfConvertVelocity(data, velocity int) int {
	r := int(velocitytable[velocity])
	return 0x3f - ((0x3f - data) * r >> 7)
}

func (ctrl *Controller) ymfWriteVelocity(slotID, velocity int, instr *smaf.VM35VoicePC) {
	ops := ctrl.chip.Channels[slotID].Operators
	for i, op := range instr.FmVoice.Operators {
		v := int(op.Tl)
		if !ops[i].IsModulator {
			v = ymfConvertVelocity(v, velocity)
		}
		ctrl.chip.WriteOperator(ymf.OpRegisters.TL, slotID, i, v)
	}
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (ctrl *Controller) ymfWriteInstrument(slotID int, instr *smaf.VM35VoicePC) {
	ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.TL, slotID, 0x3f) // no volume

	for i, op := range instr.FmVoice.Operators {
		ctrl.chip.WriteOperator(ymf.OpRegisters.EAM, slotID, i, bool2int(op.Eam))
		ctrl.chip.WriteOperator(ymf.OpRegisters.EVB, slotID, i, bool2int(op.Evb))
		ctrl.chip.WriteOperator(ymf.OpRegisters.DAM, slotID, i, int(op.Dam))
		ctrl.chip.WriteOperator(ymf.OpRegisters.DVB, slotID, i, int(op.Dvb))
		ctrl.chip.WriteOperator(ymf.OpRegisters.DT, slotID, i, int(op.Dt))
		ctrl.chip.WriteOperator(ymf.OpRegisters.KSL, slotID, i, int(op.Ksl))
		ctrl.chip.WriteOperator(ymf.OpRegisters.KSR, slotID, i, bool2int(op.Ksr))
		ctrl.chip.WriteOperator(ymf.OpRegisters.WS, slotID, i, int(op.Ws))
		ctrl.chip.WriteOperator(ymf.OpRegisters.MULT, slotID, i, int(op.Multi))
		ctrl.chip.WriteOperator(ymf.OpRegisters.FB, slotID, i, int(op.Fb))
		ctrl.chip.WriteOperator(ymf.OpRegisters.AR, slotID, i, int(op.Ar))
		ctrl.chip.WriteOperator(ymf.OpRegisters.DR, slotID, i, int(op.Dr))
		ctrl.chip.WriteOperator(ymf.OpRegisters.SL, slotID, i, int(op.Sl))
		ctrl.chip.WriteOperator(ymf.OpRegisters.SR, slotID, i, int(op.Sr))
		ctrl.chip.WriteOperator(ymf.OpRegisters.RR, slotID, i, int(op.Rr))
		ctrl.chip.WriteOperator(ymf.OpRegisters.TL, slotID, i, int(op.Tl))
		ctrl.chip.WriteOperator(ymf.OpRegisters.XOF, slotID, i, bool2int(op.Xof))
	}

	ctrl.chip.WriteChannel(ymf.ChRegisters.ALG, slotID, int(instr.FmVoice.Alg))
	ctrl.chip.WriteChannel(ymf.ChRegisters.LFO, slotID, int(instr.FmVoice.Lfo))
	ctrl.chip.WriteChannel(ymf.ChRegisters.PANPOT, slotID, int(instr.FmVoice.Panpot))
	ctrl.chip.WriteChannel(ymf.ChRegisters.BO, slotID, int(instr.FmVoice.Bo))
}

func (ctrl *Controller) ymfShutup() {
	for i := range ctrl.slots {
		ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.KSL, i, 0)
		ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.TL, i, 0x3f) // turn off volume
		ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.AR, i, 15)   // the fastest attack,
		ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.DR, i, 15)   // decay
		ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.SL, i, 0)    //
		ctrl.ymfWriteSlotAllOps(ymf.OpRegisters.RR, i, 15)   // ... and release
		ctrl.chip.WriteChannel(ymf.ChRegisters.KON, i, 0)    // KEY-OFF
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

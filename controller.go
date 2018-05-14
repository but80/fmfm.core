package fmfm

import (
	"fmt"
	"math"
	"time"

	"github.com/but80/fmfm/ymf"
	"github.com/but80/fmfm/ymf/ymfdata"
	"github.com/but80/smaf825/pb/smaf"
)

type flag int

const (
	flagSustain  flag = 0x02
	flagVibrato  flag = 0x04
	flagReleased flag = 0x40
	flagFree     flag = 0x80
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

type chipChannelState struct {
	midiChannel int
	note        int
	realnote    int
	flags       flag
	finetune    int
	pitch       int
	instrument  *smaf.VM35VoicePC
	time        time.Time
	minRR       int
}

type midiChannelState struct {
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
	registers ymf.Registers
	libraries []*smaf.VM5VoiceLib

	midiChannelStates [16]*midiChannelState
	chipChannelStates [ymfdata.ChannelCount]*chipChannelState
}

// NewController は、新しい Controller を作成します。
func NewController(registers ymf.Registers, libraries []*smaf.VM5VoiceLib) *Controller {
	ctrl := &Controller{
		registers: registers,
		libraries: libraries,
	}
	for i := range ctrl.chipChannelStates {
		ctrl.chipChannelStates[i] = &chipChannelState{}
	}
	for i := range ctrl.midiChannelStates {
		ctrl.midiChannelStates[i] = &midiChannelState{}
	}
	return ctrl
}

// NoteOn は、MIDIノートオン受信時の音源の振る舞いを再現します。
func (ctrl *Controller) NoteOn(midich, note, velocity int) {
	if velocity == 0 {
		ctrl.NoteOff(midich, note)
		return
	}

	instr, ok := ctrl.getInstrument(midich, note)
	if !ok {
		// TODO: warning
		return
	}

	if instr.VoiceType != smaf.VoiceType_FM {
		fmt.Printf("unsupported voice type: @%d-%d-%d note=%d type=%s\n", instr.BankMsb, instr.BankLsb, instr.Pc, note, instr.VoiceType)
		return
	}

	chipch := ctrl.findFreeChipChannel(midich, note)
	if 0 <= chipch {
		ctrl.occupyChipChannel(chipch, midich, note, velocity, instr)
	} else {
		fmt.Printf("no free chip channel for MIDI channel #%d\n", midich)
	}
}

// NoteOff は、MIDIノートオフ受信時の音源の振る舞いを再現します。
func (ctrl *Controller) NoteOff(ch, note int) {
	sus := ctrl.midiChannelStates[ch].sustain
	for chipch, state := range ctrl.chipChannelStates {
		if state.midiChannel == ch && state.note == note {
			if sus < 0x40 {
				ctrl.releaseChipChannel(chipch, false)
			} else {
				state.flags |= flagSustain
			}
		}
	}
}

// ControlChange は、MIDIコントロールチェンジ受信時の音源の振る舞いを再現します。
func (ctrl *Controller) ControlChange(midich, cc, value int) {
	switch cc {
	case ccBankMSB:
		ctrl.midiChannelStates[midich].bankMSB = uint8(value)
	case ccBankLSB:
		ctrl.midiChannelStates[midich].bankLSB = uint8(value)
	case ccModulation:
		ctrl.midiChannelStates[midich].modulation = uint8(value)
		for i, state := range ctrl.chipChannelStates {
			if state.midiChannel == midich {
				flags := state.flags
				state.time = time.Now()
				if modThresh <= value {
					state.flags |= flagVibrato
					if state.flags != flags {
						ctrl.writeModulation(i, state.instrument, true)
					}
				} else {
					state.flags &= ^flagVibrato
					if state.flags != flags {
						ctrl.writeModulation(i, state.instrument, false)
					}
				}
			}
		}

	case ccVolume: // change volume
		ctrl.midiChannelStates[midich].volume = uint8(value)
		for i, state := range ctrl.chipChannelStates {
			if state.midiChannel == midich {
				state.time = time.Now()
				ctrl.registers.WriteChannel(i, ymf.VOLUME, value)
			}
		}

	case ccExpression: // change expression
		ctrl.midiChannelStates[midich].expression = uint8(value)
		for i, state := range ctrl.chipChannelStates {
			if state.midiChannel == midich {
				state.time = time.Now()
				ctrl.registers.WriteChannel(i, ymf.EXPRESSION, value)
			}
		}

	case ccPan: // change pan (balance)
		ctrl.midiChannelStates[midich].pan = uint8(value)
		for i, state := range ctrl.chipChannelStates {
			if state.midiChannel == midich {
				state.time = time.Now()
				ctrl.registers.WriteChannel(i, ymf.CHPAN, value)
			}
		}

	case ccSustainPedal: // change sustain pedal (hold)
		ctrl.midiChannelStates[midich].sustain = uint8(value)
		if value < 0x40 {
			ctrl.releaseSustain(midich)
		}

	case ccNotesOff: // turn off all notes that are not sustained
		for i, state := range ctrl.chipChannelStates {
			if state.midiChannel == midich {
				if ctrl.midiChannelStates[midich].sustain < 0x40 {
					ctrl.releaseChipChannel(i, false)
				} else {
					state.flags |= flagSustain
				}
			}
		}

	case ccSoundsOff: // release all notes for this channel
		for i, state := range ctrl.chipChannelStates {
			if state.midiChannel == midich {
				ctrl.releaseChipChannel(i, false)
			}
		}

	case ccRPNHi:
		ctrl.midiChannelStates[midich].rpn = (ctrl.midiChannelStates[midich].rpn & 0x007f) | (uint16(value) << 7)

	case ccRPNLo:
		ctrl.midiChannelStates[midich].rpn = (ctrl.midiChannelStates[midich].rpn & 0x3f80) | uint16(value)

	case ccNRPNLo, ccNRPNHi:
		ctrl.midiChannelStates[midich].rpn = 0x3fff

	case ccDataEntryHi:
		if ctrl.midiChannelStates[midich].rpn == 0 {
			ctrl.midiChannelStates[midich].pitchSens = uint16(value)*100 + (ctrl.midiChannelStates[midich].pitchSens % 100)
		}

	case ccDataEntryLo:
		if ctrl.midiChannelStates[midich].rpn == 0 {
			ctrl.midiChannelStates[midich].pitchSens = uint16(value) + uint16(ctrl.midiChannelStates[midich].pitchSens/100)*100
		}
	}
}

// ProgramChange は、MIDIプログラムチェンジ受信時の音源の振る舞いを再現します。
func (ctrl *Controller) ProgramChange(midich, pc int) {
	ctrl.midiChannelStates[midich].pc = uint8(pc)
}

// PitchBend は、MIDIピッチベンド受信時の音源の振る舞いを再現します。
func (ctrl *Controller) PitchBend(midich, l, h int) {
	pitch := h*128 + l - 8192
	pitch = int(float64(pitch)*float64(ctrl.midiChannelStates[midich].pitchSens)/(200*128) + 64)
	ctrl.midiChannelStates[midich].pitch = int8(pitch)
	for i, state := range ctrl.chipChannelStates {
		if state.midiChannel == midich {
			state.time = time.Now()
			state.pitch = state.finetune + pitch
			ctrl.writeFrequency(i, midich, state.realnote, state.pitch, true)
		}
	}
}

// Reset は、音源の状態をリセットします。
func (ctrl *Controller) Reset() {
	for _, state := range ctrl.chipChannelStates {
		state.midiChannel = -1
		state.note = 0
		state.flags = 0
		state.realnote = 0
		state.finetune = 0
		state.pitch = 0
		state.instrument = nil
		state.time = time.Time{}
		state.minRR = 15
	}
	for _, state := range ctrl.midiChannelStates {
		state.volume = 100
		state.pan = 64
	}
	ctrl.muteAllChipChannels()
	ctrl.releaseAllChipChannels()
	ctrl.resetAllMIDIChannels()
}

func (ctrl *Controller) writeModulation(chipch int, instr *smaf.VM35VoicePC, state bool) {
	// TODO: モジュレータではevbだけを見る(stateは無視)？
	for i, o := range instr.FmVoice.Operators {
		ctrl.registers.WriteOperator(chipch, i, ymf.EVB, bool2int(o.Evb || state))
	}
}

func (ctrl *Controller) occupyChipChannel(chipch, midich, note, velocity int, instr *smaf.VM35VoicePC) {
	midiState := ctrl.midiChannelStates[midich]
	chipState := ctrl.chipChannelStates[chipch]
	chipState.midiChannel = midich
	chipState.note = note
	chipState.flags = 0
	if modThresh <= midiState.modulation {
		chipState.flags |= flagVibrato
	}
	chipState.time = time.Now()

	chipState.finetune = 0
	if instr.DrumNote != 0 {
		note = int(instr.FmVoice.DrumKey)
	}
	chipState.pitch = chipState.finetune + int(midiState.pitch)
	chipState.instrument = instr
	if instr.DrumNote == 0 {
		// for note < 0 {
		// 	note += 12
		// }
		// for 127 < note {
		// 	note -= 12
		// }
	}
	chipState.realnote = note

	chipState.minRR = 15
	for i, op := range instr.FmVoice.Operators {
		isCarrier := ymfdata.CarrierMatrix[instr.FmVoice.Alg][i]
		if isCarrier && int(op.Rr) < chipState.minRR {
			chipState.minRR = int(op.Rr)
		}
	}

	ctrl.writeInstrument(chipch, midich, instr)
	ctrl.writeModulation(chipch, instr, chipState.flags&flagVibrato != 0)
	ctrl.registers.WriteChannel(chipch, ymf.CHPAN, int(ctrl.midiChannelStates[midich].pan))
	ctrl.registers.WriteChannel(chipch, ymf.VOLUME, int(ctrl.midiChannelStates[midich].volume))
	// if midich != 4 {
	// 	ctrl.registers.WriteChannel(chipch, ymf.VOLUME, 0)
	// }
	ctrl.registers.WriteChannel(chipch, ymf.EXPRESSION, int(ctrl.midiChannelStates[midich].expression))
	ctrl.registers.WriteChannel(chipch, ymf.VELOCITY, velocity)
	ctrl.writeFrequency(chipch, midich, note, chipState.pitch, true)
}

func (ctrl *Controller) releaseChipChannel(chipch int, killed bool) {
	state := ctrl.chipChannelStates[chipch]
	ctrl.writeFrequency(chipch, -1, state.realnote, state.pitch, false)
	state.midiChannel = -1
	state.time = time.Now()
	state.flags = flagReleased
	if killed {
		ctrl.writeAllOperators(ymf.SL, chipch, 0)
		ctrl.writeAllOperators(ymf.RR, chipch, 15) // release rate - fastest
		ctrl.writeAllOperators(ymf.KSL, chipch, 0)
		ctrl.writeAllOperators(ymf.TL, chipch, 0x3f) // no volume
		state.flags |= flagFree
	}
	ctrl.registers.WriteChannel(chipch, ymf.KON, 0)
}

func (ctrl *Controller) releaseSustain(midich int) {
	for i, state := range ctrl.chipChannelStates {
		if state.midiChannel == midich && state.flags&flagSustain != 0 {
			ctrl.releaseChipChannel(i, false)
		}
	}
}

func (ctrl *Controller) findFreeChipChannel(midich, note int) int {
	for i, state := range ctrl.chipChannelStates {
		if state.flags&flagFree != 0 {
			return i
		}
	}

	now := time.Now()
	foundReleased := -1
	foundTotal := -1
	maxDeltaReleased := -1
	maxDeltaTotal := -1

	for i, state := range ctrl.chipChannelStates {
		delta := int(now.Sub(state.time)) * state.minRR
		if maxDeltaReleased < delta && state.flags&flagReleased != 0 {
			maxDeltaReleased = delta
			foundReleased = i
		}
		if maxDeltaTotal < delta {
			maxDeltaTotal = delta
			foundTotal = i
		}
	}

	if 0 <= foundReleased {
		ctrl.releaseChipChannel(foundReleased, true)
		return foundReleased
	}
	if 0 <= foundTotal {
		ctrl.releaseChipChannel(foundTotal, true)
		return foundTotal
	}

	// can't find any free channel
	return -1
}

func (ctrl *Controller) getInstrument(midich, note int) (*smaf.VM35VoicePC, bool) {
	// TODO: smaf825側で検索
	// TODO: ドラム音色
	s := ctrl.midiChannelStates[midich]
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
	// fmt.Printf("voice not found: @%d-%d-%d note=%d\n", s.bankMSB, s.bankLSB, s.pc, note)

	return ctrl.libraries[0].Programs[0], false
}

func (ctrl *Controller) resetMIDIChannel(midich int) {
	ctrl.midiChannelStates[midich].volume = 100
	ctrl.midiChannelStates[midich].expression = 127
	ctrl.midiChannelStates[midich].sustain = 0
	ctrl.midiChannelStates[midich].pitch = 64
	ctrl.midiChannelStates[midich].rpn = 0x3fff
	ctrl.midiChannelStates[midich].pitchSens = 200
}

func (ctrl *Controller) resetAllMIDIChannels() {
	for i := range ctrl.midiChannelStates {
		ctrl.resetMIDIChannel(i)
	}
}

func (ctrl *Controller) releaseAllChipChannels() {
	for i := range ctrl.chipChannelStates {
		if ctrl.chipChannelStates[i].flags&flagFree == 0 {
			ctrl.releaseChipChannel(i, true)
		}
	}
}

func (ctrl *Controller) writeAllOperators(regbase ymf.OpRegister, chipch, data int) {
	ctrl.registers.WriteOperator(chipch, 0, regbase, data)
	ctrl.registers.WriteOperator(chipch, 1, regbase, data)
	ctrl.registers.WriteOperator(chipch, 2, regbase, data)
	ctrl.registers.WriteOperator(chipch, 3, regbase, data)
}

func (ctrl *Controller) writeFrequency(chipch, midich, note, pitch int, keyon bool) {
	n := float64(note-ymfdata.A3Note) + float64(pitch-64)/32.0
	freq := ymfdata.A3Freq * math.Pow(2.0, n/12.0)

	block := note / 12
	if 7 < block {
		block = 7
	}

	fnum := int(freq*ymfdata.FNUMCoef) << 1 >> uint(block)
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

	ctrl.registers.WriteChannel(chipch, ymf.FNUM, fnum)
	ctrl.registers.WriteChannel(chipch, ymf.BLOCK, block)
	k := 0
	if keyon {
		k = 1
	}
	ctrl.registers.WriteChannel(chipch, ymf.KON, k)
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (ctrl *Controller) writeInstrument(chipch, midich int, instr *smaf.VM35VoicePC) {
	ctrl.writeAllOperators(ymf.TL, chipch, 0x3f) // no volume

	for i, op := range instr.FmVoice.Operators {
		ctrl.registers.WriteOperator(chipch, i, ymf.EAM, bool2int(op.Eam))
		ctrl.registers.WriteOperator(chipch, i, ymf.EVB, bool2int(op.Evb))
		ctrl.registers.WriteOperator(chipch, i, ymf.DAM, int(op.Dam))
		ctrl.registers.WriteOperator(chipch, i, ymf.DVB, int(op.Dvb))
		ctrl.registers.WriteOperator(chipch, i, ymf.DT, int(op.Dt))
		ctrl.registers.WriteOperator(chipch, i, ymf.KSL, int(op.Ksl))
		ctrl.registers.WriteOperator(chipch, i, ymf.KSR, bool2int(op.Ksr))
		ctrl.registers.WriteOperator(chipch, i, ymf.WS, int(op.Ws))
		ctrl.registers.WriteOperator(chipch, i, ymf.MULT, int(op.Multi))
		ctrl.registers.WriteOperator(chipch, i, ymf.FB, int(op.Fb))
		ctrl.registers.WriteOperator(chipch, i, ymf.AR, int(op.Ar))
		ctrl.registers.WriteOperator(chipch, i, ymf.DR, int(op.Dr))
		ctrl.registers.WriteOperator(chipch, i, ymf.SL, int(op.Sl))
		ctrl.registers.WriteOperator(chipch, i, ymf.SR, int(op.Sr))
		ctrl.registers.WriteOperator(chipch, i, ymf.RR, int(op.Rr))
		ctrl.registers.WriteOperator(chipch, i, ymf.TL, int(op.Tl))
		ctrl.registers.WriteOperator(chipch, i, ymf.XOF, bool2int(op.Xof))
	}

	ctrl.registers.WriteChannel(chipch, ymf.ALG, int(instr.FmVoice.Alg))
	ctrl.registers.WriteChannel(chipch, ymf.LFO, int(instr.FmVoice.Lfo))
	ctrl.registers.WriteChannel(chipch, ymf.PANPOT, int(instr.FmVoice.Panpot))
	ctrl.registers.WriteChannel(chipch, ymf.BO, int(instr.FmVoice.Bo))
}

func (ctrl *Controller) muteAllChipChannels() {
	for i := range ctrl.chipChannelStates {
		ctrl.writeAllOperators(ymf.KSL, i, 0)
		ctrl.writeAllOperators(ymf.TL, i, 0x3f)    // turn off volume
		ctrl.writeAllOperators(ymf.AR, i, 15)      // the fastest attack,
		ctrl.writeAllOperators(ymf.DR, i, 15)      // decay
		ctrl.writeAllOperators(ymf.SL, i, 0)       //
		ctrl.writeAllOperators(ymf.RR, i, 15)      // ... and release
		ctrl.registers.WriteChannel(i, ymf.KON, 0) // KEY-OFF
	}
}

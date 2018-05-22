package fmfm

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/but80/fmfm.core/ymf"
	"github.com/but80/fmfm.core/ymf/ymfdata"
	"github.com/but80/go-smaf/pb/smaf"
)

var defaultPC = &smaf.VM35VoicePC{
	Version:   smaf.VM35FMVoiceVersion_VM5,
	Name:      "default",
	VoiceType: smaf.VoiceType_FM,
	FmVoice: &smaf.VM35FMVoice{
		Panpot: 15,
		Bo:     1,
		Alg:    1,
		Operators: []*smaf.VM35FMOperator{
			{},
			{},
		},
	},
}

// MIDIMessage は、MIDIメッセージの種類を表す列挙子型です。
type MIDIMessage int

const (
	// MIDINoteOn は、MIDIメッセージの種類 NoteOn を表す列挙子です。
	MIDINoteOn MIDIMessage = iota + 1
	// MIDINoteOff は、MIDIメッセージの種類 NoteOff を表す列挙子です。
	MIDINoteOff
	// MIDIControlChange は、MIDIメッセージの種類 ControlChange を表す列挙子です。
	MIDIControlChange
	// MIDIProgramChange は、MIDIメッセージの種類 ProgramChange を表す列挙子です。
	MIDIProgramChange
	// MIDIPitchBend は、MIDIメッセージの種類 PitchBend を表す列挙子です。
	MIDIPitchBend
)

type midiMessage struct {
	typ         MIDIMessage
	timestamp   int
	midiChannel int
	data1       int
	data2       int
}

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
	// ccSoftPedal    = 67
	// ccReverb       = 91
	// ccChorus       = 93
	ccNRPNLo    = 98
	ccNRPNHi    = 99
	ccRPNLo     = 100
	ccRPNHi     = 101
	ccSoundsOff = 120
	ccNotesOff  = 123
	ccMono      = 126
	ccPoly      = 127
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
	bankLSB             uint8
	bankMSB             uint8
	pc                  uint8
	volume              uint8
	expression          uint8
	pan                 uint8
	pitch               int8
	sustain             uint8
	modulation          uint8
	pitchSens           uint16
	rpn                 uint16
	mono                bool
	debugLastInstrument *smaf.VM35VoicePC
}

// ControllerOpts は、 NewController のオプションです。
type ControllerOpts struct {
	Registers          ymf.Registers
	Libraries          []*smaf.VM5VoiceLib
	ForceMono          bool
	PrintStatus        bool
	IgnoreMIDIChannels []int
	SoloMIDIChannel    int
}

// Controller は、MIDIに類似するインタフェースで Chip のレジスタをコントロールします。
type Controller struct {
	mutex              sync.Mutex
	registers          ymf.Registers
	libraries          []*smaf.VM5VoiceLib
	forceMono          bool
	debugPrintStatus   bool
	ignoreMIDIChannels map[int]struct{}
	soloMIDIChannel    int
	midiMessages       []*midiMessage

	midiChannelStates [16]*midiChannelState
	chipChannelStates [ymfdata.ChannelCount]*chipChannelState
}

// NewController は、新しい Controller を作成します。
func NewController(opts *ControllerOpts) *Controller {
	ctrl := &Controller{
		registers:          opts.Registers,
		libraries:          opts.Libraries,
		forceMono:          opts.ForceMono,
		debugPrintStatus:   opts.PrintStatus,
		ignoreMIDIChannels: map[int]struct{}{},
		soloMIDIChannel:    opts.SoloMIDIChannel,
		midiMessages:       []*midiMessage{},
	}
	for _, ch := range opts.IgnoreMIDIChannels {
		ctrl.ignoreMIDIChannels[ch] = struct{}{}
	}
	for i := range ctrl.chipChannelStates {
		ctrl.chipChannelStates[i] = &chipChannelState{}
	}
	for i := range ctrl.midiChannelStates {
		ctrl.midiChannelStates[i] = &midiChannelState{}
	}
	ctrl.Reset()
	return ctrl
}

// PushMIDIMessage は、処理すべきMIDIメッセージを追加します。
func (ctrl *Controller) PushMIDIMessage(typ MIDIMessage, timestamp, midich, data1, data2 int) {
	ctrl.mutex.Lock()
	defer ctrl.mutex.Unlock()

	ctrl.midiMessages = append(ctrl.midiMessages, &midiMessage{
		typ:         typ,
		timestamp:   timestamp,
		midiChannel: midich,
		data1:       data1,
		data2:       data2,
	})
}

var lastPrintedAt = time.Time{}

// FlushMIDIMessages は、蓄積されたMIDIメッセージを処理します。
func (ctrl *Controller) FlushMIDIMessages(until int) {
	ctrl.mutex.Lock()
	defer ctrl.mutex.Unlock()

	rest := []*midiMessage{}
	for _, msg := range ctrl.midiMessages {
		if until < msg.timestamp {
			rest = append(rest, msg)
			continue
		}
		// fmt.Printf("%02d: %d\n", msg.midiChannel, until - msg.timestamp)
		switch msg.typ {
		case MIDINoteOn:
			ctrl.noteOn(msg.midiChannel, msg.data1, msg.data2)
		case MIDINoteOff:
			ctrl.noteOff(msg.midiChannel, msg.data1)
		case MIDIControlChange:
			ctrl.controlChange(msg.midiChannel, msg.data1, msg.data2)
		case MIDIProgramChange:
			ctrl.programChange(msg.midiChannel, msg.data1)
		case MIDIPitchBend:
			ctrl.pitchBend(msg.midiChannel, msg.data1, msg.data2)
		}
	}
	ctrl.midiMessages = rest

	if ctrl.debugPrintStatus {
		now := time.Now()
		if time.Millisecond*10 <= now.Sub(lastPrintedAt) {
			ctrl.printStatus()
			lastPrintedAt = now
		}
	}
}

var notes = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func (ctrl *Controller) printStatus() {
	fmt.Println("")
	fmt.Println("Ch MSB-LSB-@PC Instrument       P Vol Exp Pan Vo Note")
	for i, ms := range ctrl.midiChannelStates {
		cs := &chipChannelState{}
		voices := 0
		lastTime := time.Time{}
		for _, s := range ctrl.chipChannelStates {
			if s.midiChannel == i {
				voices++
				if s.time.After(lastTime) {
					cs = s
					lastTime = s.time
				}
			}
		}

		monopoly := "-"
		note := ""
		instr := ms.debugLastInstrument
		pc := "-----------"
		if instr == nil || instr == defaultPC {
			instr = &smaf.VM35VoicePC{}
		} else {
			if ms.mono {
				monopoly = "M"
			} else {
				monopoly = "P"
			}
			pc = fmt.Sprintf("%03d-%03d-%03d", ms.bankMSB, ms.bankLSB, ms.pc)
			note = fmt.Sprintf("%s%d", notes[cs.note%12], cs.note/12-2)
		}
		if voices == 0 {
			note = ""
		}
		fmt.Printf(
			"%2d %s %-16s %s %3d %3d %3d %2d %-4s\n",
			i+1,
			pc,
			instr.Name,
			monopoly,
			ms.volume,
			ms.expression,
			ms.pan,
			voices,
			note,
		)
	}
}

// noteOn は、MIDIノートオン受信時の音源の振る舞いを再現します。
func (ctrl *Controller) noteOn(midich, note, velocity int) {
	if velocity == 0 {
		ctrl.noteOff(midich, note)
		return
	}
	if _, ok := ctrl.ignoreMIDIChannels[midich]; ok {
		return
	}

	instr, ok := ctrl.getInstrument(midich, note)
	if !ok {
		// TODO: warning or ignore by option
		// return
	}

	if instr.VoiceType != smaf.VoiceType_FM {
		fmt.Printf("unsupported voice type: @%d-%d-%d note=%d type=%s\n", instr.BankMsb, instr.BankLsb, instr.Pc, note, instr.VoiceType)
		return
	}

	var chipch = -1
	if ctrl.midiChannelStates[midich].mono || ctrl.forceMono && instr.DrumNote == 0 {
		chipch = ctrl.findLastUsedChipChannel(midich, note)
	}
	if chipch < 0 {
		chipch = ctrl.findFreeChipChannel(midich, note)
	}
	if 0 <= chipch {
		ctrl.occupyChipChannel(chipch, midich, note, velocity, instr)
	} else {
		fmt.Printf("no free chip channel for MIDI channel #%d\n", midich)
	}
}

// noteOff は、MIDIノートオフ受信時の音源の振る舞いを再現します。
func (ctrl *Controller) noteOff(midich, note int) {
	if _, ok := ctrl.ignoreMIDIChannels[midich]; ok {
		return
	}

	sus := ctrl.midiChannelStates[midich].sustain
	for chipch, state := range ctrl.chipChannelStates {
		if state.midiChannel == midich && state.note == note {
			if sus < 0x40 {
				ctrl.keyOff(chipch)
			} else {
				state.flags |= flagSustain
			}
		}
	}
}

// controlChange は、MIDIコントロールチェンジ受信時の音源の振る舞いを再現します。
func (ctrl *Controller) controlChange(midich, cc, value int) {
	if _, ok := ctrl.ignoreMIDIChannels[midich]; ok {
		return
	}
	channel := ctrl.midiChannelStates[midich]

	switch cc {
	case ccBankMSB:
		channel.bankMSB = uint8(value)
	case ccBankLSB:
		channel.bankLSB = uint8(value)
	case ccModulation:
		channel.modulation = uint8(value)
		for i, state := range ctrl.chipChannelStates {
			if state.midiChannel == midich {
				flags := state.flags
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
		channel.volume = uint8(value)
		if ctrl.soloMIDIChannel < 0 || midich == ctrl.soloMIDIChannel {
			ctrl.writeChannelsUsingMIDIChannel(midich, ymf.VOLUME, value)
		}

	case ccExpression: // change expression
		channel.expression = uint8(value)
		ctrl.writeChannelsUsingMIDIChannel(midich, ymf.EXPRESSION, value)

	case ccPan: // change pan (balance)
		channel.pan = uint8(value)
		ctrl.writeChannelsUsingMIDIChannel(midich, ymf.CHPAN, value)

	case ccSustainPedal: // change sustain pedal (hold)
		channel.sustain = uint8(value)
		if value < 0x40 {
			ctrl.releaseSustain(midich)
		}

	case ccMono:
		channel.mono = true

	case ccPoly:
		channel.mono = false

	case ccNotesOff: // turn off all notes that are not sustained
		for i, state := range ctrl.chipChannelStates {
			if state.midiChannel == midich {
				if channel.sustain < 0x40 {
					ctrl.keyOff(i)
				} else {
					state.flags |= flagSustain
				}
			}
		}

	case ccSoundsOff: // release all notes for this channel
		for i, state := range ctrl.chipChannelStates {
			if state.midiChannel == midich {
				ctrl.keyOff(i)
			}
		}

	case ccRPNHi:
		channel.rpn = (channel.rpn & 0x007f) | (uint16(value) << 7)

	case ccRPNLo:
		channel.rpn = (channel.rpn & 0x3f80) | uint16(value)

	case ccNRPNLo, ccNRPNHi:
		channel.rpn = 0x3fff

	case ccDataEntryHi:
		if channel.rpn == 0 {
			channel.pitchSens = uint16(value)*100 + (channel.pitchSens % 100)
		}

	case ccDataEntryLo:
		if channel.rpn == 0 {
			channel.pitchSens = uint16(value) + uint16(channel.pitchSens/100)*100
		}
	}
}

// programChange は、MIDIプログラムチェンジ受信時の音源の振る舞いを再現します。
func (ctrl *Controller) programChange(midich, pc int) {
	if _, ok := ctrl.ignoreMIDIChannels[midich]; ok {
		return
	}
	ctrl.midiChannelStates[midich].pc = uint8(pc)
}

// pitchBend は、MIDIピッチベンド受信時の音源の振る舞いを再現します。
func (ctrl *Controller) pitchBend(midich, l, h int) {
	if _, ok := ctrl.ignoreMIDIChannels[midich]; ok {
		return
	}

	pitch := h*128 + l - 8192
	pitch = int(float64(pitch)*float64(ctrl.midiChannelStates[midich].pitchSens)/(200*128) + 64)
	ctrl.midiChannelStates[midich].pitch = int8(pitch)
	for i, state := range ctrl.chipChannelStates {
		if state.midiChannel == midich {
			state.pitch = state.finetune + pitch
			ctrl.writeFrequency(i, state.realnote, state.pitch)
		}
	}
}

// Reset は、音源の状態をリセットします。
func (ctrl *Controller) Reset() {
	ctrl.mutex.Lock()
	defer ctrl.mutex.Unlock()

	for i := range ctrl.chipChannelStates {
		ctrl.resetChipChannel(i)
	}
	for i := range ctrl.midiChannelStates {
		ctrl.resetMIDIChannel(i)
	}
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
	midiState.debugLastInstrument = instr
	chipState.realnote = note

	chipState.minRR = 15
	for i, op := range instr.FmVoice.Operators {
		isCarrier := ymfdata.CarrierMatrix[instr.FmVoice.Alg][i]
		if isCarrier && int(op.Rr) < chipState.minRR {
			chipState.minRR = int(op.Rr)
		}
	}

	ctrl.writeInstrument(chipch, instr)
	ctrl.writeModulation(chipch, instr, chipState.flags&flagVibrato != 0)
	ctrl.registers.WriteChannel(chipch, ymf.CHPAN, int(ctrl.midiChannelStates[midich].pan))
	if ctrl.soloMIDIChannel < 0 || midich == ctrl.soloMIDIChannel {
		ctrl.registers.WriteChannel(chipch, ymf.VOLUME, int(ctrl.midiChannelStates[midich].volume))
	} else {
		ctrl.registers.WriteChannel(chipch, ymf.VOLUME, 0)
	}
	ctrl.registers.WriteChannel(chipch, ymf.EXPRESSION, int(ctrl.midiChannelStates[midich].expression))
	ctrl.registers.WriteChannel(chipch, ymf.VELOCITY, velocity)
	ctrl.writeFrequency(chipch, note, chipState.pitch)
	ctrl.keyOn(chipch, midich)
}

func (ctrl *Controller) resetChipChannel(chipch int) {
	state := ctrl.chipChannelStates[chipch]
	state.time = time.Time{}
	state.flags = flagReleased | flagFree
	state.minRR = 15
	state.instrument = nil
	state.midiChannel = -1
	// state.note = 0
	// state.realnote = 0
	// state.finetune = 0
	// state.pitch = 0
	ctrl.registers.WriteChannel(chipch, ymf.RESET, 1)
}

func (ctrl *Controller) releaseSustain(midich int) {
	for i, state := range ctrl.chipChannelStates {
		if state.midiChannel == midich && state.flags&flagSustain != 0 {
			ctrl.keyOff(i)
		}
	}
}

// findLastUsedChipChannel は、指定MIDIチャンネルの指定ノートを発音するとき、
// MONOモード時に収容先となるチップのチャンネルを選択します。
func (ctrl *Controller) findLastUsedChipChannel(midich, note int) int {
	now := time.Now()
	found := -1
	minDelta := math.MaxInt64
	for i, state := range ctrl.chipChannelStates {
		if state.midiChannel != midich {
			continue
		}
		if state.note == note {
			return i
		}
		delta := int(now.Sub(state.time)) * state.minRR
		if delta < minDelta {
			minDelta = delta
			found = i
		}
	}
	if 0 <= found {
		return found
	}
	return -1
}

// findLastUsedChipChannel は、指定MIDIチャンネルの指定ノートを発音するとき、
// POLYモード時に収容先となるチップのチャンネルを選択します。
func (ctrl *Controller) findFreeChipChannel(midich, note int) int {
	// // 同じノートで発音済みのチャンネルがあれば最優先で選択
	// for i, state := range ctrl.chipChannelStates {
	// 	if state.midiChannel == midich && state.note == note {
	// 		return i
	// 	}
	// }

	// 無音のチャンネルがあれば選択
	for i, state := range ctrl.chipChannelStates {
		if state.flags&flagFree != 0 {
			return i
		}
	}

	now := time.Now()
	foundTotal := -1
	foundReleased := -1
	maxDeltaTotal := -1
	maxAttenuationReleased := -1
	for i, state := range ctrl.chipChannelStates {
		delta := int(now.Sub(state.time))
		if maxDeltaTotal < delta {
			maxDeltaTotal = delta
			foundTotal = i
		}
		// -dB ∝ 2^RR * time
		attenuation := delta * int(1<<uint(state.minRR))
		if maxAttenuationReleased < attenuation && state.flags&flagReleased != 0 {
			maxAttenuationReleased = attenuation
			foundReleased = i
		}
	}

	// リリース後に最も減衰していると思われるチャンネルを選択
	if 0 <= foundReleased {
		ctrl.resetChipChannel(foundReleased)
		return foundReleased
	}
	// 未リリースだが最も古くなったと思われるチャンネルを選択
	if 0 <= foundTotal {
		ctrl.resetChipChannel(foundTotal)
		return foundTotal
	}

	// 収容先がない
	return -1
}

func (ctrl *Controller) getInstrument(midich, note int) (*smaf.VM35VoicePC, bool) {
	if ctrl.libraries == nil {
		return defaultPC, false
	}

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

	return defaultPC, false
}

func (ctrl *Controller) resetMIDIChannel(midich int) {
	ctrl.midiChannelStates[midich].volume = 100
	ctrl.midiChannelStates[midich].expression = 127
	ctrl.midiChannelStates[midich].pan = 64
	ctrl.midiChannelStates[midich].sustain = 0
	ctrl.midiChannelStates[midich].pitch = 64
	ctrl.midiChannelStates[midich].rpn = 0x3fff
	ctrl.midiChannelStates[midich].pitchSens = 200
}

func (ctrl *Controller) writeChannelsUsingMIDIChannel(midich int, regbase ymf.ChRegister, value int) {
	for i, state := range ctrl.chipChannelStates {
		if state.midiChannel == midich {
			ctrl.registers.WriteChannel(i, regbase, value)
		}
	}
}

func (ctrl *Controller) writeAllOperators(chipch int, regbase ymf.OpRegister, value int) {
	ctrl.registers.WriteOperator(chipch, 0, regbase, value)
	ctrl.registers.WriteOperator(chipch, 1, regbase, value)
	ctrl.registers.WriteOperator(chipch, 2, regbase, value)
	ctrl.registers.WriteOperator(chipch, 3, regbase, value)
}

func (ctrl *Controller) writeFrequency(chipch, note, pitch int) {
	n := float64(note-ymfdata.A3Note) + float64(pitch-64)/32.0
	freq := ymfdata.A3Freq * math.Pow(2.0, n/12.0)

	block := (note + 3 - 12) / 12
	if block < 0 {
		block = 0
	} else if 7 < block {
		block = 7
	}

	fnumF64 := freq * ymfdata.FNUMCoef
	blockUint := uint(block)
	fnum := int(fnumF64*2.0+float64(uint(1)<<blockUint>>1)) >> blockUint
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
}

func (ctrl *Controller) keyOn(chipch, midich int) {
	ctrl.registers.DebugSetMIDIChannel(chipch, midich)
	ctrl.registers.WriteChannel(chipch, ymf.KON, 1)
}

func (ctrl *Controller) keyOff(chipch int) {
	state := ctrl.chipChannelStates[chipch]
	state.time = time.Now()
	state.flags = flagReleased
	ctrl.registers.WriteChannel(chipch, ymf.KON, 0)
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (ctrl *Controller) writeInstrument(chipch int, instr *smaf.VM35VoicePC) {
	ctrl.writeAllOperators(chipch, ymf.TL, 0x3f) // no volume

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

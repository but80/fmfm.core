package main

import (
	"C"
	"sync"

	"github.com/but80/fmfm.core"
	"github.com/but80/fmfm.core/sim"
	"github.com/but80/smaf825/pb/smaf"
)
import (
	"io/ioutil"
	"strings"

	"github.com/gogo/protobuf/proto"
)

func main() {
	// noop
}

var lib *smaf.VM5VoiceLib
var chip *sim.Chip
var ctrl *fmfm.Controller
var initOnce sync.Once

// FMFMInit は、音源を初期化します。
//export FMFMInit
func FMFMInit(sampleRate C.double, voicePath *C.char) C.int {
	result := 0
	initOnce.Do(func() {
		info, err := ioutil.ReadDir("voice")
		if err != nil {
			panic(err)
		}
		libs := []*smaf.VM5VoiceLib{}
		for _, i := range info {
			if i.IsDir() || !strings.HasSuffix(i.Name(), ".vm5.pb") {
				continue
			}
			b, err := ioutil.ReadFile("voice/" + i.Name())
			if err != nil {
				panic(err)
			}
			var lib smaf.VM5VoiceLib
			err = proto.Unmarshal(b, &lib)
			if err != nil {
				panic(err)
			}
			libs = append(libs, &lib)
		}

		chip = sim.NewChip(float64(sampleRate), -15.0, -1)
		regs := sim.NewRegisters(chip)
		opts := &fmfm.ControllerOpts{
			Registers: regs,
			Libraries: libs,
		}
		ctrl = fmfm.NewController(opts)
		result = 1
	})
	return C.int(result)
}

// FMFMFlushMIDIMessages は、蓄積されたMIDIメッセージを処理します。
func FMFMFlushMIDIMessages(until int) {
	ctrl.FlushMIDIMessages(until)
}

// FMFMNoteOn は、MIDIノートオン受信時の音源の振る舞いを再現します。
//export FMFMNoteOn
func FMFMNoteOn(timestamp, ch, note, velocity C.longlong) {
	ctrl.PushMIDIMessage(fmfm.MIDINoteOn, int(timestamp), int(ch), int(note), int(velocity))
}

// FMFMNoteOff は、MIDIノートオフ受信時の音源の振る舞いを再現します。
//export FMFMNoteOff
func FMFMNoteOff(timestamp, ch, note C.longlong) {
	ctrl.PushMIDIMessage(fmfm.MIDINoteOff, int(timestamp), int(ch), int(note), 0)
}

// FMFMControlChange は、MIDIコントロールチェンジ受信時の音源の振る舞いを再現します。
//export FMFMControlChange
func FMFMControlChange(timestamp, ch, cc, value C.longlong) {
	ctrl.PushMIDIMessage(fmfm.MIDIControlChange, int(timestamp), int(ch), int(cc), int(value))
}

// FMFMProgramChange は、MIDIプログラムチェンジ受信時の音源の振る舞いを再現します。
//export FMFMProgramChange
func FMFMProgramChange(timestamp, ch, value C.longlong) {
	ctrl.PushMIDIMessage(fmfm.MIDIProgramChange, int(timestamp), int(ch), int(value), 0)
}

// FMFMPitchBend は、MIDIピッチベンド受信時の音源の振る舞いを再現します。
//export FMFMPitchBend
func FMFMPitchBend(timestamp, ch, l, h C.longlong) {
	ctrl.PushMIDIMessage(fmfm.MIDIPitchBend, int(timestamp), int(ch), int(l), int(h))
}

// FMFMNext は、次のサンプルを生成・取得します。
//export FMFMNext
func FMFMNext() (C.double, C.double) {
	l, r := chip.Next()
	return C.double(l), C.double(r)
}

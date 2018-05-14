package main

import (
	"C"
	"sync"

	"github.com/but80/fmfm.core"
	"github.com/but80/fmfm.core/ymf"
	"github.com/but80/smaf825/pb/smaf"
)

func main() {
	// noop
}

var lib *smaf.VM5VoiceLib
var chip *ymf.Chip
var ctrl *fmfm.Controller
var initOnce sync.Once

// FMFMInit は、音源を初期化します。
//export FMFMInit
func FMFMInit(sampleRate C.double, voicePath *C.char) C.int {
	result := 0
	initOnce.Do(func() {
		var err error
		lib, err = smaf.NewVM5VoiceLib(C.GoString(voicePath))
		if err != nil {
			panic(err)
		}
		chip = ymf.NewChip(float64(sampleRate))
		ctrl = fmfm.NewController(chip, lib)
		ctrl.Reset()
		result = 1
	})
	return C.int(result)
}

// FMFMNoteOn は、MIDIノートオン受信時の音源の振る舞いを再現します。
//export FMFMNoteOn
func FMFMNoteOn(ch, note, velocity C.longlong) {
	ctrl.NoteOn(int(ch), int(note), int(velocity))
}

// FMFMNoteOff は、MIDIノートオフ受信時の音源の振る舞いを再現します。
//export FMFMNoteOff
func FMFMNoteOff(ch, note C.longlong) {
	ctrl.NoteOff(int(ch), int(note))
}

// FMFMControlChange は、MIDIコントロールチェンジ受信時の音源の振る舞いを再現します。
//export FMFMControlChange
func FMFMControlChange(ch, cc, value C.longlong) {
	ctrl.ControlChange(int(ch), int(cc), int(value))
}

// FMFMProgramChange は、MIDIプログラムチェンジ受信時の音源の振る舞いを再現します。
//export FMFMProgramChange
func FMFMProgramChange(ch, value C.longlong) {
	ctrl.ProgramChange(int(ch), int(value))
}

// FMFMPitchBend は、MIDIピッチベンド受信時の音源の振る舞いを再現します。
//export FMFMPitchBend
func FMFMPitchBend(ch, l, h C.longlong) {
	ctrl.PitchBend(int(ch), int(l), int(h))
}

// FMFMNext は、次のサンプルを生成・取得します。
//export FMFMNext
func FMFMNext() (C.double, C.double) {
	l, r := chip.Next()
	return C.double(l), C.double(r)
}

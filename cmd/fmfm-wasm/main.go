package main

import (
	"sync"
	"syscall/js"

	fmfm "github.com/but80/fmfm.core"
	"github.com/but80/fmfm.core/sim"
	"gopkg.in/but80/go-smaf.v1/pb/smaf"
)

var (
	lib      smaf.VM5VoiceLib
	chip     *sim.Chip
	ctrl     *fmfm.Controller
	initOnce sync.Once
	wait     = make(chan struct{})
)

// // fmfmLoadLibrary は、ライブラリをロードします。
// func fmfmLoadLibrary(voicePath *C.char) C.int {
// 	voicePathGo := C.GoString(voicePath)
// 	info, err := ioutil.ReadDir(voicePathGo)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return 0
// 	}
// 	for _, i := range info {
// 		if i.IsDir() || !strings.HasSuffix(i.Name(), ".vm5.pb") {
// 			continue
// 		}
// 		err := lib.LoadFile(voicePathGo + "/" + i.Name())
// 		if err != nil {
// 			fmt.Println(err.Error())
// 			return 0
// 		}
// 	}
// 	return 1
// }

// fmfmInit は、音源を初期化します。
func fmfmInit(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return false
	}
	sampleRate := args[0].Float()
	initOnce.Do(func() {
		chip = sim.NewChip(sampleRate, -15.0, -1)
		regs := sim.NewRegisters(chip)
		opts := &fmfm.ControllerOpts{
			Registers: regs,
			Library:   &lib,
		}
		ctrl = fmfm.NewController(opts)
	})
	return chip.SampleRate() == sampleRate
}

// fmfmNoteOn は、MIDIノートオン受信時の音源の振る舞いを再現します。
func fmfmNoteOn(this js.Value, args []js.Value) interface{} {
	if len(args) < 4 {
		return false
	}
	timestamp := args[0].Int()
	ch := args[1].Int()
	note := args[2].Int()
	velocity := args[3].Int()
	ctrl.PushMIDIMessage(fmfm.MIDINoteOn, timestamp, ch, note, velocity)
	return true
}

// fmfmNoteOff は、MIDIノートオフ受信時の音源の振る舞いを再現します。
func fmfmNoteOff(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return false
	}
	timestamp := args[0].Int()
	ch := args[1].Int()
	note := args[2].Int()
	ctrl.PushMIDIMessage(fmfm.MIDINoteOff, timestamp, ch, note, 0)
	return true
}

// fmfmControlChange は、MIDIコントロールチェンジ受信時の音源の振る舞いを再現します。
func fmfmControlChange(this js.Value, args []js.Value) interface{} {
	if len(args) < 4 {
		return false
	}
	timestamp := args[0].Int()
	ch := args[1].Int()
	cc := args[2].Int()
	value := args[3].Int()
	ctrl.PushMIDIMessage(fmfm.MIDIControlChange, timestamp, ch, cc, value)
	return true
}

// fmfmProgramChange は、MIDIプログラムチェンジ受信時の音源の振る舞いを再現します。
func fmfmProgramChange(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return false
	}
	timestamp := args[0].Int()
	ch := args[1].Int()
	value := args[2].Int()
	ctrl.PushMIDIMessage(fmfm.MIDIProgramChange, timestamp, ch, value, 0)
	return true
}

// fmfmPitchBend は、MIDIピッチベンド受信時の音源の振る舞いを再現します。
func fmfmPitchBend(this js.Value, args []js.Value) interface{} {
	if len(args) < 4 {
		return false
	}
	timestamp := args[0].Int()
	ch := args[1].Int()
	l := args[2].Int()
	h := args[3].Int()
	ctrl.PushMIDIMessage(fmfm.MIDIPitchBend, timestamp, ch, l, h)
	return true
}

// fmfmRender は、サンプルを生成・取得します。
func fmfmRender(this js.Value, args []js.Value) interface{} {
	if len(args) < 4 {
		return false
	}
	outL := args[0]
	outR := args[1]
	size := args[2].Int()
	now := args[3].Float()
	delta := 1000.0 / chip.SampleRate()
	for i := 0; i < size; i++ {
		ctrl.FlushMIDIMessages(int(now))
		now += delta
		l, r := chip.Next()
		outL.SetIndex(i, l)
		outR.SetIndex(i, r)
	}
	return now
}

// fmfmExit は、このサービスを終了します。
func fmfmExit(this js.Value, args []js.Value) interface{} {
	wait <- struct{}{}
	return true
}

func main() {
	// js.Global().Set("fmfmLoadLibrary", js.FuncOf(fmfmLoadLibrary))
	js.Global().Set("fmfmInit", js.FuncOf(fmfmInit))
	js.Global().Set("fmfmNoteOn", js.FuncOf(fmfmNoteOn))
	js.Global().Set("fmfmNoteOff", js.FuncOf(fmfmNoteOff))
	js.Global().Set("fmfmControlChange", js.FuncOf(fmfmControlChange))
	js.Global().Set("fmfmProgramChange", js.FuncOf(fmfmProgramChange))
	js.Global().Set("fmfmPitchBend", js.FuncOf(fmfmPitchBend))
	js.Global().Set("fmfmRender", js.FuncOf(fmfmRender))
	js.Global().Set("fmfmExit", js.FuncOf(fmfmExit))
	<-wait
}

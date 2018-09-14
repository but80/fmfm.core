package main

import "C"

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	fmfm "gopkg.in/but80/fmfm.core.v1"
	"gopkg.in/but80/fmfm.core.v1/sim"
	"gopkg.in/but80/go-smaf.v1/pb/smaf"
)

func main() {
	// noop
}

var lib smaf.VM5VoiceLib
var chip *sim.Chip
var ctrl *fmfm.Controller
var initOnce sync.Once

// FMFMLoadLibrary は、ライブラリをロードします。
//export FMFMLoadLibrary
func FMFMLoadLibrary(voicePath *C.char) C.int {
	voicePathGo := C.GoString(voicePath)
	info, err := ioutil.ReadDir(voicePathGo)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	for _, i := range info {
		if i.IsDir() || !strings.HasSuffix(i.Name(), ".vm5.pb") {
			continue
		}
		err := lib.LoadFile(voicePathGo + "/" + i.Name())
		if err != nil {
			fmt.Println(err.Error())
			return 0
		}
	}
	return 1
}

// FMFMInit は、音源を初期化します。
//export FMFMInit
func FMFMInit(sampleRate C.double) C.int {
	result := 0
	initOnce.Do(func() {
		chip = sim.NewChip(float64(sampleRate), -15.0, -1)
		regs := sim.NewRegisters(chip)
		opts := &fmfm.ControllerOpts{
			Registers: regs,
			Library:   &lib,
		}
		ctrl = fmfm.NewController(opts)
		result = 1
	})
	return C.int(result)
}

// FMFMFlushMIDIMessages は、蓄積されたMIDIメッセージを処理します。
//export FMFMFlushMIDIMessages
func FMFMFlushMIDIMessages(until C.longlong) {
	ctrl.FlushMIDIMessages(int(until))
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

// FMFMListBankMSB は、登録されている音色の選択可能なMSBの一覧を返します。
//export FMFMListBankMSB
func FMFMListBankMSB(out *C.longlong) C.longlong {
	return writeInts(out, collectInts(func(ch chan<- int) {
		for _, p := range lib.Programs {
			ch <- int(p.BankMsb)
		}
	}))
}

// FMFMListBankLSB は、登録されている音色の選択可能なLSBの一覧を返します。
//export FMFMListBankLSB
func FMFMListBankLSB(out *C.longlong, msb C.longlong) C.longlong {
	return writeInts(out, collectInts(func(ch chan<- int) {
		for _, p := range lib.Programs {
			if p.BankMsb == uint32(msb) {
				ch <- int(p.BankLsb)
			}
		}
	}))
}

// FMFMListPC は、登録されている音色の選択可能なプログラムチェンジの一覧を返します。
//export FMFMListPC
func FMFMListPC(out *C.longlong, msb, lsb C.longlong) C.longlong {
	return writeInts(out, collectInts(func(ch chan<- int) {
		for _, p := range lib.Programs {
			if p.BankMsb == uint32(msb) && p.BankLsb == uint32(lsb) {
				ch <- int(p.Pc)
			}
		}
	}))
}

// FMFMListDrumNote は、登録されている音色の選択可能なドラムノートの一覧を返します。
//export FMFMListDrumNote
func FMFMListDrumNote(out *C.longlong, msb, lsb, pc C.longlong) C.longlong {
	return writeInts(out, collectInts(func(ch chan<- int) {
		for _, p := range lib.Programs {
			if p.BankMsb == uint32(msb) && p.BankLsb == uint32(lsb) && p.Pc == uint32(pc) {
				ch <- int(p.DrumNote)
			}
		}
	}))
}

// FMFMGetVoice は、音色データを Protocol Buffers 形式にエンコードして返します。
//export FMFMGetVoice
func FMFMGetVoice(out *C.uchar, msb, lsb, pc, drumNote C.longlong) C.longlong {
	// TODO: implement
	for _, p := range lib.Programs {
		if p.BankMsb == uint32(msb) && p.BankLsb == uint32(lsb) && p.Pc == uint32(pc) {
			data, err := proto.Marshal(p)
			if err != nil {
				fmt.Println(err.Error())
				return 0
			}
			return writeBytes(out, data)
		}
	}
	return 0
}

// FMFMNext は、次のサンプルを生成・取得します。
//export FMFMNext
func FMFMNext() (C.double, C.double) {
	l, r := chip.Next()
	return C.double(l), C.double(r)
}

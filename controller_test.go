package fmfm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/but80/fmfm.core.v1/ymf"
	"gopkg.in/but80/fmfm.core.v1/ymf/ymfdata"
)

type registers struct {
	channels     [ymfdata.ChannelCount]map[ymf.ChRegister]int
	operators    [ymfdata.ChannelCount][4]map[ymf.OpRegister]int
	midiChannels [ymfdata.ChannelCount]int
}

func newRegisters() *registers {
	result := &registers{}
	for i := 0; i < ymfdata.ChannelCount; i++ {
		m := map[ymf.ChRegister]int{}
		m[ymf.PANPOT] = 15
		m[ymf.CHPAN] = 64
		m[ymf.VOLUME] = 100
		m[ymf.EXPRESSION] = 127
		m[ymf.BO] = 1
		result.channels[i] = m
		result.midiChannels[i] = -1
		for j := 0; j < 4; j++ {
			m := map[ymf.OpRegister]int{}
			m[ymf.MULT] = 1
			m[ymf.AR] = 15
			m[ymf.RR] = 15
			result.operators[i][j] = m
		}
	}
	return result
}

// WriteOperator は、オペレータレジスタに値を書き込みます。
func (regs *registers) WriteOperator(channel, operatorIndex int, offset ymf.OpRegister, v int) {
	regs.operators[channel][operatorIndex][offset] = v
}

// WriteTL は、TLレジスタに値を書き込みます。
func (regs *registers) WriteTL(channel, operatorIndex int, tlCarrier, tlModulator int) {
	alg := regs.channels[channel][ymf.ALG]
	for i := 0; i < 4; i++ {
		v := 31
		if ymfdata.CarrierMatrix[alg][i] {
			v = tlCarrier
		} else if ymfdata.ModulatorMatrix[alg][i] {
			v = tlModulator
		}
		regs.operators[channel][operatorIndex][ymf.TL] = v
	}
}

// WriteChannel は、チャンネルレジスタに値を書き込みます。
func (regs *registers) WriteChannel(channel int, offset ymf.ChRegister, v int) {
	regs.channels[channel][offset] = v
}

// DebugSetMIDIChannel は、チャンネルを使用しているMIDIチャンネル番号をデバッグ用にセットします。
func (regs *registers) DebugSetMIDIChannel(channel, midiChannel int) {
	regs.midiChannels[channel] = midiChannel
}

func TestController_writeFrequency(t *testing.T) {
	regs := newRegisters()
	ctrl := NewController(&ControllerOpts{Registers: regs})
	fnumPrev := 300
	for i := 0; i < 12; i++ {
		n := ymfdata.A3Note + i
		ctrl.noteOn(0, n, 127)
		ch := regs.channels[0]
		fnum := ch[ymf.FNUM]
		if i == 0 {
			assert.Equal(t, 300, fnum)
		} else {
			assert.True(t, fnumPrev < fnum)
		}
		fnumPrev = fnum
		// t.Errorf("%d: block=%d bo=%d fnum=%d", n, ch[ymf.BLOCK], ch[ymf.BO], ch[ymf.FNUM])
		ctrl.noteOff(0, n)
		ctrl.resetChipChannel(0)
	}
}

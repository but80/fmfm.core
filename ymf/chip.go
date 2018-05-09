package ymf

import (
	"math"
	"github.com/but80/fmfm/ymf/ymfdata"
)

// Chip は、FM音源チップ全体を表す型です。
type Chip struct {
	// SampleRate は、出力波形の目標サンプルレートです。
	SampleRate float64
	// TotalLevel は、出力のトータルな音量[dB]です。
	TotalLevel float64
	// Channels は、このチップが備える全チャンネルです。
	Channels []*Channel

	registers     Registers
	currentOutput []float64
}

// NewChip は、新しい Chip を作成します。
func NewChip(sampleRate, totalLevel float64) *Chip {
	chip := &Chip{
		SampleRate:    sampleRate,
		TotalLevel:    totalLevel,
		Channels:      make([]*Channel, ymfdata.ChannelCount),
		currentOutput: make([]float64, 2),
	}
	chip.initChannels()
	return chip
}

// Next は、次のサンプルを生成し、その左右それぞれの振幅を返します。
func (chip *Chip) Next() (float64, float64) {
	var l, r float64
	for _, channel := range chip.Channels {
		cl, cr := channel.getChannelOutput()
		l += cl
		r += cr
	}
	v := math.Pow(10, chip.TotalLevel / 20)
	return l*v, r*v
}

// WriteChannel は、チャンネルレジスタに値を書き込みます。
func (chip *Chip) WriteChannel(address ChRegister, channelID, data int) {
	chip.registers.write(int(address)+channelID, data)
	switch address {

	case ChRegisters.KON:
		chip.Channels[channelID].updateKON()

	case ChRegisters.BLOCK:
		chip.Channels[channelID].updateBLOCK()

	case ChRegisters.FNUM:
		chip.Channels[channelID].updateFNUM()

	case ChRegisters.ALG:
		chip.Channels[channelID].updateALG()

	case ChRegisters.LFO:
		chip.Channels[channelID].updateLFO()

	case ChRegisters.PANPOT:
		chip.Channels[channelID].updatePANPOT()

	case ChRegisters.CHPAN:
		chip.Channels[channelID].updateCHPAN()

	case ChRegisters.VOLUME:
		chip.Channels[channelID].updateVOLUME()

	case ChRegisters.EXPRESSION:
		chip.Channels[channelID].updateEXPRESSION()

	case ChRegisters.BO:
		chip.Channels[channelID].updateBO()
	}
}

// WriteOperator は、オペレータレジスタに値を書き込みます。
func (chip *Chip) WriteOperator(address OpRegister, channelID, operatorIndex, data int) {
	chip.registers.writeOperator(channelID, operatorIndex, address, data)
	switch address {
	case OpRegisters.EAM:
		chip.Channels[channelID].Operators[operatorIndex].updateEAM()
	case OpRegisters.EVB:
		chip.Channels[channelID].Operators[operatorIndex].updateEVB()
	case OpRegisters.DAM:
		chip.Channels[channelID].Operators[operatorIndex].updateDAM()
	case OpRegisters.DVB:
		chip.Channels[channelID].Operators[operatorIndex].updateDVB()
	case OpRegisters.DT:
		chip.Channels[channelID].Operators[operatorIndex].updateDT()
	case OpRegisters.KSR:
		chip.Channels[channelID].Operators[operatorIndex].updateKSR()
	case OpRegisters.MULT:
		chip.Channels[channelID].Operators[operatorIndex].updateMULT()
	case OpRegisters.KSL:
		chip.Channels[channelID].Operators[operatorIndex].updateKSL()
	case OpRegisters.TL:
		chip.Channels[channelID].Operators[operatorIndex].updateTL()
	case OpRegisters.AR:
		chip.Channels[channelID].Operators[operatorIndex].updateAR()
	case OpRegisters.DR:
		chip.Channels[channelID].Operators[operatorIndex].updateDR()
	case OpRegisters.SL:
		chip.Channels[channelID].Operators[operatorIndex].updateSL()
	case OpRegisters.SR:
		chip.Channels[channelID].Operators[operatorIndex].updateSR()
	case OpRegisters.RR:
		chip.Channels[channelID].Operators[operatorIndex].updateRR()
	case OpRegisters.XOF:
		chip.Channels[channelID].Operators[operatorIndex].updateXOF()
	case OpRegisters.WS:
		chip.Channels[channelID].Operators[operatorIndex].updateWS()
	case OpRegisters.FB:
		chip.Channels[channelID].Operators[operatorIndex].updateFB()
	}
}

func (chip *Chip) initChannels() {
	chip.Channels = make([]*Channel, ymfdata.ChannelCount)
	for i := range chip.Channels {
		chip.Channels[i] = newChannel4op(i, chip)
	}
}

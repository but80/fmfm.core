package ymf

import "github.com/but80/fmfm/ymf/ymfdata"

type Chip struct {
	registers     Registers
	Channels      []*Channel
	currentOutput []float64
}

func NewChip() *Chip {
	chip := &Chip{
		Channels:      make([]*Channel, ymfdata.CHANNEL_COUNT),
		currentOutput: make([]float64, 2),
	}
	chip.initChannels()
	return chip
}

func (chip *Chip) Next() (float64, float64) {
	var l, r float64
	for _, channel := range chip.Channels {
		cl, cr := channel.getChannelOutput()
		l += cl
		r += cr
	}
	return l, r
}

func (chip *Chip) WriteChannel(address ChRegister, channelID, data int) {
	chip.registers.write(int(address)+channelID, data)
	switch address {

	case ChRegister_KON:
		chip.Channels[channelID].updateKON()

	case ChRegister_BLOCK:
		chip.Channels[channelID].updateBLOCK()

	case ChRegister_FNUM:
		chip.Channels[channelID].updateFNUM()

	case ChRegister_ALG:
		chip.Channels[channelID].updateALG()

	case ChRegister_LFO:
		chip.Channels[channelID].updateLFO()

	case ChRegister_PANPOT:
		chip.Channels[channelID].updatePANPOT()

	case ChRegister_CHPAN:
		chip.Channels[channelID].updateCHPAN()

	case ChRegister_VOLUME:
		chip.Channels[channelID].updateVOLUME()

	case ChRegister_EXPRESSION:
		chip.Channels[channelID].updateEXPRESSION()

	case ChRegister_BO:
		chip.Channels[channelID].updateBO()
	}
}

func (chip *Chip) WriteOperator(address OpRegister, channelID, operatorIndex, data int) {
	chip.registers.writeOperator(channelID, operatorIndex, address, data)
	switch address {
	case OpRegister_EAM:
		chip.Channels[channelID].Operators[operatorIndex].updateEAM()
	case OpRegister_EVB:
		chip.Channels[channelID].Operators[operatorIndex].updateEVB()
	case OpRegister_DAM:
		chip.Channels[channelID].Operators[operatorIndex].updateDAM()
	case OpRegister_DVB:
		chip.Channels[channelID].Operators[operatorIndex].updateDVB()
	case OpRegister_DT:
		chip.Channels[channelID].Operators[operatorIndex].updateDT()
	case OpRegister_KSR:
		chip.Channels[channelID].Operators[operatorIndex].updateKSR()
	case OpRegister_MULT:
		chip.Channels[channelID].Operators[operatorIndex].updateMULT()
	case OpRegister_KSL:
		chip.Channels[channelID].Operators[operatorIndex].updateKSL()
	case OpRegister_TL:
		chip.Channels[channelID].Operators[operatorIndex].updateTL()
	case OpRegister_AR:
		chip.Channels[channelID].Operators[operatorIndex].updateAR()
	case OpRegister_DR:
		chip.Channels[channelID].Operators[operatorIndex].updateDR()
	case OpRegister_SL:
		chip.Channels[channelID].Operators[operatorIndex].updateSL()
	case OpRegister_SR:
		chip.Channels[channelID].Operators[operatorIndex].updateSR()
	case OpRegister_RR:
		chip.Channels[channelID].Operators[operatorIndex].updateRR()
	case OpRegister_XOF:
		chip.Channels[channelID].Operators[operatorIndex].updateXOF()
	case OpRegister_WS:
		chip.Channels[channelID].Operators[operatorIndex].updateWS()
	case OpRegister_FB:
		chip.Channels[channelID].Operators[operatorIndex].updateFB()
	}
}

func (chip *Chip) initChannels() {
	chip.Channels = make([]*Channel, ymfdata.CHANNEL_COUNT)
	for i := range chip.Channels {
		chip.Channels[i] = newChannel4op(i, chip)
	}
}

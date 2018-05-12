package sim

import (
	"github.com/but80/fmfm/ymf"
)

// Registers は、全レジスタのコンテナです。
type Registers struct {
	chip *Chip
}

var _ ymf.Registers = &Registers{}

// NewRegisters は、新しい Registers を作成します。
func NewRegisters(chip *Chip) *Registers {
	return &Registers{
		chip: chip,
	}
}

// WriteOperator は、オペレータレジスタに値を書き込みます。
func (regs *Registers) WriteOperator(channel, operatorIndex int, offset ymf.OpRegister, v int) {
	switch offset {
	case ymf.EAM:
		regs.chip.channels[channel].operators[operatorIndex].setEAM(v)
	case ymf.EVB:
		regs.chip.channels[channel].operators[operatorIndex].setEVB(v)
	case ymf.DAM:
		regs.chip.channels[channel].operators[operatorIndex].setDAM(v)
	case ymf.DVB:
		regs.chip.channels[channel].operators[operatorIndex].setDVB(v)
	case ymf.DT:
		regs.chip.channels[channel].operators[operatorIndex].setDT(v)
	case ymf.KSR:
		regs.chip.channels[channel].operators[operatorIndex].setKSR(v)
	case ymf.MULT:
		regs.chip.channels[channel].operators[operatorIndex].setMULT(v)
	case ymf.KSL:
		regs.chip.channels[channel].operators[operatorIndex].setKSL(v)
	case ymf.TL:
		regs.chip.channels[channel].operators[operatorIndex].setTL(v)
	case ymf.AR:
		regs.chip.channels[channel].operators[operatorIndex].setAR(v)
	case ymf.DR:
		regs.chip.channels[channel].operators[operatorIndex].setDR(v)
	case ymf.SL:
		regs.chip.channels[channel].operators[operatorIndex].setSL(v)
	case ymf.SR:
		regs.chip.channels[channel].operators[operatorIndex].setSR(v)
	case ymf.RR:
		regs.chip.channels[channel].operators[operatorIndex].setRR(v)
	case ymf.XOF:
		regs.chip.channels[channel].operators[operatorIndex].setXOF(v)
	case ymf.WS:
		regs.chip.channels[channel].operators[operatorIndex].setWS(v)
	case ymf.FB:
		regs.chip.channels[channel].operators[operatorIndex].setFB(v)
	}
}

// WriteTL は、TLレジスタに値を書き込みます。
func (regs *Registers) WriteTL(channel, operatorIndex int, tlCarrier, tlModulator int) {
	if regs.chip.channels[channel].operators[operatorIndex].isModulator {
		regs.WriteOperator(channel, operatorIndex, ymf.TL, tlModulator)
	} else {
		regs.WriteOperator(channel, operatorIndex, ymf.TL, tlCarrier)
	}
}

// WriteChannel は、チャンネルレジスタに値を書き込みます。
func (regs *Registers) WriteChannel(channel int, offset ymf.ChRegister, v int) {
	switch offset {
	case ymf.KON:
		regs.chip.channels[channel].setKON(v)
		// regs.chip.channels[channel].midiChannelID = midich
		// if midich == 4 {
		// 	fmt.Print(regs.chip.channels[channel].dump())
		// }
	case ymf.BLOCK:
		regs.chip.channels[channel].setBLOCK(v)
	case ymf.FNUM:
		regs.chip.channels[channel].setFNUM(v)
	case ymf.ALG:
		regs.chip.channels[channel].setALG(v)
	case ymf.LFO:
		regs.chip.channels[channel].setLFO(v)
	case ymf.PANPOT:
		regs.chip.channels[channel].setPANPOT(v)
	case ymf.CHPAN:
		regs.chip.channels[channel].setCHPAN(v)
	case ymf.VOLUME:
		regs.chip.channels[channel].setVOLUME(v)
	case ymf.EXPRESSION:
		regs.chip.channels[channel].setEXPRESSION(v)
	case ymf.VELOCITY:
		regs.chip.channels[channel].setVELOCITY(v)
	case ymf.BO:
		regs.chip.channels[channel].setBO(v)
	}
}

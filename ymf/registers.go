package ymf

import (
	"github.com/but80/fmfm/ymf/ymfdata"
)

const registerSize = 0x500

type OpRegister int

var OpRegisters = struct {
	IDMask OpRegister
	EAM    OpRegister
	EVB    OpRegister
	DAM    OpRegister
	DVB    OpRegister
	DT     OpRegister
	KSL    OpRegister
	KSR    OpRegister
	WS     OpRegister
	MULT   OpRegister
	FB     OpRegister
	AR     OpRegister
	DR     OpRegister
	SL     OpRegister
	SR     OpRegister
	RR     OpRegister
	TL     OpRegister
	XOF    OpRegister
}{
	IDMask: 0x1f,
	EAM:    0xc0,
	EVB:    0x100,
	DAM:    0x140,
	DVB:    0x180,
	DT:     0x1c0,
	KSL:    0x200,
	KSR:    0x240,
	WS:     0x280,
	MULT:   0x2c0,
	FB:     0x300,
	AR:     0x340,
	DR:     0x380,
	SL:     0x3c0,
	SR:     0x400,
	RR:     0x440,
	TL:     0x480,
	XOF:    0x4c0,
}

type ChRegister int

var ChRegisters = struct {
	IDMask     ChRegister
	KON        ChRegister
	BLOCK      ChRegister
	FNUM       ChRegister
	ALG        ChRegister
	LFO        ChRegister
	PANPOT     ChRegister
	CHPAN      ChRegister
	VOLUME     ChRegister
	EXPRESSION ChRegister
	BO         ChRegister
}{
	IDMask:     0x07,
	KON:        0x10,
	BLOCK:      0x20,
	FNUM:       0x30,
	ALG:        0x40,
	LFO:        0x50,
	PANPOT:     0x60,
	CHPAN:      0x70,
	VOLUME:     0x80,
	EXPRESSION: 0x90,
	BO:         0xa0,
}

type Registers struct {
	registers [registerSize]int
}

func (regs *Registers) write(address, data int) {
	regs.registers[address] = data
}

func (regs *Registers) writeOperator(channel, operatorIndex int, offset OpRegister, data int) {
	operatorID := channel + operatorIndex*ymfdata.CHANNEL_COUNT
	regs.registers[operatorID+int(offset)] = data
}

func (regs *Registers) readOperator(channel, operatorIndex int, offset OpRegister) int {
	operatorID := channel + operatorIndex*ymfdata.CHANNEL_COUNT
	return regs.registers[operatorID+int(offset)]
}

func (regs *Registers) writeChannel(channel int, offset ChRegister, data int) {
	regs.registers[channel+int(offset)] = data
}

func (regs *Registers) readChannel(channel int, offset ChRegister) int {
	return regs.registers[channel+int(offset)]
}

package ymf

import (
	"github.com/but80/fmfm/ymf/ymfdata"
)

/*

=============================================
*.vm5

    | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
------------------------------------ Global
+0 |                               |
+1 |                               |  // Drumkey?
+2 |      PANPOT       |       | ? |
+3 |  LFO  |P E|       |    ALG    |
------------------------------------ Op0
+4 |      S R      |XOF| - |SUS|KSR|
+5 |      R R      |      D R      |
+6 |      A R      |      S L      |
+7 |         T L           |  KSL  |
+8 |                               |
+9 | - |  DAM  |EAM| - |  DVB  |EVB|
+A |      MUL      | - |    DT     |
+B |        W S        |    FB     |
------------------------------------ Op1
...

=============================================
OPL3

|------+---+---+---+---+---+---+---+---|
| ADDR | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
|------+---+---+---+---+---+---+---+---| ↓システムグローバル
|01    |           LSI  TEST           |
|02    |            TIMER 1            |
|03    |            TIMER 2            |
|04    |RST|MT1|MT2|     -     |ST2|ST1|
|05    |               -               |
|08    | - |NTS|           -           |
|------+---+---+---+---+---+---+---+---| ↓オペレータ単位
|20..35|A M|VIB|EGT|KSR|     MULT      |
|40..55|  KSL  |          T L          |
|60..75|      A R      |      D R      |
|80..95|      S L      |      R R      |
|------+---+---+---+---+---+---+---+---| ↓チャンネルグローバル
|A0..A8|          F NUMBER (L)         |
|B0..B8|   -   |KON|   BLOCK   |FNUM(H)|
|BD    |DAM|DVB|RYT|B D|S D|TOM|T C|H H|
|C0..C8|CHD|CHC|CHB|CHA|    F B    |CNT|
|E0..F5|         -         |    W S    |
|------+---+---+---+---+---+---+---+---|

=============================================
bit数が増える・名前が異なるパラメータ:

- CNT 2 → ALG 3
- AM  1 → EAM 1
- VIB 1 → EVB 1
- WS がオペレータ単位5bitに
- EGT が 0 なら SR=RR | 1 なら SR=0
- FB がオペレータ単位に
- CHA-CHD → PANPOT
- DAM, DVB がオペレータ単位で2bitに
- DT がオペレータに新設
- LFO, PE がチャンネルに新設

*/

const REGISTER_SIZE = 0x500

type OpRegister int

const (
	OpRegister_OPERATOR_ID_MASK OpRegister = 0x1f
	OpRegister_EAM              OpRegister = 0xc0
	OpRegister_EVB              OpRegister = 0x100
	OpRegister_DAM              OpRegister = 0x140
	OpRegister_DVB              OpRegister = 0x180
	OpRegister_DT               OpRegister = 0x1c0
	OpRegister_KSL              OpRegister = 0x200
	OpRegister_KSR              OpRegister = 0x240
	OpRegister_WS               OpRegister = 0x280
	OpRegister_MULT             OpRegister = 0x2c0
	OpRegister_FB               OpRegister = 0x300
	OpRegister_AR               OpRegister = 0x340
	OpRegister_DR               OpRegister = 0x380
	OpRegister_SL               OpRegister = 0x3c0
	OpRegister_SR               OpRegister = 0x400
	OpRegister_RR               OpRegister = 0x440
	OpRegister_TL               OpRegister = 0x480
	OpRegister_XOF              OpRegister = 0x4c0
)

type ChRegister int

const (
	ChRegister_CHANNEL_ID_MASK ChRegister = 0x07
	ChRegister_KON             ChRegister = 0x10
	ChRegister_BLOCK           ChRegister = 0x20
	ChRegister_FNUM            ChRegister = 0x30
	ChRegister_ALG             ChRegister = 0x40
	ChRegister_LFO             ChRegister = 0x50
	ChRegister_PANPOT          ChRegister = 0x60
	ChRegister_CHPAN           ChRegister = 0x70
	ChRegister_VOLUME          ChRegister = 0x80
	ChRegister_EXPRESSION      ChRegister = 0x90
	ChRegister_BO              ChRegister = 0xa0
)

type Registers struct {
	registers [REGISTER_SIZE]int
}

func (regs *Registers) write(address, data int) {
	regs.registers[address] = data
}

func (regs *Registers) writeOperator(channelID, operatorIndex int, offset OpRegister, data int) {
	operatorID := channelID + operatorIndex*ymfdata.CHANNEL_COUNT
	regs.registers[operatorID+int(offset)] = data
}

func (regs *Registers) readOperator(channelID, operatorIndex int, offset OpRegister) int {
	operatorID := channelID + operatorIndex*ymfdata.CHANNEL_COUNT
	return regs.registers[operatorID+int(offset)]
}

func (regs *Registers) writeChannel(channel int, offset ChRegister, data int) {
	regs.registers[channel+int(offset)] = data
}

func (regs *Registers) readChannel(channel int, offset ChRegister) int {
	return regs.registers[channel+int(offset)]
}

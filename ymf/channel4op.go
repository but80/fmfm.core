package ymf

import (
	"github.com/but80/fmfm/ymf/ymfdata"
)

const noModulator = 0

var IS_MODULATOR = [8][4]bool{
	{true, false, false, false},
	{false, false, false, false},
	{false, false, false, false},
	{true, true, true, false},
	{true, true, true, false},
	{true, false, true, false},
	{false, true, true, false},
	{false, true, false, false},
}

/*

==================================================
MA-5

ALG=0
  (FB)1 -> 2 -> OUT

ALG=1
  (FB)1 -> | -> OUT
      2 -> |

ALG=2
  (FB)1 -> | -> OUT
      2 -> |
  (FB)3 -> |
      4 -> |

ALG=3
  (FB)1 ------> | -> 4 -> OUT
      2 -> 3 -> |

ALG=4
  (FB)1 -> 2 -> 3 -> 4 -> OUT

ALG=5
  (FB)1 -> 2 -> | -> OUT
  (FB)3 -> 4 -> |

ALG=6
  (FB)1 -----------> | -> OUT
      2 -> 3 -> 4 -> |

ALG=7
  (FB)1 ------> | -> OUT
      2 -> 3 -> |
      4 ------> |

==================================================
OPL3

| ADDR | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
|C0..C8|CHD|CHC|CHB|CHA|    F B    |CNT|

===== 2 operators mode =====

CNT = 0
  (FB)OP1 -> OP2 -> OUT

CNT = 1
  (FB)OP1 -> |
      OP2 -> | -> OUT

===== 4 operators mode =====

|Channel No.|  1  |  2  |  3  |  4  |  5  |  6  |
|CNT Address|C0,C3|C1,C4|C2,C5|C0,C3|C1,C4|C2,C5|
|    A1     |       L         |        H        |

CNT(Cn) = 0, CNT(Cn+3) = 0
  (FB)OP1 -> OP2 -> OP3 -> OP4 -> OUT

CNT(Cn) = 0, CNT(Cn+3) = 1
  (FB)OP1 -> OP2 -> |
      OP3 -> OP4 -> | -> OUT

CNT(Cn) = 1, CNT(Cn+3) = 0
  (FB)OP1 ---------------> |
      OP2 -> OP3 -> OP4 -> | -> OUT

CNT(Cn) = 1, CNT(Cn+3) = 1
  (FB)OP1 --------> |
      OP2 -> OP3 -> |
      OP4 --------> | -> OUT
*/

type Channel4op struct {
	channelID int

	chip       *Chip
	fnum       int
	kon        int
	block      int
	alg        int
	lfo        int
	panpot     int
	chpan      int
	volume     int
	expression int
	bo         int

	feedback    [2][2]float64
	feedbackOut [2]float64
	toPhase     float64

	Operators [4]*Operator
}

func newChannel4op(channelID int, chip *Chip) *Channel4op {
	ch := &Channel4op{
		chip:      chip,
		channelID: channelID,

		fnum:       0,
		kon:        0,
		block:      0,
		alg:        0,
		lfo:        0,
		panpot:     15,
		chpan:      64,
		volume:     0,
		expression: 0,
		bo:         1,

		toPhase: 4,
	}
	for i := range ch.Operators {
		ch.Operators[i] = newOperator(channelID, i, chip)
	}
	return ch
}

func (ch *Channel4op) updateKON() {
	newKon := ch.chip.registers.readChannel(ch.channelID, ChRegister_KON)
	if newKon == ch.kon {
		return
	}
	if newKon == 1 {
		ch.keyOn()
	} else {
		ch.keyOff()
	}
	ch.kon = newKon
}

func (ch *Channel4op) updateBLOCK() {
	ch.block = ch.chip.registers.readChannel(ch.channelID, ChRegister_BLOCK)
	ch.updateOperators()
}

func (ch *Channel4op) updateFNUM() {
	ch.fnum = ch.chip.registers.readChannel(ch.channelID, ChRegister_FNUM)
	ch.updateOperators()
}

func (ch *Channel4op) updateALG() {
	ch.alg = ch.chip.registers.readChannel(ch.channelID, ChRegister_ALG)
	ch.updateOperators()
}

func (ch *Channel4op) updateLFO() {
	ch.lfo = ch.chip.registers.readChannel(ch.channelID, ChRegister_LFO)
	ch.updateOperators()
}

func (ch *Channel4op) updatePANPOT() {
	ch.panpot = ch.chip.registers.readChannel(ch.channelID, ChRegister_PANPOT)
}

func (ch *Channel4op) updateCHPAN() {
	ch.chpan = ch.chip.registers.readChannel(ch.channelID, ChRegister_CHPAN)
}

func (ch *Channel4op) updateVOLUME() {
	ch.volume = ch.chip.registers.readChannel(ch.channelID, ChRegister_VOLUME)
}

func (ch *Channel4op) updateEXPRESSION() {
	ch.expression = ch.chip.registers.readChannel(ch.channelID, ChRegister_EXPRESSION)
}

func (ch *Channel4op) updateBO() {
	ch.bo = ch.chip.registers.readChannel(ch.channelID, ChRegister_BO)
	ch.updateOperators()
}

func (ch *Channel4op) updateChannel() {
	ch.updateKON()
	ch.updateBLOCK()
	ch.updateFNUM()
	ch.updateALG()
	ch.updateLFO()
}

func (ch *Channel4op) toStereo(channelOutput float64, volume32, expression32, pan128 int) (float64, float64) {
	channelOutput *= ymfdata.VolumeTable[volume32]
	channelOutput *= ymfdata.VolumeTable[expression32]
	p := ymfdata.PanTable[pan128]
	return channelOutput * p[0], channelOutput * p[1]
}

func (ch *Channel4op) getChannelOutput() (float64, float64) {
	var channelOutput float64
	var op1Output float64
	var op2Output float64
	var op3Output float64
	var op4Output float64

	op1 := ch.Operators[0]
	op2 := ch.Operators[1]
	op3 := ch.Operators[2]
	op4 := ch.Operators[3]

	switch ch.alg {

	case 0:
		// (FB)1 -> 2 -> OUT
		if op2.envelopeGenerator.stage == Stage_OFF {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut[0])

		channelOutput = op2.getOperatorOutput(op1Output * ch.toPhase)

	case 1:
		// (FB)1 -> | -> OUT
		//     2 -> |
		if op1.envelopeGenerator.stage == Stage_OFF && op2.envelopeGenerator.stage == Stage_OFF {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut[0])
		op2Output = op2.getOperatorOutput(noModulator)

		channelOutput = op2Output + op4Output

	case 2:
		// (FB)1 -> | -> OUT
		//     2 -> |
		// (FB)3 -> |
		//     4 -> |
		if op1.envelopeGenerator.stage == Stage_OFF &&
			op2.envelopeGenerator.stage == Stage_OFF &&
			op3.envelopeGenerator.stage == Stage_OFF &&
			op4.envelopeGenerator.stage == Stage_OFF {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut[0])
		op2Output = op2.getOperatorOutput(noModulator)
		op3Output = op3.getOperatorOutput(ch.feedbackOut[1])
		op4Output = op4.getOperatorOutput(noModulator)

		channelOutput = op1Output + op2Output + op3Output + op4Output

	case 3:
		// (FB)OP1 --------> | -> OP4 -> OUT
		//     OP2 -> OP3 -> |
		if op4.envelopeGenerator.stage == Stage_OFF {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut[0])
		op2Output = op2.getOperatorOutput(noModulator)
		op3Output = op3.getOperatorOutput(op2Output * ch.toPhase)

		channelOutput = op4.getOperatorOutput((op1Output + op3Output) * ch.toPhase)

	case 4:
		// (FB)OP1 -> OP2 -> OP3 -> OP4 -> OUT
		if op4.envelopeGenerator.stage == Stage_OFF {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut[0])
		op2Output = op2.getOperatorOutput(op1Output * ch.toPhase)
		op3Output = op3.getOperatorOutput(op2Output * ch.toPhase)

		channelOutput = op4.getOperatorOutput(op3Output * ch.toPhase)

	case 5:
		// (FB)OP1 -> OP2 -> | -> OUT
		// (FB)OP3 -> OP4 -> |
		if op2.envelopeGenerator.stage == Stage_OFF && op4.envelopeGenerator.stage == Stage_OFF {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut[0])
		op2Output = op2.getOperatorOutput(op1Output * ch.toPhase)

		op3Output = op3.getOperatorOutput(ch.feedbackOut[1])
		op4Output = op4.getOperatorOutput(op3Output * ch.toPhase)

		channelOutput = op2Output + op4Output

	case 6:
		// (FB)OP1 ---------------> | -> OUT
		//     OP2 -> OP3 -> OP4 -> |
		if op1.envelopeGenerator.stage == Stage_OFF && op4.envelopeGenerator.stage == Stage_OFF {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut[0])
		op2Output = op2.getOperatorOutput(noModulator)
		op3Output = op3.getOperatorOutput(op2Output * ch.toPhase)
		op4Output = op4.getOperatorOutput(op3Output * ch.toPhase)

		channelOutput = op1Output + op4Output

	case 7:
		// (FB)OP1 --------> | -> OUT
		//     OP2 -> OP3 -> |
		//     OP4 --------> |
		if op1.envelopeGenerator.stage == Stage_OFF &&
			op3.envelopeGenerator.stage == Stage_OFF &&
			op4.envelopeGenerator.stage == Stage_OFF {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut[0])
		op2Output = op2.getOperatorOutput(noModulator)
		op3Output = op3.getOperatorOutput(op2Output * ch.toPhase)
		op4Output = op4.getOperatorOutput(noModulator)

		channelOutput = op1Output + op3Output + op4Output
	}

	ch.feedback[0][0] = ch.feedback[0][1]
	ch.feedback[0][1] = op1Output * ymfdata.FeedbackTable[op1.fb]
	ch.feedback[1][0] = ch.feedback[1][1]
	ch.feedback[1][1] = op3Output * ymfdata.FeedbackTable[op3.fb]

	ch.feedbackOut[0] = (ch.feedback[0][0] + ch.feedback[0][1]) / 2.0
	ch.feedbackOut[1] = (ch.feedback[1][0] + ch.feedback[1][1]) / 2.0

	// TODO: cache in 5bits
	pan := ch.chpan + (ch.panpot-15)*4
	if pan < 0 {
		pan = 0
	} else if 127 < pan {
		pan = 127
	}
	return ch.toStereo(channelOutput, ch.volume>>2, ch.expression>>2, pan)
}

func (ch *Channel4op) keyOn() {
	for _, op := range ch.Operators {
		op.keyOn()
	}
	ch.feedback[0][0] = 0
	ch.feedback[0][1] = 0
	ch.feedback[1][0] = 0
	ch.feedback[1][1] = 0
}

func (ch *Channel4op) keyOff() {
	for _, op := range ch.Operators {
		op.keyOff()
	}
}

func (ch *Channel4op) updateOperators() {
	// Key Scale Number, used in EnvelopeGenerator.setActualRates().
	keyScaleNumber := ch.block*2 + (ch.fnum >> 9)
	for i, op := range ch.Operators {
		op.updateOperator(keyScaleNumber, ch.fnum, ch.block, ch.bo, IS_MODULATOR[ch.alg][i])
	}
}

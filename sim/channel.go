package sim

import (
	"github.com/but80/fmfm/ymf/ymfdata"
)

const noModulator = 0

var isModulatorMatrix = [8][4]bool{
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

// Channel は、音源のチャンネルです。
type Channel struct {
	channelID int

	chip       *Chip
	fnum       int
	kon        int
	block      int
	alg        int
	panpot     int
	chpan      int
	volume     int
	expression int
	bo         int

	feedback1Prev        float64
	feedback1Curr        float64
	feedback3Prev        float64
	feedback3Curr        float64
	feedbackOut1         float64
	feedbackOut3         float64
	toPhase              float64
	volumeExpressionCoef float64
	panCoefL             float64
	panCoefR             float64

	operators [4]*operator
}

func newChannel4op(channelID int, chip *Chip) *Channel {
	ch := &Channel{
		chip:      chip,
		channelID: channelID,

		fnum:       0,
		kon:        0,
		block:      0,
		alg:        0,
		panpot:     15,
		chpan:      64,
		volume:     0,
		expression: 0,
		bo:         1,

		toPhase: 4,
	}
	for i := range ch.operators {
		ch.operators[i] = newOperator(channelID, i, chip)
	}
	ch.updatePanCoef()
	return ch
}

func (ch *Channel) setKON(v int) {
	if v == ch.kon {
		return
	}
	ch.kon = v
	if v == 0 {
		ch.keyOff()
	} else {
		ch.keyOn()
	}
}

func (ch *Channel) setBLOCK(v int) {
	ch.block = v
	ch.updateFrequency()
}

func (ch *Channel) setFNUM(v int) {
	ch.fnum = v
	ch.updateFrequency()
}

func (ch *Channel) setALG(v int) {
	ch.alg = v
	ch.feedback1Prev = 0
	ch.feedback1Curr = 0
	ch.feedback3Prev = 0
	ch.feedback3Curr = 0
	for i, op := range ch.operators {
		op.isModulator = isModulatorMatrix[ch.alg][i]
	}
}

func (ch *Channel) setLFO(v int) {
	for _, op := range ch.operators {
		op.setLFO(v)
	}
}

func (ch *Channel) setPANPOT(v int) {
	ch.panpot = v
	ch.updatePanCoef()
}

func (ch *Channel) setCHPAN(v int) {
	ch.chpan = v
	ch.updatePanCoef()
}

func (ch *Channel) updatePanCoef() {
	pan := ch.chpan + (ch.panpot-15)*4
	if pan < 0 {
		pan = 0
	} else if 127 < pan {
		pan = 127
	}
	ch.panCoefL = ymfdata.PanTable[pan][0]
	ch.panCoefR = ymfdata.PanTable[pan][1]
}

func (ch *Channel) setVOLUME(v int) {
	ch.volume = v
	ch.volumeExpressionCoef = ymfdata.VolumeTable[ch.volume>>2] * ymfdata.VolumeTable[ch.expression>>2]
}

func (ch *Channel) setEXPRESSION(v int) {
	ch.expression = v
	ch.volumeExpressionCoef = ymfdata.VolumeTable[ch.volume>>2] * ymfdata.VolumeTable[ch.expression>>2]
}

func (ch *Channel) setBO(v int) {
	ch.bo = v
	ch.updateFrequency()
}

func (ch *Channel) getChannelOutput() (float64, float64) {
	var channelOutput float64
	var op1Output float64
	var op2Output float64
	var op3Output float64
	var op4Output float64

	op1 := ch.operators[0]
	op2 := ch.operators[1]
	op3 := ch.operators[2]
	op4 := ch.operators[3]

	switch ch.alg {

	case 0:
		// (FB)1 -> 2 -> OUT
		if op2.envelopeGenerator.stage == stageOff {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut1)

		channelOutput = op2.getOperatorOutput(op1Output * ch.toPhase)

	case 1:
		// (FB)1 -> | -> OUT
		//     2 -> |
		if op1.envelopeGenerator.stage == stageOff && op2.envelopeGenerator.stage == stageOff {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut1)
		op2Output = op2.getOperatorOutput(noModulator)

		channelOutput = op1Output + op2Output

	case 2:
		// (FB)1 -> | -> OUT
		//     2 -> |
		// (FB)3 -> |
		//     4 -> |
		if op1.envelopeGenerator.stage == stageOff &&
			op2.envelopeGenerator.stage == stageOff &&
			op3.envelopeGenerator.stage == stageOff &&
			op4.envelopeGenerator.stage == stageOff {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut1)
		op2Output = op2.getOperatorOutput(noModulator)
		op3Output = op3.getOperatorOutput(ch.feedbackOut3)
		op4Output = op4.getOperatorOutput(noModulator)

		channelOutput = op1Output + op2Output + op3Output + op4Output

	case 3:
		// (FB)OP1 --------> | -> OP4 -> OUT
		//     OP2 -> OP3 -> |
		if op4.envelopeGenerator.stage == stageOff {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut1)
		op2Output = op2.getOperatorOutput(noModulator)
		op3Output = op3.getOperatorOutput(op2Output * ch.toPhase)

		channelOutput = op4.getOperatorOutput((op1Output + op3Output) * ch.toPhase)

	case 4:
		// (FB)OP1 -> OP2 -> OP3 -> OP4 -> OUT
		if op4.envelopeGenerator.stage == stageOff {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut1)
		op2Output = op2.getOperatorOutput(op1Output * ch.toPhase)
		op3Output = op3.getOperatorOutput(op2Output * ch.toPhase)

		channelOutput = op4.getOperatorOutput(op3Output * ch.toPhase)

	case 5:
		// (FB)OP1 -> OP2 -> | -> OUT
		// (FB)OP3 -> OP4 -> |
		if op2.envelopeGenerator.stage == stageOff && op4.envelopeGenerator.stage == stageOff {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut1)
		op2Output = op2.getOperatorOutput(op1Output * ch.toPhase)

		op3Output = op3.getOperatorOutput(ch.feedbackOut3)
		op4Output = op4.getOperatorOutput(op3Output * ch.toPhase)

		channelOutput = op2Output + op4Output

	case 6:
		// (FB)OP1 ---------------> | -> OUT
		//     OP2 -> OP3 -> OP4 -> |
		if op1.envelopeGenerator.stage == stageOff && op4.envelopeGenerator.stage == stageOff {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut1)
		op2Output = op2.getOperatorOutput(noModulator)
		op3Output = op3.getOperatorOutput(op2Output * ch.toPhase)
		op4Output = op4.getOperatorOutput(op3Output * ch.toPhase)

		channelOutput = op1Output + op4Output

	case 7:
		// (FB)OP1 --------> | -> OUT
		//     OP2 -> OP3 -> |
		//     OP4 --------> |
		if op1.envelopeGenerator.stage == stageOff &&
			op3.envelopeGenerator.stage == stageOff &&
			op4.envelopeGenerator.stage == stageOff {
			return 0, 0
		}

		op1Output = op1.getOperatorOutput(ch.feedbackOut1)
		op2Output = op2.getOperatorOutput(noModulator)
		op3Output = op3.getOperatorOutput(op2Output * ch.toPhase)
		op4Output = op4.getOperatorOutput(noModulator)

		channelOutput = op1Output + op3Output + op4Output
	}

	if op1.feedbackCoef != .0 {
		ch.feedback1Prev = ch.feedback1Curr
		ch.feedback1Curr = op1Output * op1.feedbackCoef
		ch.feedbackOut1 = (ch.feedback1Prev + ch.feedback1Curr) / 2.0
	}

	if op3.feedbackCoef != .0 {
		ch.feedback3Prev = ch.feedback3Curr
		ch.feedback3Curr = op3Output * op3.feedbackCoef
		ch.feedbackOut3 = (ch.feedback3Prev + ch.feedback3Curr) / 2.0
	}

	channelOutput *= ch.volumeExpressionCoef
	return channelOutput * ch.panCoefL, channelOutput * ch.panCoefR
}

func (ch *Channel) keyOn() {
	for _, op := range ch.operators {
		op.keyOn()
	}
	ch.feedback1Prev = 0
	ch.feedback1Curr = 0
	ch.feedback3Prev = 0
	ch.feedback3Curr = 0
}

func (ch *Channel) keyOff() {
	for _, op := range ch.operators {
		op.keyOff()
	}
}

func (ch *Channel) updateFrequency() {
	for _, op := range ch.operators {
		op.setFrequency(ch.fnum, ch.block, ch.bo)
	}
}
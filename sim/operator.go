package sim

import (
	"github.com/but80/fmfm/ymf/ymfdata"
)

type operator struct {
	isModulator bool

	dt             int
	ksr            int
	mult           int
	ksl            int
	ar             int
	dr             int
	sl             int
	sr             int
	rr             int
	xof            int
	ws             int
	feedbackCoef   float64
	keyScaleNumber int
	fnum           int
	block          int
	bo             int

	envelope    float64
	phaseFrac64 uint64

	envelopeGenerator *envelopeGenerator

	chip           *Chip
	channelID      int
	operatorIndex  int
	phaseGenerator *phaseGenerator
}

func newOperator(channelID, operatorIndex int, chip *Chip) *operator {
	return &operator{
		chip:              chip,
		channelID:         channelID,
		operatorIndex:     operatorIndex,
		phaseGenerator:    newPhaseGenerator(chip),
		envelopeGenerator: newEnvelopeGenerator(),
		isModulator:       false,
		bo:                1,
	}
}

func (op *operator) setEAM(v int) {
	op.envelopeGenerator.eam = v != 0
}

func (op *operator) setEVB(v int) {
	op.phaseGenerator.evb = v != 0
}

func (op *operator) setDAM(v int) {
	op.envelopeGenerator.dam = v
}

func (op *operator) setDVB(v int) {
	op.phaseGenerator.dvb = v
}

func (op *operator) setDT(v int) {
	op.dt = v
	op.updateFrequency()
}

func (op *operator) setKSR(v int) {
	// TODO: BOの影響は受けるのか？
	op.ksr = v
	op.updateEnvelope()
}

func (op *operator) setMULT(v int) {
	op.mult = v
	op.updateFrequency()
}

func (op *operator) setKSL(v int) {
	// TODO: BOの影響は受けるのか？
	op.ksl = v
	op.envelopeGenerator.setKeyScalingLevel(op.fnum, op.block, op.ksl)
}

func (op *operator) setTL(v int) {
	op.envelopeGenerator.setTotalLevel(v)
}

func (op *operator) setAR(v int) {
	op.ar = v
	op.envelopeGenerator.setActualAttackRate(op.ar, op.ksr, op.keyScaleNumber)
}

func (op *operator) setDR(v int) {
	op.dr = v
	op.envelopeGenerator.setActualDR(op.dr, op.ksr, op.keyScaleNumber)
}

func (op *operator) setSL(v int) {
	op.sl = v
	op.envelopeGenerator.setActualSustainLevel(op.sl)
}

func (op *operator) setSR(v int) {
	op.sr = v
	op.envelopeGenerator.setActualSR(op.sr, op.ksr, op.keyScaleNumber)
}

func (op *operator) setRR(v int) {
	op.rr = v
	op.envelopeGenerator.setActualRR(op.rr, op.ksr, op.keyScaleNumber)
}

func (op *operator) setXOF(v int) {
	op.xof = v
}

func (op *operator) setWS(v int) {
	op.ws = v
}

func (op *operator) setFB(v int) {
	op.feedbackCoef = ymfdata.FeedbackTable[v]
}

func (op *operator) next(modIndex int, modulator float64) float64 {
	if op.envelopeGenerator.stage == stageOff {
		return 0
	}

	op.envelope = op.envelopeGenerator.getEnvelope(modIndex)
	op.phaseFrac64 = op.phaseGenerator.getPhase(modIndex)

	sampleIndex := op.phaseFrac64 >> ymfdata.WaveformIndexShift
	sampleIndex += uint64((modulator + 1024.0) * ymfdata.WaveformLen)
	return ymfdata.Waveforms[op.ws][sampleIndex&1023] * op.envelope
}

func (op *operator) keyOn() {
	if 0 < op.ar {
		op.envelopeGenerator.keyOn()
		op.phaseGenerator.keyOn()
	} else {
		op.envelopeGenerator.stage = stageOff
	}
}

func (op *operator) keyOff() {
	if op.xof == 0 {
		op.envelopeGenerator.keyOff()
	}
}

func (op *operator) setFrequency(fnum, blk, bo int) {
	op.keyScaleNumber = blk*2 + (fnum >> 9)
	op.fnum = fnum
	op.block = blk
	op.bo = bo
	op.updateFrequency()
	op.updateEnvelope()
}

func (op *operator) updateFrequency() {
	op.phaseGenerator.setFrequency(op.fnum, op.block, op.bo, op.mult, op.dt)
}

func (op *operator) updateEnvelope() {
	op.envelopeGenerator.setActualAttackRate(op.ar, op.ksr, op.keyScaleNumber)
	op.envelopeGenerator.setActualDR(op.dr, op.ksr, op.keyScaleNumber)
	op.envelopeGenerator.setActualSR(op.sr, op.ksr, op.keyScaleNumber)
	op.envelopeGenerator.setActualRR(op.rr, op.ksr, op.keyScaleNumber)
}

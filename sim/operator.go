package sim

import (
	"fmt"
	"math"
	"strings"

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
		phaseGenerator:    newPhaseGenerator(chip.sampleRate),
		envelopeGenerator: newEnvelopeGenerator(chip.sampleRate),
		isModulator:       false,
		bo:                1,
	}
}

func (op *operator) dump() string {
	eg := op.envelopeGenerator
	pg := op.phaseGenerator

	lv := int((96.0 + math.Log10(eg.currentLevel)*20.0) / 8.0)
	lvstr := strings.Repeat("|", lv)

	cm := "C"
	if op.isModulator {
		cm = "M"
	}
	am := "-"
	if eg.eam {
		am = fmt.Sprintf("%d", eg.dam)
	}
	vb := "-"
	if pg.evb {
		vb = fmt.Sprintf("%d", pg.dvb)
	}
	phase := pg.phaseFrac64 >> ymfdata.WaveformIndexShift
	phstr := []byte("        ")
	phstr[phase>>(ymfdata.WaveformLenBits-3)] = '|'
	return fmt.Sprintf(
		"%d: %s mul=%02d ws=%02d adssr=%02d,%02d,%02d,%02d,%02d tl=%f am=%s vb=%s dt=%d ksr=%d fb=%3.2f ksn=%02d ksl=%f st=%s ph=%s lv=%s",
		op.operatorIndex,
		cm,
		op.mult,
		op.ws,
		op.ar,
		op.dr,
		op.sl,
		op.sr,
		op.rr,
		eg.tlCoef,
		am,
		vb,

		// actualAR        ,
		// arDiffPerSample ,
		// drCoefPerSample ,
		// srCoefPerSample ,
		// rrCoefPerSample ,
		// sustainLevel    ,
		// currentLevel    ,

		// op.phaseGenerator,
		op.dt,
		op.ksr,
		op.feedbackCoef,
		op.keyScaleNumber,
		eg.kslCoef,
		// op.fnum,
		// op.block,
		// op.bo,
		// op.xof,
		eg.stage.String(),
		string(phstr),
		lvstr,
	)
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
	phaseFrac64 := op.phaseGenerator.getPhase(modIndex)
	if op.envelopeGenerator.stage == stageOff {
		return 0
	}
	envelope := op.envelopeGenerator.getEnvelope(modIndex)

	sampleIndex := phaseFrac64 >> ymfdata.WaveformIndexShift
	sampleIndex += uint64((modulator + 1024.0) * ymfdata.WaveformLen)
	return ymfdata.Waveforms[op.ws][sampleIndex&1023] * envelope
}

func (op *operator) keyOn() {
	op.phaseGenerator.resetPhase()
	if 0 < op.ar {
		op.envelopeGenerator.keyOn()
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
	op.envelopeGenerator.setKeyScalingLevel(op.fnum, op.block, op.ksl)
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

package ymf

import (
	"math"

	"github.com/but80/fmfm/ymf/ymfdata"
)

type Operator struct {
	IsModulator bool

	eam            int
	evb            int
	dam            int
	dvb            int
	dt             int
	ksr            int
	mult           int
	ksl            int
	tl             int
	ar             int
	dr             int
	sl             int
	sr             int
	rr             int
	xof            int
	ws             int
	fb             int
	keyScaleNumber int
	f_number       int
	block          int
	bo             int

	envelope     float64
	phase        float64
	vibratoIndex float64 // TODO: modulation reset timing
	tremoloIndex float64 // TODO: modulation reset timing

	envelopeGenerator *EnvelopeGenerator

	chip           *Chip
	channelID      int
	operatorIndex  int
	phaseGenerator *PhaseGenerator
}

func newOperator(channelID, operatorIndex int, chip *Chip) *Operator {
	return &Operator{
		chip:              chip,
		channelID:         channelID,
		operatorIndex:     operatorIndex,
		phaseGenerator:    newPhaseGenerator(),
		envelopeGenerator: newEnvelopeGenerator(),
		IsModulator:       false,
		bo:                1,
	}
}

func (op *Operator) updateEAM() {
	op.eam = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_EAM)
}

func (op *Operator) updateEVB() {
	op.evb = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_EVB)
}

func (op *Operator) updateDAM() {
	op.dam = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_DAM)
}

func (op *Operator) updateDVB() {
	op.dvb = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_DVB)
}

func (op *Operator) updateDT() {
	op.dt = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_DT)
	op.phaseGenerator.setFrequency(op.f_number, op.block, op.bo, op.mult, op.dt)
}

func (op *Operator) updateKSR() {
	// TODO: BOの影響は受けるのか？
	op.ksr = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_KSR)
	op.envelopeGenerator.setActualAttackRate(op.ar, op.ksr, op.keyScaleNumber)
	op.envelopeGenerator.setActualDR(op.dr, op.ksr, op.keyScaleNumber)
	op.envelopeGenerator.setActualSR(op.sr, op.ksr, op.keyScaleNumber)
	op.envelopeGenerator.setActualRR(op.rr, op.ksr, op.keyScaleNumber)
}

func (op *Operator) updateMULT() {
	op.mult = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_MULT)
	op.phaseGenerator.setFrequency(op.f_number, op.block, op.bo, op.mult, op.dt)
}

func (op *Operator) updateKSL() {
	// TODO: BOの影響は受けるのか？
	op.ksl = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_KSL)
	op.envelopeGenerator.setAtennuation(op.f_number, op.block, op.ksl)
}

func (op *Operator) updateTL() {
	op.tl = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_TL)
	op.envelopeGenerator.setTotalLevel(op.tl)
}

func (op *Operator) updateAR() {
	op.ar = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_AR)
	op.envelopeGenerator.setActualAttackRate(op.ar, op.ksr, op.keyScaleNumber)
}

func (op *Operator) updateDR() {
	op.dr = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_DR)
	op.envelopeGenerator.setActualDR(op.dr, op.ksr, op.keyScaleNumber)
}

func (op *Operator) updateSL() {
	op.sl = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_SL)
	op.envelopeGenerator.setActualSustainLevel(op.sl)
}

func (op *Operator) updateSR() {
	op.sr = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_SR)
	op.envelopeGenerator.setActualSR(op.sr, op.ksr, op.keyScaleNumber)
}

func (op *Operator) updateRR() {
	op.rr = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_RR)
	op.envelopeGenerator.setActualRR(op.rr, op.ksr, op.keyScaleNumber)
}

func (op *Operator) updateXOF() {
	op.xof = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_XOF)
}

func (op *Operator) updateWS() {
	op.ws = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_WS)
}

func (op *Operator) updateFB() {
	op.fb = op.chip.registers.readOperator(op.channelID, op.operatorIndex, OpRegister_FB)
}

func (op *Operator) getOperatorOutput(modulator float64) float64 {
	if op.envelopeGenerator.stage == Stage_OFF {
		return 0
	}

	envelopeInDB := op.envelopeGenerator.getEnvelope(op.eam, op.dam, int(op.tremoloIndex))
	op.envelope = math.Pow(10, envelopeInDB/10)

	waveform := ymfdata.Waveforms[op.ws]

	op.phase = op.phaseGenerator.getPhase(op.evb, op.dvb, int(op.vibratoIndex))

	lfoFreq := ymfdata.LFOFrequency[op.chip.registers.readChannel(op.channelID, ChRegister_LFO)]
	lfoCycleSamples := float64(ymfdata.SampleRate) / lfoFreq

	vibratoTableLength := float64(len(ymfdata.VibratoTable[op.dvb]))
	op.vibratoIndex += vibratoTableLength / lfoCycleSamples
	op.vibratoIndex = math.Mod(op.vibratoIndex, vibratoTableLength)

	tremoloTableLength := float64(len(ymfdata.TremoloTable[op.dam]))
	op.tremoloIndex += tremoloTableLength / lfoCycleSamples
	op.tremoloIndex = math.Mod(op.tremoloIndex, tremoloTableLength)

	return op.getOutput(modulator, op.phase, waveform)
}

func (op *Operator) getOutput(modulator, outputPhase float64, waveform [1024]float64) float64 {
	_, outputPhase = math.Modf(outputPhase + modulator + 2)
	sampleIndex := int(outputPhase * 1024)
	return waveform[sampleIndex&1023] * op.envelope
}

func (op *Operator) keyOn() {
	if 0 < op.ar {
		op.envelopeGenerator.keyOn()
		op.phaseGenerator.keyOn()
		// op.vibratoIndex = 0
		// op.tremoloIndex = 0
	} else {
		op.envelopeGenerator.stage = Stage_OFF
	}
}

func (op *Operator) keyOff() {
	if op.xof == 0 {
		op.envelopeGenerator.keyOff()
	}
}

func (op *Operator) updateOperator(ksn, fnum, blk, bo int, isModulator bool) {
	op.keyScaleNumber = ksn
	op.f_number = fnum
	op.block = blk
	op.bo = bo
	op.IsModulator = isModulator

	op.updateEAM()
	op.updateEVB()
	op.updateDAM()
	op.updateDVB()
	op.updateDT()
	op.updateKSR()
	op.updateMULT()
	op.updateKSL()
	op.updateTL()
	op.updateAR()
	op.updateDR()
	op.updateSL()
	op.updateSR()
	op.updateRR()
	op.updateXOF()
	op.updateWS()
	op.updateFB()
}

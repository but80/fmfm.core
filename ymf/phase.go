package ymf

import (
	"github.com/but80/fmfm/ymf/ymfdata"
)

type PhaseGenerator struct {
	chip                 *Chip
	phaseFrac64          uint64
	phaseIncrementFrac64 uint64
}

func newPhaseGenerator(chip *Chip) *PhaseGenerator {
	return &PhaseGenerator{
		chip: chip,
	}
}

func (pg *PhaseGenerator) setFrequency(f_number, block, bo, mult, dt int) {
	baseFrequency := float64(f_number) * float64(uint(1)<<uint(block+3-bo)) / 16.0 / ymfdata.FnumK

	ksn := block<<1 | f_number>>9
	operatorFrequency := baseFrequency + ymfdata.DTCoef[dt][ksn]
	operatorFrequency *= ymfdata.MultTable[mult]

	pg.phaseIncrementFrac64 = uint64(operatorFrequency / pg.chip.SampleRate * ymfdata.Pow64Of2)
}

func (pg *PhaseGenerator) getPhase(evb, dvb, vibratoIndex int) uint64 {
	if 0 < evb {
		pg.phaseFrac64 += (pg.phaseIncrementFrac64 >> 32) * ymfdata.VibratoTableInt32Frac32[dvb][vibratoIndex]
	} else {
		pg.phaseFrac64 += pg.phaseIncrementFrac64
	}
	return pg.phaseFrac64
}

func (pg *PhaseGenerator) keyOn() {
	pg.phaseFrac64 = 0
}

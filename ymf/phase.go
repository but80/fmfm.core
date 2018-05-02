package ymf

import (
	"github.com/but80/fmfm/ymf/ymfdata"
)

type PhaseGenerator struct {
	phase          float64
	phaseIncrement float64
}

func newPhaseGenerator() *PhaseGenerator {
	return &PhaseGenerator{}
}

func (pg *PhaseGenerator) setFrequency(f_number, block, bo, mult, dt int) {
	baseFrequency := float64(f_number) * float64(uint(1)<<uint(block+1-bo)) * float64(ymfdata.SampleRate) / float64(1<<20)

	ksn := block<<1 | f_number>>9
	operatorFrequency := baseFrequency + ymfdata.DTCoef[dt][ksn]
	operatorFrequency *= ymfdata.MultTable[mult]

	pg.phaseIncrement = operatorFrequency / ymfdata.SampleRate
}

func (pg *PhaseGenerator) getPhase(evb, dvb, vibratoIndex int) float64 {
	if 0 < evb {
		pg.phase += pg.phaseIncrement * ymfdata.VibratoTable[dvb][vibratoIndex]
	} else {
		pg.phase += pg.phaseIncrement
	}
	pg.phase -= float64(int(pg.phase))
	return pg.phase
}

func (pg *PhaseGenerator) keyOn() {
	pg.phase = 0
}

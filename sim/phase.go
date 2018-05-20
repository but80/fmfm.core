package sim

import (
	"github.com/but80/fmfm.core/ymf/ymfdata"
)

type phaseGenerator struct {
	sampleRate           float64
	evb                  bool
	dvb                  int
	phaseFrac64          uint64
	phaseIncrementFrac64 uint64
}

func newPhaseGenerator(sampleRate float64) *phaseGenerator {
	pg := &phaseGenerator{sampleRate: sampleRate}
	pg.reset()
	return pg
}

func (pg *phaseGenerator) reset() {
	pg.phaseFrac64 = 0
}

func (pg *phaseGenerator) resetAll() {
	pg.evb = false
	pg.dvb = 0
	pg.phaseIncrementFrac64 = 0
	pg.reset()
}

func (pg *phaseGenerator) setFrequency(fnum, block, bo, mult, dt int) {
	baseFrequency := float64(fnum<<uint(block+3-bo)) / (16.0 * ymfdata.FNUMCoef)

	ksn := block<<1 | fnum>>9
	operatorFrequency := baseFrequency + ymfdata.DTCoef[dt][ksn]

	pg.phaseIncrementFrac64 = uint64(operatorFrequency / pg.sampleRate * ymfdata.Pow64Of2)

	// 端数切り捨て後に掛けないとオペレータ間でズレる
	pg.phaseIncrementFrac64 *= ymfdata.MultTable2[mult]
	pg.phaseIncrementFrac64 >>= 1
}

func (pg *phaseGenerator) getPhase(vibratoIndex int) uint64 {
	if pg.evb {
		pg.phaseFrac64 += (pg.phaseIncrementFrac64 >> 32) * ymfdata.VibratoTableInt32Frac32[pg.dvb][vibratoIndex]
	} else {
		pg.phaseFrac64 += pg.phaseIncrementFrac64
	}
	return pg.phaseFrac64
}

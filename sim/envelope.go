package sim

import (
	"math"

	"github.com/but80/fmfm.core/ymf/ymfdata"
)

type stage int

const (
	stageOff stage = iota
	stageAttack
	stageDecay
	stageSustain
	stageRelease
)

func (s stage) String() string {
	switch s {
	case stageOff:
		return "-"
	case stageAttack:
		return "A"
	case stageDecay:
		return "D"
	case stageSustain:
		return "S"
	case stageRelease:
		return "R"
	default:
		return "?"
	}
}

const epsilon = 1.0 / 32768.0

type envelopeGenerator struct {
	sampleRate      float64
	stage           stage
	eam             bool
	dam             int
	actualAR        int
	arDiffPerSample float64
	drCoefPerSample float64
	srCoefPerSample float64
	rrCoefPerSample float64
	kslCoef         float64
	tlCoef          float64
	kslTlCoef       float64
	sustainLevel    float64
	currentLevel    float64
}

func newEnvelopeGenerator(sampleRate float64) *envelopeGenerator {
	eg := &envelopeGenerator{
		sampleRate:   sampleRate,
		stage:        stageOff,
		currentLevel: 0,
	}
	eg.setTotalLevel(0)
	eg.setKeyScalingLevel(0, 0, 0)
	return eg
}

func (eg *envelopeGenerator) setActualSustainLevel(sl int) {
	if sl == 0x0f {
		eg.sustainLevel = 0
	} else {
		slDB := -3.0 * float64(sl)
		eg.sustainLevel = math.Pow(10.0, slDB/20.0)
	}
}

func (eg *envelopeGenerator) setTotalLevel(tl int) {
	tlDB := float64(tl) * -0.75
	eg.tlCoef = math.Pow(10.0, tlDB/20.0)
	eg.kslTlCoef = eg.kslCoef * eg.tlCoef
}

func (eg *envelopeGenerator) setKeyScalingLevel(fnum, block, ksl int) {
	eg.kslCoef = ymfdata.KSLTable[ksl][block][fnum>>5]
	eg.kslTlCoef = eg.kslCoef * eg.tlCoef
}

func (eg *envelopeGenerator) setActualAttackRate(attackRate, ksr, keyScaleNumber int) {
	eg.actualAR = calculateActualRate(attackRate, ksr, keyScaleNumber)
	if eg.actualAR == 0 {
		eg.arDiffPerSample = 0
	} else {
		sec := 1.75 * math.Pow(.5, float64(eg.actualAR)/4.0-1.0)
		eg.arDiffPerSample = 1.0 / (sec * eg.sampleRate)
	}
}

func (eg *envelopeGenerator) setActualDR(dr, ksr, keyScaleNumber int) {
	if dr == 0 {
		eg.drCoefPerSample = 1.0
	} else {
		dbPerSecAt4 := decayDBPerSecAt4[ksr][keyScaleNumber] / 2.0
		dbPerSample := dbPerSecAt4 * float64(uint(1)<<uint(dr)) / 16.0 / eg.sampleRate
		eg.drCoefPerSample = math.Pow(10, -dbPerSample/10)
	}
}

func (eg *envelopeGenerator) setActualSR(sr, ksr, keyScaleNumber int) {
	if sr == 0 {
		eg.srCoefPerSample = 1.0
	} else {
		dbPerSecAt4 := decayDBPerSecAt4[ksr][keyScaleNumber] / 2.0
		dbPerSample := dbPerSecAt4 * float64(uint(1)<<uint(sr)) / 16.0 / eg.sampleRate
		eg.srCoefPerSample = math.Pow(10, -dbPerSample/10)
	}
}

func (eg *envelopeGenerator) setActualRR(rr, ksr, keyScaleNumber int) {
	if rr == 0 {
		eg.rrCoefPerSample = 1.0
	} else {
		dbPerSecAt4 := decayDBPerSecAt4[ksr][keyScaleNumber] / 2.0
		dbPerSample := dbPerSecAt4 * float64(uint(1)<<uint(rr)) / 16.0 / eg.sampleRate
		eg.rrCoefPerSample = math.Pow(10, -dbPerSample/10)
	}
}

func calculateActualRate(rate, ksr, keyScaleNumber int) int {
	rof := rateOffset[ksr][keyScaleNumber]
	actualRate := rate*4 + rof
	if 63 < actualRate {
		actualRate = 63
	}
	return actualRate
}

func (eg *envelopeGenerator) getEnvelope(tremoloIndex int) float64 {
	switch eg.stage {

	case stageAttack:
		eg.currentLevel += eg.arDiffPerSample
		if eg.currentLevel < 1.0 {
			break
		}
		eg.currentLevel = 1.0
		eg.stage = stageDecay
		fallthrough

	case stageDecay:
		if eg.sustainLevel < eg.currentLevel {
			eg.currentLevel *= eg.drCoefPerSample
			break
		}
		eg.stage = stageSustain
		fallthrough

	case stageSustain:
		if epsilon < eg.currentLevel {
			eg.currentLevel *= eg.srCoefPerSample
		} else {
			eg.stage = stageOff
		}
		break

	case stageRelease:
		if epsilon < eg.currentLevel {
			eg.currentLevel *= eg.rrCoefPerSample
		} else {
			eg.stage = stageOff
		}
		break
	}

	result := eg.currentLevel
	if eg.eam {
		result *= ymfdata.TremoloTable[eg.dam][tremoloIndex]
	}
	return result * eg.kslTlCoef
}

func (eg *envelopeGenerator) keyOn() {
	eg.stage = stageAttack
}

func (eg *envelopeGenerator) keyOff() {
	if eg.stage != stageOff {
		eg.stage = stageRelease
	}
}

// This table is indexed by the value of operator.ksr
// and the value of ChannelRegister.keyScaleNumber.
// TODO: ARのKSR影響の検証
var rateOffset = [2][16]int{
	{0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
}

// DR/SR/RR=4 における共通の減衰速度 [振幅dB/sec]
// ・使用時は2で割ってエネルギーdBに変換
// ・DR/SR/RR が1増えると速度は2倍になる
var decayDBPerSecAt4 = [2][16]float64{
	// 添字は keyScaleNumber (0..15)
	{17.9342, 17.9342, 17.9342, 17.9342, 17.9342, 22.4116, 22.4116, 22.4116, 22.4116, 26.9076, 26.9076, 26.9076, 26.9076, 31.3661, 31.3661, 31.3661},      // KSR=0
	{17.9465, 22.4376, 22.4376, 31.4026, 31.4026, 44.8696, 44.8696, 62.7959, 62.7959, 89.6707, 89.6707, 125.5546, 125.5546, 179.2684, 179.2684, 250.9128}, // KSR=1
}

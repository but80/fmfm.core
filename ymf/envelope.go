package ymf

import (
	"math"

	"github.com/but80/fmfm/ymf/ymfdata"
)

type stage int

const (
	stageOff stage = iota
	stageAttack
	stageDecay
	stageSustain
	stageRelease
)

const epsilon = 1.0 / 32768.0

type EnvelopeGenerator struct {
	stage stage

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

func newEnvelopeGenerator() *EnvelopeGenerator {
	eg := &EnvelopeGenerator{
		stage:        stageOff,
		currentLevel: 0,
	}
	eg.setTotalLevel(0)
	eg.setKeyScalingLevel(0, 0, 0)
	return eg
}

func (eg *EnvelopeGenerator) setActualSustainLevel(sl int) {
	if sl == 0x0f {
		eg.sustainLevel = 0
	} else {
		slDB := -3.0 * float64(sl)
		eg.sustainLevel = math.Pow(10.0, slDB/20.0)
	}
}

func (eg *EnvelopeGenerator) setTotalLevel(tl int) {
	tlDB := float64(tl) * -0.75
	eg.tlCoef = math.Pow(10.0, tlDB/20.0)
	eg.kslTlCoef = eg.kslCoef * eg.tlCoef
}

func (eg *EnvelopeGenerator) setKeyScalingLevel(f_number, block, ksl int) {
	hi4bits := f_number >> 6 & 0x0f
	attenuation := .0
	switch ksl {
	case 0:
		attenuation = .0
	case 1:
		// ~3 dB/Octave
		attenuation = ymfdata.KSL3DBTable[hi4bits][block]
	case 2:
		// ~1.5 dB/Octave
		attenuation = ymfdata.KSL3DBTable[hi4bits][block] / 2.0
	case 3:
		// ~6 dB/Octave
		attenuation = ymfdata.KSL3DBTable[hi4bits][block] * 2.0
	}
	eg.kslCoef = math.Pow(10, attenuation/20.0)
	eg.kslTlCoef = eg.kslCoef * eg.tlCoef
}

func (eg *EnvelopeGenerator) setActualAttackRate(attackRate, ksr, keyScaleNumber int) {
	eg.actualAR = calculateActualRate(attackRate, ksr, keyScaleNumber)
	if eg.actualAR == 0 {
		eg.arDiffPerSample = 0
	} else {
		sec := 1.75 * math.Pow(.5, float64(eg.actualAR)/4.0-1.0)
		eg.arDiffPerSample = 1.0 / (sec * ymfdata.SampleRate)
	}
}

func (eg *EnvelopeGenerator) setActualDR(dr, ksr, keyScaleNumber int) {
	if dr == 0 {
		eg.drCoefPerSample = 1.0
	} else {
		dbPerSecAt4 := decayDBPerSecAt4[ksr][keyScaleNumber] / 2.0
		dbPerSample := dbPerSecAt4 * float64(uint(1)<<uint(dr)) / 16.0 / ymfdata.SampleRate
		eg.drCoefPerSample = math.Pow(10, -dbPerSample/10)
	}
}

func (eg *EnvelopeGenerator) setActualSR(sr, ksr, keyScaleNumber int) {
	if sr == 0 {
		eg.srCoefPerSample = 1.0
	} else {
		dbPerSecAt4 := decayDBPerSecAt4[ksr][keyScaleNumber] / 2.0
		dbPerSample := dbPerSecAt4 * float64(uint(1)<<uint(sr)) / 16.0 / ymfdata.SampleRate
		eg.srCoefPerSample = math.Pow(10, -dbPerSample/10)
	}
}

func (eg *EnvelopeGenerator) setActualRR(rr, ksr, keyScaleNumber int) {
	if rr == 0 {
		eg.rrCoefPerSample = 1.0
	} else {
		dbPerSecAt4 := decayDBPerSecAt4[ksr][keyScaleNumber] / 2.0
		dbPerSample := dbPerSecAt4 * float64(uint(1)<<uint(rr)) / 16.0 / ymfdata.SampleRate
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

func (eg *EnvelopeGenerator) getEnvelope(eam, dam, tremoloIndex int) float64 {
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
	if eam != 0 {
		result *= ymfdata.TremoloTable[dam][tremoloIndex]
	}
	return result * eg.kslTlCoef
}

func (eg *EnvelopeGenerator) keyOn() {
	eg.stage = stageAttack
}

func (eg *EnvelopeGenerator) keyOff() {
	if eg.stage != stageOff {
		eg.stage = stageRelease
	}
}

func dbToX(dB float64) float64 {
	return math.Log2(-dB)
}

func percentageToDB(percentage float64) float64 {
	return math.Log10(percentage) * 10
}

func percentageToX(percentage float64) float64 {
	return dbToX(percentageToDB(percentage))
}

// This table is indexed by the value of Operator.ksr
// and the value of ChannelRegister.keyScaleNumber.
var rateOffset = [2][16]int{
	{0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
}

// These attack periods in miliseconds were taken from the YMF278B manual.
// The attack actual rates range from 0 to 63, with different data for
// 0%-100% and for 10%-90%:
var attackTimeValuesTable = [...][2]float64{
	{1e30, 1e30}, {1e30, 1e30}, {1e30, 1e30}, {1e30, 1e30}, // TODO: replace with math.Inf
	{2826.24, 1482.75}, {2252.80, 1155.07}, {1884.16, 991.23}, {1597.44, 868.35},
	{1413.12, 741.38}, {1126.40, 577.54}, {942.08, 495.62}, {798.72, 434.18},
	{706.56, 370.69}, {563.20, 288.77}, {471.04, 247.81}, {399.36, 217.09},

	{353.28, 185.34}, {281.60, 144.38}, {235.52, 123.90}, {199.68, 108.54},
	{176.76, 92.67}, {140.80, 72.19}, {117.76, 61.95}, {99.84, 54.27},
	{88.32, 46.34}, {70.40, 36.10}, {58.88, 30.98}, {49.92, 27.14},
	{44.16, 23.17}, {35.20, 18.05}, {29.44, 15.49}, {24.96, 13.57},

	{22.08, 11.58}, {17.60, 9.02}, {14.72, 7.74}, {12.48, 6.78},
	{11.04, 5.79}, {8.80, 4.51}, {7.36, 3.87}, {6.24, 3.39},
	{5.52, 2.90}, {4.40, 2.26}, {3.68, 1.94}, {3.12, 1.70},
	{2.76, 1.45}, {2.20, 1.13}, {1.84, 0.97}, {1.56, 0.85},

	{1.40, 0.73}, {1.12, 0.61}, {0.92, 0.49}, {0.80, 0.43},
	{0.70, 0.37}, {0.56, 0.31}, {0.46, 0.26}, {0.42, 0.22},
	{0.38, 0.19}, {0.30, 0.14}, {0.24, 0.11}, {0.20, 0.11},
	{0.00, 0.00}, {0.00, 0.00}, {0.00, 0.00}, {0.00, 0.00},
}

// DR/SR/RR=4 における共通の減衰速度 [振幅dB/sec]
// ・使用時は2で割ってエネルギーdBに変換
// ・DR/SR/RR が1増えると速度は2倍になる
var decayDBPerSecAt4 = [2][16]float64{
	// 添字は keyScaleNumber (0..15)
	{17.9342, 17.9342, 17.9342, 17.9342, 17.9342, 22.4116, 22.4116, 22.4116, 22.4116, 26.9076, 26.9076, 26.9076, 26.9076, 31.3661, 31.3661, 31.3661},      // KSR=0
	{17.9465, 22.4376, 22.4376, 31.4026, 31.4026, 44.8696, 44.8696, 62.7959, 62.7959, 89.6707, 89.6707, 125.5546, 125.5546, 179.2684, 179.2684, 250.9128}, // KSR=1
}

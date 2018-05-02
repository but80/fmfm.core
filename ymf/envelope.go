package ymf

import (
	"math"

	"github.com/but80/fmfm/ymf/ymfdata"
)

type Stage int

const (
	Stage_ATTACK Stage = iota
	Stage_DECAY
	Stage_SUSTAIN
	Stage_RELEASE
	Stage_OFF
)

const envelopeMinimum = -96
const envelopeResolution = 0.1875

type EnvelopeGenerator struct {
	stage Stage

	actualAR          int
	xAttackIncrement  float64
	drDBPerSample     float64
	srDBPerSample     float64
	rrDBPerSample     float64
	kslAttenuation    float64
	tl                float64
	sl                float64
	resolutionMaximum float64
	percentage10      float64
	percentage90      float64

	currentLevel float64
	currentDB    float64
}

func newEnvelopeGenerator() *EnvelopeGenerator {
	return &EnvelopeGenerator{
		stage:        Stage_OFF,
		currentLevel: 0,
		percentage10: percentageToX(0.1),
		percentage90: percentageToX(0.9),
		currentDB:    -96,
	}
}

func (eg *EnvelopeGenerator) setActualSustainLevel(sl int) {
	// If all SL bits are 1, sustain level is set to -93 dB:
	if sl == 0x0f {
		eg.sl = -93
		return
	}
	// The datasheet states that the SL formula is
	// sustainLevel = -24*d7 -12*d6 -6*d5 -3*d4,
	// translated as:
	eg.sl = float64(-3 * sl)
}

func (eg *EnvelopeGenerator) setTotalLevel(tl int) {
	// The datasheet states that the TL formula is
	// TL = -(24*d5 + 12*d4 + 6*d3 + 3*d2 + 1.5*d1 + 0.75*d0),
	// translated as:
	eg.tl = float64(tl) * -0.75
}

func (eg *EnvelopeGenerator) setAtennuation(f_number, block, ksl int) {
	hi4bits := f_number >> 6 & 0x0f
	switch ksl {
	case 0:
		eg.kslAttenuation = 0
	case 1:
		// ~3 dB/Octave
		eg.kslAttenuation = ymfdata.KSL3DBTable[hi4bits][block]
	case 2:
		// ~1.5 dB/Octave
		eg.kslAttenuation = ymfdata.KSL3DBTable[hi4bits][block] / 2
	case 3:
		// ~6 dB/Octave
		eg.kslAttenuation = ymfdata.KSL3DBTable[hi4bits][block] * 2
	}
}

func (eg *EnvelopeGenerator) setActualAttackRate(attackRate, ksr, keyScaleNumber int) {
	eg.actualAR = calculateActualRate(attackRate, ksr, keyScaleNumber)
	if eg.actualAR == 0 {
		eg.xAttackIncrement = 0
	} else {
		sec := 1.75 * math.Pow(.5, float64(eg.actualAR)/4.0-1.0)
		eg.xAttackIncrement = 1.0 / (sec * ymfdata.SampleRate)
	}
}

func (eg *EnvelopeGenerator) setActualDR(dr, ksr, keyScaleNumber int) {
	if dr == 0 {
		eg.drDBPerSample = 0
	} else {
		dbPerSec := decayDBPerSecAt4[ksr][keyScaleNumber] * float64(uint(1)<<uint(dr)) / 16
		eg.drDBPerSample = dbPerSec / 2 / ymfdata.SampleRate
	}
}

func (eg *EnvelopeGenerator) setActualSR(sr, ksr, keyScaleNumber int) {
	if sr == 0 {
		eg.srDBPerSample = 0
	} else {
		dbPerSec := decayDBPerSecAt4[ksr][keyScaleNumber] * float64(uint(1)<<uint(sr)) / 16
		eg.srDBPerSample = dbPerSec / 2 / ymfdata.SampleRate
	}
}

func (eg *EnvelopeGenerator) setActualRR(rr, ksr, keyScaleNumber int) {
	if rr == 0 {
		eg.rrDBPerSample = 0
	} else {
		dbPerSec := decayDBPerSecAt4[ksr][keyScaleNumber] * float64(uint(1)<<uint(rr)) / 16
		eg.rrDBPerSample = dbPerSec / 2 / ymfdata.SampleRate
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
	// The datasheets attenuation values
	// must be halved to match the real OPL3 output.
	envelopeSustainLevel := float64(eg.sl) / 2.0
	envelopeTremolo := ymfdata.TremoloTable[dam][tremoloIndex] / 2.0
	envelopeAttenuation := eg.kslAttenuation / 2.0
	envelopeTotalLevel := float64(eg.tl) / 2.0

	//
	// Envelope Generation
	//
	switch eg.stage {

	case Stage_ATTACK:
		eg.currentLevel += eg.xAttackIncrement
		eg.currentDB = 10.0 * math.Log10(eg.currentLevel)
		if eg.currentDB < .0 {
			break
		}
		eg.currentLevel = 1.0
		eg.currentDB = .0
		eg.stage = Stage_DECAY
		fallthrough

	case Stage_DECAY:
		// The decay and release are linear.
		if envelopeSustainLevel < eg.currentDB {
			eg.currentDB -= eg.drDBPerSample
			break
		}
		eg.stage = Stage_SUSTAIN
		fallthrough

	case Stage_SUSTAIN:
		if envelopeMinimum < eg.currentDB {
			eg.currentDB -= eg.srDBPerSample
		} else {
			eg.stage = Stage_OFF
		}
		break

	case Stage_RELEASE:
		// If we have Key OFF, only here we are in the Release stage.
		// Now, we can turn EGT back and forth and it will have no effect,i.e.,
		// it will release inexorably to the Off stage.
		if envelopeMinimum < eg.currentDB {
			eg.currentDB -= eg.rrDBPerSample
		} else {
			eg.stage = Stage_OFF
		}
		break
	}

	// Ongoing original envelope
	outputEnvelope := eg.currentDB

	// Tremolo
	if eam != 0 {
		outputEnvelope += envelopeTremolo
	}

	// Attenuation
	outputEnvelope += envelopeAttenuation

	// Total Level
	outputEnvelope += envelopeTotalLevel

	return outputEnvelope
}

func (eg *EnvelopeGenerator) keyOn() {
	eg.currentLevel = math.Pow(10, eg.currentDB/10.0)
	eg.stage = Stage_ATTACK
}

func (eg *EnvelopeGenerator) keyOff() {
	if eg.stage != Stage_OFF {
		eg.stage = Stage_RELEASE
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

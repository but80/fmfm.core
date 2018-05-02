package ymfdata

import (
	"math"
)

const CHANNEL_COUNT = 16
const SampleRate = 44100 // 49700

var VolumeTable = [...]float64{
	1e30, 47.9, 42.6, 37.2, 33.1, 29.8, 27.0, 24.6,
	22.4, 20.6, 18.9, 17.3, 15.9, 14.6, 13.4, 12.2,
	11.1, 10.1, 9.2, 8.3, 7.4, 6.6, 5.8, 5.1,
	4.4, 3.6, 3.0, 2.3, 1.7, 1.1, 0.6, 0.0,
}

var PanTable [128][2]float64

var DTCoef = [8][16]float64{
	{0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00},
	{0.00, 0.00, 0.05, 0.05, 0.05, 0.05, 0.09, 0.09, 0.14, 0.14, 0.18, 0.23, 0.27, 0.32, 0.37, 0.37},
	{0.05, 0.05, 0.09, 0.09, 0.14, 0.14, 0.18, 0.23, 0.27, 0.32, 0.41, 0.46, 0.59, 0.64, 0.73, 0.73},
	{0.09, 0.09, 0.14, 0.14, 0.18, 0.23, 0.28, 0.32, 0.41, 0.46, 0.59, 0.64, 0.87, 0.91, 1.00, 1.00},
	{0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00},
	{-0.00, -0.00, -0.05, -0.05, -0.05, -0.05, -0.09, -0.09, -0.14, -0.14, -0.18, -0.23, -0.27, -0.32, -0.37, -0.37},
	{-0.05, -0.05, -0.09, -0.09, -0.14, -0.14, -0.18, -0.23, -0.27, -0.32, -0.41, -0.46, -0.59, -0.64, -0.73, -0.73},
	{-0.09, -0.09, -0.14, -0.14, -0.18, -0.23, -0.28, -0.32, -0.41, -0.46, -0.59, -0.64, -0.87, -0.91, -1.00, -1.00},
}

var LFOFrequency = [4]float64{1.8, 4.0, 5.9, 7.0}

const ModTableLen = 8192

var VibratoTable [4][ModTableLen]float64

var TremoloTable [4][ModTableLen]float64

var FeedbackTable = [8]float64{0, 1 / 32, 1 / 16, 1 / 8, 1 / 4, 1 / 2, 1, 2}

var MultTable = [16]float64{0.5, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 12, 12, 15, 15}

var KSL3DBTable = [16][8]float64{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, -3, -6, -9},
	{0, 0, 0, 0, -3, -6, -9, -12},
	{0, 0, 0, -1.875, -4.875, -7.875, -10.875, -13.875},

	{0, 0, 0, -3, -6, -9, -12, -15},
	{0, 0, -1.125, -4.125, -7.125, -10.125, -13.125, -16.125},
	{0, 0, -1.875, -4.875, -7.875, -10.875, -13.875, -16.875},
	{0, 0, -2.625, -5.625, -8.625, -11.625, -14.625, -17.625},

	{0, 0, -3, -6, -9, -12, -15, -18},
	{0, -0.750, -3.750, -6.750, -9.750, -12.750, -15.750, -18.750},
	{0, -1.125, -4.125, -7.125, -10.125, -13.125, -16.125, -19.125},
	{0, -1.500, -4.500, -7.500, -10.500, -13.500, -16.500, -19.500},

	{0, -1.875, -4.875, -7.875, -10.875, -13.875, -16.875, -19.875},
	{0, -2.250, -5.250, -8.250, -11.250, -14.250, -17.250, -20.250},
	{0, -2.625, -5.625, -8.625, -11.625, -14.625, -17.625, -20.625},
	{0, -3, -6, -9, -12, -15, -18, -21},
}

var Waveforms [32][]float64

func CalculateIncrement(begin, end, period float64) float64 {
	return (end - begin) / SampleRate * (1 / period)
}

func init() {
	// generate volume table
	for i := range VolumeTable {
		VolumeTable[i] = math.Pow(10, -VolumeTable[i]/20)
	}
	VolumeTable[0] = 0

	// generate pan table
	for i := 0; i < 128; i++ {
		a := math.Pi * .5 * float64(i) / 127
		PanTable[i][0] = math.Cos(a)
		PanTable[i][1] = math.Sin(a)
	}

	// generate vibrato table
	{
		semitone := math.Pow(2, float64(1)/12)
		cent := math.Pow(semitone, float64(1)/100)

		// https://github.com/yamaha-webmusic/ymf825board/blob/991485a4cbbe07d84cca707701999875fbc17c74/manual/fbd_spec3.md#dam-eam-dvb-evb
		vibratoDepth := [4]float64{
			math.Pow(cent, 3.4),
			math.Pow(cent, 6.7),
			math.Pow(cent, 13.5),
			math.Pow(cent, 26.8),
		}
		for dvb := 0; dvb < 4; dvb++ {
			i := 0
			for ; i < 1024; i++ {
				VibratoTable[dvb][i] = 1
			}
			for ; i < 2048; i++ {
				VibratoTable[dvb][i] = math.Sqrt(vibratoDepth[dvb])
			}
			for ; i < 3072; i++ {
				VibratoTable[dvb][i] = vibratoDepth[dvb]
			}
			for ; i < 4096; i++ {
				VibratoTable[dvb][i] = math.Sqrt(vibratoDepth[dvb])
			}
			for ; i < 5120; i++ {
				VibratoTable[dvb][i] = 1
			}
			for ; i < 6144; i++ {
				VibratoTable[dvb][i] = 1 / math.Sqrt(vibratoDepth[dvb])
			}
			for ; i < 7168; i++ {
				VibratoTable[dvb][i] = 1 / vibratoDepth[dvb]
			}
			for ; i < 8192; i++ {
				VibratoTable[dvb][i] = 1 / math.Sqrt(vibratoDepth[dvb])
			}
		}
	}

	// generate tremolo table
	{
		tremoloFrequency := float64(SampleRate) / float64(ModTableLen)

		// https://github.com/yamaha-webmusic/ymf825board/blob/991485a4cbbe07d84cca707701999875fbc17c74/manual/fbd_spec3.md#dam-eam-dvb-evb
		tremoloDepth := [4]float64{-1.3, -2.8, -5.8, -11.8} // dB

		for dam := 0; dam < 4; dam++ {
			tremoloIncrement := CalculateIncrement(tremoloDepth[dam], 0, 1/(2*tremoloFrequency))
			TremoloTable[dam][0] = tremoloDepth[dam]
			counter := 0
			for TremoloTable[0][counter] < 0 {
				counter++
				TremoloTable[dam][counter] = TremoloTable[dam][counter-1] + tremoloIncrement
			}
			for tremoloDepth[0] < TremoloTable[0][counter] && counter < ModTableLen-1 {
				counter++
				TremoloTable[dam][counter] = TremoloTable[dam][counter-1] - tremoloIncrement
			}
		}

		// convert dB -> coef
		for dam := 0; dam < 4; dam++ {
			for i := range TremoloTable[dam] {
				TremoloTable[dam][i] = math.Pow(10.0, TremoloTable[dam][i]/20.0)
			}
		}
	}

	// convert LFO frequency
	{
		for i := range LFOFrequency {
			LFOFrequency[i] /= SampleRate
			LFOFrequency[i] *= ModTableLen
		}
	}

	// generate waveform table
	{
		/*

			  波形は完全に上位互換

			  OPL3:
				SIN   | 0:^v 1:^- 2:^^ 3:''
				SINx2 | 4:▚- 5:"-
				SQR   | 6:▀▄
				EXP   | 7:＼

			  MA-5:
				SIN   | 0:^v 1:^- 2:^^ 3:''
				SINx2 | 4:▚- 5:"-
				SQR   | 6:▀▄ 14:▀-
				SQRx2 | 22:▘▘ 30:▘-
				EXP   | 7:＼
				CSIN  | 8:▀▄ 9:^- 10:^^ 11:''
				CSINx2| 12:▚- 13:"-
				TRI   | 16:▀▄ 17:^- 18:^^ 19:''
				TRIx2 | 20:▚- 21:"-
				SAW   | 24:▀▄ 25:▀- 26:▀▀ 27:''
				SAWx2 | 28:▚- 29:"-
		*/

		for i := range Waveforms {
			Waveforms[i] = make([]float64, 1024)
		}

		// ^v to ^-
		copyHalf := func(src []float64) []float64 {
			dst := make([]float64, 1024)
			for i := 0; i < 512; i++ {
				dst[i] = src[i]
				dst[512+i] = 0
			}
			return dst
		}

		// ^v to ^^
		copyAbs := func(src []float64) []float64 {
			dst := make([]float64, 1024)
			for i := 0; i < 512; i++ {
				dst[i] = src[i]
				dst[512+i] = src[i]
			}
			return dst
		}

		// ^v to ''
		copyAbsQuarter := func(src []float64) []float64 {
			dst := make([]float64, 1024)
			for i := 0; i < 256; i++ {
				dst[i] = src[i]
				dst[512+i] = src[i]
				dst[256+i] = 0
				dst[768+i] = 0
			}
			return dst
		}

		// ^v to ▚-
		copyOct := func(src []float64) []float64 {
			dst := make([]float64, 1024)
			for i := 0; i < 512; i++ {
				dst[i] = src[i*2]
				dst[512+i] = 0
			}
			return dst
		}

		// ^v to "-
		copyAbsOct := func(src []float64) []float64 {
			dst := make([]float64, 1024)
			for i := 0; i < 256; i++ {
				dst[i] = src[i*2]
				dst[256+i] = src[i*2]
				dst[512+i] = 0
				dst[768+i] = 0
			}
			return dst
		}

		// ==================================================
		// sine wave
		for i := 0; i < 1024; i++ {
			Waveforms[0][i] = math.Sin(2 * math.Pi * float64(i) / 1024)
		}
		sineTable := Waveforms[0]

		// SIN   | 0:^v 1:^- 2:^^ 3:''
		Waveforms[1] = copyHalf(sineTable)
		Waveforms[2] = copyAbs(sineTable)
		Waveforms[3] = copyAbsQuarter(sineTable)
		// SINx2 | 4:▚- 5:"-
		Waveforms[4] = copyOct(sineTable)
		Waveforms[5] = copyAbsOct(sineTable)

		// ==================================================
		// square wave
		for i := 0; i < 512; i++ {
			Waveforms[6][i] = 1
			Waveforms[6][512+i] = -1
		}
		squareTable := Waveforms[6]

		// SQR   | 6:▀▄ 14:▀-
		Waveforms[14] = copyHalf(squareTable)
		// SQRx2 | 22:▘▘ 30:▘-
		Waveforms[22] = copyOct(squareTable)
		Waveforms[30] = copyOct(Waveforms[14])

		// ==================================================
		// exponential
		for i := 0; i < 512; i++ {
			x := float64(i) * 16 / 256
			Waveforms[7][i] = math.Pow(2, -x)
			Waveforms[7][1023-i] = -math.Pow(2, -(x + 1/16))
		}

		// ==================================================
		// clipped sinewave
		for i := 0; i < 1024; i++ {
			theta := 2 * math.Pi * float64(i) / 1024
			Waveforms[8][i] = math.Max(-1, math.Min(math.Sin(theta)*math.Sqrt2, 1))
		}
		csineTable := Waveforms[8]

		// CSIN  | 8:▀▄ 9:^- 10:^^ 11:''
		Waveforms[9] = copyHalf(csineTable)
		Waveforms[10] = copyAbs(csineTable)
		Waveforms[11] = copyAbsQuarter(csineTable)
		// CSINx2| 12:▚- 13:"-
		Waveforms[12] = copyOct(csineTable)
		Waveforms[13] = copyAbsOct(csineTable)

		// ==================================================
		// triangle wave
		for i := 0; i < 256; i++ {
			Waveforms[16][i] = float64(i) / 256
			Waveforms[16][256+i] = (256 - float64(i)) / 256
			Waveforms[16][512+i] = -float64(i) / 256
			Waveforms[16][768+i] = -(256 - float64(i)) / 256
		}
		triTable := Waveforms[16]

		// TRI   | 16:▀▄ 17:^- 18:^^ 19:''
		Waveforms[17] = copyHalf(triTable)
		Waveforms[18] = copyAbs(triTable)
		Waveforms[19] = copyAbsQuarter(triTable)
		// TRIx2 | 20:▚- 21:"-
		Waveforms[20] = copyOct(triTable)
		Waveforms[21] = copyAbsOct(triTable)

		// ==================================================
		// saw wave
		for i := 0; i < 512; i++ {
			Waveforms[24][i] = float64(i) / 512
			Waveforms[24][i+512] = float64(i)/512 - 1
		}
		sawTable := Waveforms[24]

		// SAW   | 24:▀▄ 25:▀- 26:▀▀ 27:''
		Waveforms[25] = copyHalf(sawTable)
		Waveforms[26] = copyAbs(sawTable)
		Waveforms[27] = copyAbsQuarter(sawTable)
		// SAWx2 | 28:▚- 29:"-
		Waveforms[28] = copyOct(sawTable)
		Waveforms[29] = copyAbsOct(sawTable)

	}

}

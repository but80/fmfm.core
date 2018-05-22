package ymfdata

import (
	"math"
)

// Frac64 は、0 以上 1 未満の固定小数点数を符号なし64ビット整数で表現する型です。
type Frac64 uint64

// FloatToFrac64 は、float64 から Frac64 に値を変換します。
func FloatToFrac64(v float64) Frac64 {
	return Frac64(v * Pow64Of2)
}

// MulUint64 は、Frac64 に uint64 型の値を掛けた値を返します。
func (v Frac64) MulUint64(rhs uint64) Frac64 {
	return v * Frac64(rhs)
}

// MulInt32Frac32 は、Frac64 に Int32Frac32 型の値を掛けた値を返します。
func (v Frac64) MulInt32Frac32(rhs Int32Frac32) Frac64 {
	return (v >> 32) * Frac64(rhs)
}

// Int32Frac32 は、0 以上 2^32 未満の固定小数点数を符号なし64ビット整数で表現する型です。
type Int32Frac32 uint64

// DebugDumpFPS は、デバッグとしてダンプ表示を行う頻度 [FPS] です。
const DebugDumpFPS = 30

// ChannelCount は、最大チャンネル数です。
const ChannelCount = 16

// SampleRate は、内部的なサンプルレート[Hz]です。
const SampleRate = float64(48000)

// A3Note は、MIDIメッセージにおけるA3のノートナンバーです。
const A3Note = 9 + 12*4

// A3Freq は、A3の周波数[Hz]です。
const A3Freq = float64(440.0)

// FNUMCoef は、周波数とFNUMを相互に変換する際に使用する係数です。
const FNUMCoef = float64(1<<19) / SampleRate * .5

// Pow32Of2 は、2の32乗です。
var Pow32Of2 = float64(1 << 32)

// Pow63Of2 は、2の63乗です。
var Pow63Of2 = float64(1 << 63)

// Pow64Of2 は、2の64乗です。
var Pow64Of2 = Pow63Of2 * 2.0

// ModulatorMultiplier は、モジュレータの出力を他のオペレータに入力する際の増幅率です。
const ModulatorMultiplier = 4.0

// ModulatorMatrix は、各 ALG でモジュレータとして使用されるオペレータを表すマトリクスです。
var ModulatorMatrix = [8][4]bool{
	{true, false, false, false},
	{false, false, false, false},
	{false, false, false, false},
	{true, true, true, false},
	{true, true, true, false},
	{true, false, true, false},
	{false, true, true, false},
	{false, true, false, false},
}

// CarrierMatrix は、各 ALG でキャリアとして使用されるオペレータを表すマトリクスです。
var CarrierMatrix = [8][4]bool{
	{false, true, false, false},
	{true, true, false, false},
	{true, true, true, true},
	{false, false, false, true},
	{false, false, false, true},
	{false, true, false, true},
	{true, false, false, true},
	{true, false, true, true},
}

// VolumeTable は、MIDIメッセージのボリュームやエクスプレッションによって振幅にかかる係数のテーブルです。
var VolumeTable = [...]float64{
	1e30, 47.9, 42.6, 37.2, 33.1, 29.8, 27.0, 24.6,
	22.4, 20.6, 18.9, 17.3, 15.9, 14.6, 13.4, 12.2,
	11.1, 10.1, 9.2, 8.3, 7.4, 6.6, 5.8, 5.1,
	4.4, 3.6, 3.0, 2.3, 1.7, 1.1, 0.6, 0.0,
}

// PanTable は、MIDIメッセージのパンによって左右それぞれの振幅にかかる係数のテーブルです。
var PanTable [128][2]float64

// DTCoef は、DTパラメータ および BLOCKとFNUM上位1ビットによって加わる周波数差分[Hz]のテーブルです。
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

// LFOFrequency は、LFOパラメータによって決まるビブラートやトレモロの周波数のテーブルです。
// 単位は、2の64乗を1周とする1サンプルあたりの増分です。
var LFOFrequency = [4]Frac64{}

// ModTableLen は、モジュレーション（ビブラートやトレモロ）の振幅テーブルの長さです。
const ModTableLen = 8192

// ModTableLenBits は、モジュレーション振幅テーブルのインデックスに必要なビット数です。
// 2 の ModTableLenBits 乗が ModTableLen になります。
const ModTableLenBits = 13

// ModTableIndexShift は、2の64乗を1周とする値からモジュレーション振幅テーブルの
// インデックスに変換する際、右シフトするビット数です。
const ModTableIndexShift = 64 - ModTableLenBits

// VibratoTableInt32Frac32 は、ビブラート（DVB）によって周波数にかかる係数のテーブルです。
// 整数部32bit・小数部32bitで表されます。
var VibratoTableInt32Frac32 [4][ModTableLen]Int32Frac32

// TremoloTable は、トレモロ（DAM）によって振幅にかかる係数のテーブルです。
var TremoloTable [4][ModTableLen]float64

// FeedbackTable は、FBパラメータによってフィードバックされる信号の振幅にかかる係数のテーブルです。
var FeedbackTable = [8]float64{0, 1 / 32, 1 / 16, 1 / 8, 1 / 4, 1 / 2, 1, 2}

// MultTable2 は、MULTパラメータによって周波数にかかる係数のテーブルです。2で割って使用します。
var MultTable2 = [16]uint64{1, 1 * 2, 2 * 2, 3 * 2, 4 * 2, 5 * 2, 6 * 2, 7 * 2, 8 * 2, 9 * 2, 10 * 2, 10 * 2, 12 * 2, 12 * 2, 15 * 2, 15 * 2}

// KSLTable は、KSLパラメータによる振幅の減衰量のテーブルです。
// 添字は順に KSL, BLOCK, FNUM上位5bit です。
var KSLTable = [4][8][32]float64{}

// WaveformLen は、波形テーブルの長さです。
const WaveformLen = 1024

// WaveformLenBits は、波形テーブルのインデックスに必要なビット数です。
// 2 の WaveformLenBits 乗が WaveformLen になります。
const WaveformLenBits = 10

// WaveformIndexShift は、2の64乗を1周とする値から波形テーブルの
// インデックスに変換する際、右シフトするビット数です。
const WaveformIndexShift = 64 - WaveformLenBits

// Waveforms は、波形テーブルです。
var Waveforms [32][]float64

func calculateIncrement(begin, end, period float64) float64 {
	return (end - begin) / SampleRate * (1 / period)
}

func triSin(phase float64) float64 {
	phase *= 4.0
	if phase < 1.0 {
		return phase
	}
	if phase < 3.0 {
		return 2.0 - phase
	}
	return phase - 4.0
}

func triCos(phase float64) float64 {
	phase *= 4.0
	if phase < 2.0 {
		return 1.0 - phase
	}
	return phase - 3.0
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
	// https://github.com/yamaha-webmusic/ymf825board/blob/991485a4cbbe07d84cca707701999875fbc17c74/manual/fbd_spec3.md#dam-eam-dvb-evb
	vibratoDepth := [4]float64{3.4, 6.7, 13.5, 26.8}
	for dvb := 0; dvb < 4; dvb++ {
		for i := 0; i < ModTableLen; i++ {
			phase := float64(i) / float64(ModTableLen)
			cent := triSin(phase) * vibratoDepth[dvb]
			v := math.Pow(2.0, cent/1200)
			VibratoTableInt32Frac32[dvb][i] = Int32Frac32(v * Pow32Of2)
		}
	}

	// generate tremolo table
	// https://github.com/yamaha-webmusic/ymf825board/blob/991485a4cbbe07d84cca707701999875fbc17c74/manual/fbd_spec3.md#dam-eam-dvb-evb
	tremoloDepth := [4]float64{1.3, 2.8, 5.8, 11.8} // dB
	for dam := 0; dam < 4; dam++ {
		for i := 0; i < ModTableLen; i++ {
			phase := float64(i) / float64(ModTableLen)
			v := (triCos(phase) - 1.0) * .5 * tremoloDepth[dam]
			TremoloTable[dam][i] = math.Pow(10.0, v/20.0)
		}
	}

	// generate KSL table
	kslBases := [4]float64{.0, .08, 1.0 / 15.0, 1.0 / 15.0}
	kslBlockCoefs := [4]float64{.0, 3.0, 1.5, 6.01}
	kslFnum5Coefs := [4]float64{.0, .38, .185, .75}
	for ksl := 0; ksl < 4; ksl++ {
		for block := 0; block < 8; block++ {
			for fnum5 := 0; fnum5 < 32; fnum5++ {
				fnum5lim := fnum5
				if 15 < fnum5lim {
					fnum5lim = 15
				}
				v := kslBases[ksl] - kslBlockCoefs[ksl]*float64(block-2) - kslFnum5Coefs[ksl]*float64(fnum5lim-7)
				if block < 2 || .0 <= v {
					v = 0
				}
				KSLTable[ksl][block][fnum5] = math.Pow(10.0, v/20.0)
			}
		}
	}

	// convert LFO frequency
	{
		lfoFreqHz := [4]float64{1.8, 4.0, 5.9, 7.0}
		for i, hz := range lfoFreqHz {
			LFOFrequency[i] = Frac64(hz / SampleRate * Pow64Of2)
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
			Waveforms[i] = make([]float64, WaveformLen)
		}

		// ^v to ^-
		copyHalf := func(src []float64) []float64 {
			dst := make([]float64, WaveformLen)
			for i := 0; i < 512; i++ {
				dst[i] = src[i]
				dst[512+i] = 0
			}
			return dst
		}

		// ^v to ^^
		copyAbs := func(src []float64) []float64 {
			dst := make([]float64, WaveformLen)
			for i := 0; i < 512; i++ {
				dst[i] = src[i]
				dst[512+i] = src[i]
			}
			return dst
		}

		// ^v to ''
		copyAbsQuarter := func(src []float64) []float64 {
			dst := make([]float64, WaveformLen)
			for i := 0; i < 256; i++ {
				dst[i] = src[i]
				dst[256+i] = 0
				dst[512+i] = src[i]
				dst[768+i] = 0
			}
			return dst
		}

		// ^v to ▚-
		copyOct := func(src []float64) []float64 {
			dst := make([]float64, WaveformLen)
			for i := 0; i < 512; i++ {
				dst[i] = src[i*2]
				dst[512+i] = 0
			}
			return dst
		}

		// ^v to "-
		copyAbsOct := func(src []float64) []float64 {
			dst := make([]float64, WaveformLen)
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
		for i := 0; i < WaveformLen; i++ {
			Waveforms[0][i] = math.Sin(2 * math.Pi * float64(i) / WaveformLen)
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
		Waveforms[22] = copyAbsQuarter(squareTable)
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
		for i := 0; i < WaveformLen; i++ {
			theta := 2 * math.Pi * float64(i) / WaveformLen
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

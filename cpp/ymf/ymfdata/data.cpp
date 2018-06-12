#pragma once
#include "./data.h"

namespace ymf {
namespace ymfdata {



// FloatToFrac64 は、float64 から Frac64 に値を変換します。
Frac64 FloatToFrac64(float64 v) {
	return Frac64(v*Pow64Of2);
}

// MulUint64 は、Frac64 に uint64 型の値を掛けた値を返します。
Frac64 Frac64__MulUint64(Frac64 v, uint64 rhs) {
	return v*Frac64(rhs);
}

// MulInt32Frac32 は、Frac64 に Int32Frac32 型の値を掛けた値を返します。
Frac64 Frac64__MulInt32Frac32(Frac64 v, Int32Frac32 rhs) {
	return (v >> 32)*Frac64(rhs);
}


#define DebugDumpFPS (30)

#define ChannelCount (32)

#define SampleRate (float64(48000))

#define A3Note (9 + 12*4)

#define A3Freq (float64(440.0))

#define FNUMCoef (float64(1 << 19)/SampleRate*.5)

auto Pow32Of2 = float64(1 << 32);

auto Pow63Of2 = float64(1 << 63);

auto Pow64Of2 = Pow63Of2*2.0;

#define ModulatorMultiplier (4.0)

auto ModulatorMatrix = {
	(const bool[4]){
		true,
		false,
		false,
		false,
	},
	(const bool[4]){
		false,
		false,
		false,
		false,
	},
	(const bool[4]){
		false,
		false,
		false,
		false,
	},
	(const bool[4]){
		true,
		true,
		true,
		false,
	},
	(const bool[4]){
		true,
		true,
		true,
		false,
	},
	(const bool[4]){
		true,
		false,
		true,
		false,
	},
	(const bool[4]){
		false,
		true,
		true,
		false,
	},
	(const bool[4]){
		false,
		true,
		false,
		false,
	},
};

auto CarrierMatrix = {
	(const bool[4]){
		false,
		true,
		false,
		false,
	},
	(const bool[4]){
		true,
		true,
		false,
		false,
	},
	(const bool[4]){
		true,
		true,
		true,
		true,
	},
	(const bool[4]){
		false,
		false,
		false,
		true,
	},
	(const bool[4]){
		false,
		false,
		false,
		true,
	},
	(const bool[4]){
		false,
		true,
		false,
		true,
	},
	(const bool[4]){
		true,
		false,
		false,
		true,
	},
	(const bool[4]){
		true,
		false,
		true,
		true,
	},
};

auto VolumeTable = {
	1e30,
	47.9,
	42.6,
	37.2,
	33.1,
	29.8,
	27.0,
	24.6,
	22.4,
	20.6,
	18.9,
	17.3,
	15.9,
	14.6,
	13.4,
	12.2,
	11.1,
	10.1,
	9.2,
	8.3,
	7.4,
	6.6,
	5.8,
	5.1,
	4.4,
	3.6,
	3.0,
	2.3,
	1.7,
	1.1,
	0.6,
	0.0,
};

float64 PanTable[2][128];

auto DTCoef = {
	(const float64[16]){
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
	},
	(const float64[16]){
		0.00,
		0.00,
		0.05,
		0.05,
		0.05,
		0.05,
		0.09,
		0.09,
		0.14,
		0.14,
		0.18,
		0.23,
		0.27,
		0.32,
		0.37,
		0.37,
	},
	(const float64[16]){
		0.05,
		0.05,
		0.09,
		0.09,
		0.14,
		0.14,
		0.18,
		0.23,
		0.27,
		0.32,
		0.41,
		0.46,
		0.59,
		0.64,
		0.73,
		0.73,
	},
	(const float64[16]){
		0.09,
		0.09,
		0.14,
		0.14,
		0.18,
		0.23,
		0.28,
		0.32,
		0.41,
		0.46,
		0.59,
		0.64,
		0.87,
		0.91,
		1.00,
		1.00,
	},
	(const float64[16]){
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
		0.00,
	},
	(const float64[16]){
		-0.00,
		-0.00,
		-0.05,
		-0.05,
		-0.05,
		-0.05,
		-0.09,
		-0.09,
		-0.14,
		-0.14,
		-0.18,
		-0.23,
		-0.27,
		-0.32,
		-0.37,
		-0.37,
	},
	(const float64[16]){
		-0.05,
		-0.05,
		-0.09,
		-0.09,
		-0.14,
		-0.14,
		-0.18,
		-0.23,
		-0.27,
		-0.32,
		-0.41,
		-0.46,
		-0.59,
		-0.64,
		-0.73,
		-0.73,
	},
	(const float64[16]){
		-0.09,
		-0.09,
		-0.14,
		-0.14,
		-0.18,
		-0.23,
		-0.28,
		-0.32,
		-0.41,
		-0.46,
		-0.59,
		-0.64,
		-0.87,
		-0.91,
		-1.00,
		-1.00,
	},
};

auto LFOFrequency = {};

#define ModTableLen (8192)

#define ModTableLenBits (13)

#define ModTableIndexShift (64 - ModTableLenBits)

Int32Frac32 VibratoTableInt32Frac32[8192][4];

float64 TremoloTable[8192][4];

auto FeedbackTable = {
	0,
	1.0/32.0,
	1.0/16.0,
	1.0/8.0,
	1.0/4.0,
	1.0/2.0,
	1.0,
	2.0,
};

auto MultTable2 = {
	1,
	1*2,
	2*2,
	3*2,
	4*2,
	5*2,
	6*2,
	7*2,
	8*2,
	9*2,
	10*2,
	10*2,
	12*2,
	12*2,
	15*2,
	15*2,
};

auto KSLTable = {};

#define WaveformLen (1024)

#define WaveformLenBits (10)

#define WaveformIndexShift (64 - WaveformLenBits)

vector<float64> Waveforms[32];

float64 calculateIncrement(float64 begin, float64 end, float64 period) {
	return (end - begin)/SampleRate*(1.0/period);
}

float64 triSin(float64 phase) {
	phase = 4.0;
	if (phase < 1.0) {
		return phase;
	}
	if (phase < 3.0) {
		return 2.0 - phase;
	}
	return phase - 4.0;
}

float64 triCos(float64 phase) {
	phase = 4.0;
	if (phase < 2.0) {
		return 1.0 - phase;
	}
	return phase - 3.0;
}

void init() {
	for (int i = 0; i < sizeof(VolumeTable) / sizeof(VolumeTable[0]); i++) {
		VolumeTable[i] = math::Pow(10.0, -VolumeTable[i]/20.0);
	}
	VolumeTable[0] = .0;
	auto i = 0;
	while i < 128 {
		auto a = math->Pi*.5*float64(i)/127.0;
		PanTable[i][0] = math::Cos(a);
		PanTable[i][1] = math::Sin(a);
		i++
	}
	auto vibratoDepth = {
		3.4,
		6.7,
		13.5,
		26.8,
	};
	auto dvb = 0;
	while dvb < 4 {
		auto i = 0;
		while i < ModTableLen {
			auto phase = float64(i)/float64(ModTableLen);
			auto cent = triSin(phase)*vibratoDepth[dvb];
			auto v = math::Pow(2.0, cent/1200.0);
			VibratoTableInt32Frac32[dvb][i] = Int32Frac32(v*Pow32Of2);
			i++
		}
		dvb++
	}
	auto tremoloDepth = {
		1.3,
		2.8,
		5.8,
		11.8,
	};
	auto dam = 0;
	while dam < 4 {
		auto i = 0;
		while i < ModTableLen {
			auto phase = float64(i)/float64(ModTableLen);
			auto v = (triCos(phase) - 1.0)*.5*tremoloDepth[dam];
			TremoloTable[dam][i] = math::Pow(10.0, v/20.0);
			i++
		}
		dam++
	}
	auto kslBases = {
		.0,
		.08,
		1.0/15.0,
		1.0/15.0,
	};
	auto kslBlockCoefs = {
		.0,
		3.0,
		1.5,
		6.01,
	};
	auto kslFnum5Coefs = {
		.0,
		.38,
		.185,
		.75,
	};
	auto ksl = 0;
	while ksl < 4 {
		auto block = 0;
		while block < 8 {
			auto fnum5 = 0;
			while fnum5 < 32 {
				auto fnum5lim = fnum5;
				if (15 < fnum5lim) {
					fnum5lim = 15;
				}
				auto v = kslBases[ksl] - kslBlockCoefs[ksl]*float64(block - 2) - kslFnum5Coefs[ksl]*float64(fnum5lim - 7);
				if (block < 2 || .0 <= v) {
					v = .0;
				}
				KSLTable[ksl][block][fnum5] = math::Pow(10.0, v/20.0);
				fnum5++
			}
			block++
		}
		ksl++
	}
	auto lfoFreqHz = {
		1.8,
		4.0,
		5.9,
		7.0,
	};
	for (int i = 0; i < sizeof(lfoFreqHz) / sizeof(lfoFreqHz[0]); i++) {
		auto hz = lfoFreqHz[i];
		LFOFrequency[i] = Frac64(hz/SampleRate*Pow64Of2);
	}
	for (int i = 0; i < sizeof(Waveforms) / sizeof(Waveforms[0]); i++) {
		Waveforms[i] = make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:9535, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc42000a100)})[0m, WaveformLen);
	}
	auto copyHalf = 	vector<float64> UNKNOWN(vector<float64> src) {
		auto dst = make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:9639, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc42000a320)})[0m, WaveformLen);
		auto i = 0;
		while i < 512 {
			dst[i] = src[i];
			dst[512 + i] = 0;
			i++
		}
		return dst;
	};
	auto copyAbs = 	vector<float64> UNKNOWN(vector<float64> src) {
		auto dst = make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:9830, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc42000a720)})[0m, WaveformLen);
		auto i = 0;
		while i < 512 {
			dst[i] = src[i];
			dst[512 + i] = src[i];
			i++
		}
		return dst;
	};
	auto copyAbsQuarter = 	vector<float64> UNKNOWN(vector<float64> src) {
		auto dst = make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:10033, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc42000ab00)})[0m, WaveformLen);
		auto i = 0;
		while i < 256 {
			dst[i] = src[i];
			dst[256 + i] = .0;
			dst[512 + i] = src[i];
			dst[768 + i] = .0;
			i++
		}
		return dst;
	};
	auto copyOct = 	vector<float64> UNKNOWN(vector<float64> src) {
		auto dst = make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:10271, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc42000afe0)})[0m, WaveformLen);
		auto i = 0;
		while i < 512 {
			dst[i] = src[i*2];
			dst[512 + i] = .0;
			i++
		}
		return dst;
	};
	auto copyAbsOct = 	vector<float64> UNKNOWN(vector<float64> src) {
		auto dst = make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:10468, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc42000b3c0)})[0m, WaveformLen);
		auto i = 0;
		while i < 256 {
			dst[i] = src[i*2];
			dst[256 + i] = src[i*2];
			dst[512 + i] = .0;
			dst[768 + i] = .0;
			i++
		}
		return dst;
	};
	auto i = 0;
	while i < WaveformLen {
		Waveforms[0][i] = math::Sin(2*math->Pi*float64(i)/WaveformLen);
		i++
	}
	auto sineTable = Waveforms[0];
	Waveforms[1] = copyHalf(sineTable);
	Waveforms[2] = copyAbs(sineTable);
	Waveforms[3] = copyAbsQuarter(sineTable);
	Waveforms[4] = copyOct(sineTable);
	Waveforms[5] = copyAbsOct(sineTable);
	auto i = 0;
	while i < 512 {
		Waveforms[6][i] = 1.0;
		Waveforms[6][512 + i] = -1.0;
		i++
	}
	auto squareTable = Waveforms[6];
	Waveforms[14] = copyHalf(squareTable);
	Waveforms[22] = copyAbsQuarter(squareTable);
	Waveforms[30] = copyOct(Waveforms[14]);
	auto i = 0;
	while i < 512 {
		auto x = float64(i)*16.0/256.0;
		Waveforms[7][i] = math::Pow(2.0, -x);
		Waveforms[7][1023 - i] = -math::Pow(2.0, -(x + 1.0/16.0));
		i++
	}
	auto i = 0;
	while i < WaveformLen {
		auto theta = 2*math->Pi*float64(i)/WaveformLen;
		Waveforms[8][i] = math::Max(-1.0, math::Min(math::Sin(theta)*math->Sqrt2, 1.0));
		i++
	}
	auto csineTable = Waveforms[8];
	Waveforms[9] = copyHalf(csineTable);
	Waveforms[10] = copyAbs(csineTable);
	Waveforms[11] = copyAbsQuarter(csineTable);
	Waveforms[12] = copyOct(csineTable);
	Waveforms[13] = copyAbsOct(csineTable);
	auto i = 0;
	while i < 256 {
		Waveforms[16][i] = float64(i)/256.0;
		Waveforms[16][256 + i] = (256.0 - float64(i))/256.0;
		Waveforms[16][512 + i] = -float64(i)/256.0;
		Waveforms[16][768 + i] = -(256.0 - float64(i))/256.0;
		i++
	}
	auto triTable = Waveforms[16];
	Waveforms[17] = copyHalf(triTable);
	Waveforms[18] = copyAbs(triTable);
	Waveforms[19] = copyAbsQuarter(triTable);
	Waveforms[20] = copyOct(triTable);
	Waveforms[21] = copyAbsOct(triTable);
	auto i = 0;
	while i < 512 {
		Waveforms[24][i] = float64(i)/512.0;
		Waveforms[24][i + 512] = float64(i)/512.0 - 1.0;
		i++
	}
	auto sawTable = Waveforms[24];
	Waveforms[25] = copyHalf(sawTable);
	Waveforms[26] = copyAbs(sawTable);
	Waveforms[27] = copyAbsQuarter(sawTable);
	Waveforms[28] = copyOct(sawTable);
	Waveforms[29] = copyAbsOct(sawTable);
}

}
}

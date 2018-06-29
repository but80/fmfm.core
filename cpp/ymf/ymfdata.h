// namespace ymf::ymfdata
#pragma once
#include "go2cpp.h"
#include "math.h"
namespace ymf {
namespace ymfdata {

typedef uint64 Frac64;
typedef uint64 Int32Frac32;
/* map[string]string{"math":"math"} */
Frac64 FloatToFrac64(float64 v);
Frac64 Frac64__MulUint64(Frac64 v, uint64 rhs);
Frac64 Frac64__MulInt32Frac32(Frac64 v, Int32Frac32 rhs);
/* map[string]string{"math":"math"} */
const auto DebugDumpFPS = 30;
const auto ChannelCount = 32;
const auto SampleRate = float64(48000);
const auto A3Note = 9 + 12*4;
const auto A3Freq = float64(440.0);
const auto FNUMCoef = float64(1 << 19)/SampleRate*.5;
extern float64 Pow32Of2;
extern float64 Pow63Of2;
extern float64 Pow64Of2;
const auto ModulatorMultiplier = 4.0;
extern bool ModulatorMatrix[4][8];
extern bool CarrierMatrix[4][8];
extern float64 VolumeTable[32];
extern float64 PanTable[2][128];
extern float64 DTCoef[16][8];
extern Frac64 LFOFrequency[4];
const auto ModTableLen = 8192;
const auto ModTableLenBits = 13;
const auto ModTableIndexShift = 64 - ModTableLenBits;
extern Int32Frac32 VibratoTableInt32Frac32[8192][4];
extern float64 TremoloTable[8192][4];
extern float64 FeedbackTable[8];
extern uint64 MultTable2[16];
extern float64 KSLTable[32][8][4];
const auto WaveformLen = 1024;
const auto WaveformLenBits = 10;
const auto WaveformIndexShift = 64 - WaveformLenBits;
extern std::vector<float64> Waveforms[32];
float64 calculateIncrement(float64 begin, float64 end, float64 period);
float64 triSin(float64 phase);
float64 triCos(float64 phase);
void init();
}
}

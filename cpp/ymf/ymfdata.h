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
const auto ModulatorMultiplier = 4.0;
const auto ModTableLen = 8192;
const auto ModTableLenBits = 13;
const auto ModTableIndexShift = 64 - ModTableLenBits;
const auto WaveformLen = 1024;
const auto WaveformLenBits = 10;
const auto WaveformIndexShift = 64 - WaveformLenBits;
float64 calculateIncrement(float64 begin, float64 end, float64 period);
float64 triSin(float64 phase);
float64 triCos(float64 phase);
void init();
}
}

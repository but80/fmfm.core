#pragma once
#include "go2cpp.h"
namespace ymf {
namespace ymfdata {

#include "math.h"
typedef uint64 Frac64;
Frac64 FloatToFrac64(float64 v);
Frac64 Frac64__MulUint64(Frac64 v, uint64 rhs);
Frac64 Frac64__MulInt32Frac32(Frac64 v, Int32Frac32 rhs);
typedef uint64 Int32Frac32;
float64 calculateIncrement(float64 begin, float64 end, float64 period);
float64 triSin(float64 phase);
float64 triCos(float64 phase);
void init();
}
}

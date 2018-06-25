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
float64 calculateIncrement(float64 begin, float64 end, float64 period);
float64 triSin(float64 phase);
float64 triCos(float64 phase);
void init();
}
}

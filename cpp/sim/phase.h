#pragma once
#include "ymf/ymfdata.h"
namespace ymfdata = ymf::ymfdata;
typedef struct __phaseGenerator {
	float64 sampleRate;
	bool evb;
	int dvb;
	gopkg_in::but80::fmfm_core_v1::ymf::ymfdata::Frac64 phaseFrac64;
	gopkg_in::but80::fmfm_core_v1::ymf::ymfdata::Frac64 phaseIncrementFrac64;
} phaseGenerator;

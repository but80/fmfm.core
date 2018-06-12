#pragma once
#include "./phase.h"

namespace sim {



phaseGenerator *newPhaseGenerator(float64 sampleRate) {
	auto pg = __ptr((const phaseGenerator){
		sampleRate: sampleRate,
	});
	phaseGeneratorPtr__reset(pg);
	return pg;
}

void phaseGeneratorPtr__reset(phaseGenerator *pg) {
	pg->phaseFrac64 = 0;
}

void phaseGeneratorPtr__resetAll(phaseGenerator *pg) {
	pg->evb = false;
	pg->dvb = 0;
	pg->phaseIncrementFrac64 = 0;
	phaseGeneratorPtr__reset(pg);
}

void phaseGeneratorPtr__setFrequency(phaseGenerator *pg, int fnum, int block, int bo, int mult, int dt) {
	auto baseFrequency = float64(fnum << uint(block + 3 - bo))/(16.0*ymfdata->FNUMCoef);
	auto ksn = block << 1 | fnum >> 9;
	auto operatorFrequency = baseFrequency + ymfdata->DTCoef[dt][ksn];
	pg->phaseIncrementFrac64 = ymfdata::FloatToFrac64(operatorFrequency/pg->sampleRate);
	pg->phaseIncrementFrac64 = gopkg_in::but80::fmfm_core_v1::ymf::ymfdata::Frac64__MulUint64(pg->phaseIncrementFrac64, ymfdata->MultTable2[mult]);
	pg->phaseIncrementFrac64 = 1;
}

gopkg_in::but80::fmfm_core_v1::ymf::ymfdata::Frac64 phaseGeneratorPtr__getPhase(phaseGenerator *pg, int vibratoIndex) {
	if (pg->evb) {
		pg->phaseFrac64 = gopkg_in::but80::fmfm_core_v1::ymf::ymfdata::Frac64__MulInt32Frac32(pg->phaseIncrementFrac64, ymfdata->VibratoTableInt32Frac32[pg->dvb][vibratoIndex]);
	} else {
		pg->phaseFrac64 = pg->phaseIncrementFrac64;
	}
	return pg->phaseFrac64;
}

}

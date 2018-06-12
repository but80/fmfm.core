#pragma once
#include "./envelope.h"

namespace sim {



#define stageOff stage (0)
#define stageAttack (1)
#define stageDecay (2)
#define stageSustain (3)
#define stageRelease (4)

string stage__String(stage s) {
	auto __tag = s;
	if (__tag == stageOff) {
		return "-";
	} else if (__tag == stageAttack) {
		return "A";
	} else if (__tag == stageDecay) {
		return "D";
	} else if (__tag == stageSustain) {
		return "S";
	} else if (__tag == stageRelease) {
		return "R";
	} else {
		return "?";
	}
}

#define epsilon (1.0/32768.0)


envelopeGenerator *newEnvelopeGenerator(float64 sampleRate) {
	auto eg = __ptr((const envelopeGenerator){
		sampleRate: sampleRate,
	});
	envelopeGeneratorPtr__resetAll(eg);
	return eg;
}

void envelopeGeneratorPtr__reset(envelopeGenerator *eg) {
	eg->currentLevel = .0;
	eg->stage = stageOff;
}

void envelopeGeneratorPtr__resetAll(envelopeGenerator *eg) {
	eg->eam = false;
	eg->dam = 0;
	eg->sustainLevel = .0;
	envelopeGeneratorPtr__setTotalLevel(eg, 63);
	envelopeGeneratorPtr__setKeyScalingLevel(eg, 0, 0, 1, 0);
	envelopeGeneratorPtr__reset(eg);
}

void envelopeGeneratorPtr__setActualSustainLevel(envelopeGenerator *eg, int sl) {
	if (sl == 0x0f) {
		eg->sustainLevel = 0;
	} else {
		auto slDB = -3.0*float64(sl);
		eg->sustainLevel = math::Pow(10.0, slDB/20.0);
	}
}

void envelopeGeneratorPtr__setTotalLevel(envelopeGenerator *eg, int tl) {
	if (63 <= tl) {
		eg->tlCoef = .0;
		eg->kslTlCoef = .0;
		return;
	}
	auto tlDB = float64(tl)*-0.75;
	eg->tlCoef = math::Pow(10.0, tlDB/20.0);
	eg->kslTlCoef = eg->kslCoef*eg->tlCoef;
}

void envelopeGeneratorPtr__setKeyScalingLevel(envelopeGenerator *eg, int fnum, int block, int bo, int ksl) {
	auto blkbo = block + 1 - bo;
	if (blkbo < 0) {
		blkbo = 0;
	} else {
		if (7 < blkbo) {
			blkbo = 7;
		}
	}
	eg->kslCoef = ymfdata->KSLTable[ksl][blkbo][fnum >> 5];
	eg->kslTlCoef = eg->kslCoef*eg->tlCoef;
}

void envelopeGeneratorPtr__setActualAR(envelopeGenerator *eg, int attackRate, int ksr, int keyScaleNumber) {
	if (attackRate <= 0) {
		eg->arDiffPerSample = .0;
		return;
	}
	auto ksn = (keyScaleNumber >> 1) + (keyScaleNumber & 1);
	auto sec = attackTimeSecAt1[ksr][ksn]/float64(uint(1) << uint(attackRate - 1));
	eg->arDiffPerSample = 1.0/(sec*eg->sampleRate);
}

void envelopeGeneratorPtr__setActualDR(envelopeGenerator *eg, int dr, int ksr, int keyScaleNumber) {
	if (dr == 0) {
		eg->drCoefPerSample = 1.0;
	} else {
		auto dbPerSecAt4 = decayDBPerSecAt4[ksr][keyScaleNumber]/2.0;
		auto dbPerSample = dbPerSecAt4*float64(uint(1) << uint(dr))/16.0/eg->sampleRate;
		eg->drCoefPerSample = math::Pow(10, -dbPerSample/10);
	}
}

void envelopeGeneratorPtr__setActualSR(envelopeGenerator *eg, int sr, int ksr, int keyScaleNumber) {
	if (sr == 0) {
		eg->srCoefPerSample = 1.0;
	} else {
		auto dbPerSecAt4 = decayDBPerSecAt4[ksr][keyScaleNumber]/2.0;
		auto dbPerSample = dbPerSecAt4*float64(uint(1) << uint(sr))/16.0/eg->sampleRate;
		eg->srCoefPerSample = math::Pow(10, -dbPerSample/10);
	}
}

void envelopeGeneratorPtr__setActualRR(envelopeGenerator *eg, int rr, int ksr, int keyScaleNumber) {
	if (rr == 0) {
		eg->rrCoefPerSample = 1.0;
	} else {
		auto dbPerSecAt4 = decayDBPerSecAt4[ksr][keyScaleNumber]/2.0;
		auto dbPerSample = dbPerSecAt4*float64(uint(1) << uint(rr))/16.0/eg->sampleRate;
		eg->rrCoefPerSample = math::Pow(10, -dbPerSample/10);
	}
}

float64 envelopeGeneratorPtr__getEnvelope(envelopeGenerator *eg, int tremoloIndex) {
	auto __tag = eg->stage;
	if (__tag == stageAttack) {
		eg->currentLevel = eg->arDiffPerSample;
		if (eg->currentLevel < 1.0) {
			break;
		}
		eg->currentLevel = 1.0;
		eg->stage = stageDecay;
		fallthrough;
	} else if (__tag == stageDecay) {
		if (eg->sustainLevel < eg->currentLevel) {
			eg->currentLevel = eg->drCoefPerSample;
			break;
		}
		eg->stage = stageSustain;
		fallthrough;
	} else if (__tag == stageSustain) {
		if (epsilon < eg->currentLevel) {
			eg->currentLevel = eg->srCoefPerSample;
		} else {
			eg->stage = stageOff;
		}
		break;
	} else if (__tag == stageRelease) {
		if (epsilon < eg->currentLevel) {
			eg->currentLevel = eg->rrCoefPerSample;
		} else {
			eg->currentLevel = .0;
			eg->stage = stageOff;
		}
		break;
	}
	auto result = eg->currentLevel;
	if (eg->eam) {
		result = ymfdata->TremoloTable[eg->dam][tremoloIndex];
	}
	return result*eg->kslTlCoef;
}

void envelopeGeneratorPtr__keyOn(envelopeGenerator *eg) {
	eg->stage = stageAttack;
}

void envelopeGeneratorPtr__keyOff(envelopeGenerator *eg) {
	if (eg->stage != stageOff) {
		eg->stage = stageRelease;
	}
}

auto decayDBPerSecAt4 = {
	(const float64[16]){
		17.9342,
		17.9342,
		17.9342,
		17.9342,
		17.9342,
		22.4116,
		22.4116,
		22.4116,
		22.4116,
		26.9076,
		26.9076,
		26.9076,
		26.9076,
		31.3661,
		31.3661,
		31.3661,
	},
	(const float64[16]){
		17.9465,
		22.4376,
		22.4376,
		31.4026,
		31.4026,
		44.8696,
		44.8696,
		62.7959,
		62.7959,
		89.6707,
		89.6707,
		125.5546,
		125.5546,
		179.2684,
		179.2684,
		250.9128,
	},
};

auto attackTimeSecAt1 = {
	(const float64[9]){
		3.07068,
		3.07068,
		3.07068,
		2.45670,
		2.45670,
		2.04699,
		2.04699,
		1.75471,
		1.75471,
	},
	(const float64[9]){
		3.07082,
		2.45660,
		1.75489,
		1.22816,
		0.87737,
		0.61414,
		0.43876,
		0.30714,
		0.21935,
	},
};

}

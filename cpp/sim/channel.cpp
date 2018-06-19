#include "channel.h"

namespace sim {


#define noModulator (0)


Channel *newChannel(int channelID, Chip *chip) {
	auto ch = __ptr((const Channel){
		chip: chip,
		channelID: channelID,
	});
	ch->feedbackBlendCurr = .5*ymfdata->SampleRate/chip->sampleRate;
	if (1.0 < ch->feedbackBlendCurr) {
		ch->feedbackBlendCurr = 1.0;
	}
	ch->feedbackBlendPrev = 1.0 - ch->feedbackBlendCurr;
	for (int i = 0; i < sizeof(ch->operators) / sizeof(ch->operators[0]); i++) {
		ch->operators[i] = newOperator(channelID, i, chip);
	}
	ChannelPtr__resetAll(ch);
	return ch;
}

void ChannelPtr__reset(Channel *ch) {
	ch->modIndexFrac64 = 0;
	ch->feedback1Prev = .0;
	ch->feedback1Curr = .0;
	ch->feedback3Prev = .0;
	ch->feedback3Curr = .0;
	ch->feedbackOut1 = .0;
	ch->feedbackOut3 = .0;
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		phaseGeneratorPtr__reset(op->phaseGenerator);
		envelopeGeneratorPtr__reset(op->envelopeGenerator);
	}
}

void ChannelPtr__resetAll(Channel *ch) {
	ch->midiChannelID = -1;
	ch->fnum = 0;
	ch->kon = 0;
	ch->block = 0;
	ch->alg = 0;
	ch->panpot = 15;
	ch->chpan = 64;
	ch->volume = 100;
	ch->expression = 127;
	ch->velocity = 0;
	ch->bo = 1;
	ChannelPtr__setLFO(ch, 0);
	ChannelPtr__updatePanCoef(ch);
	ChannelPtr__updateAttenuation(ch);
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		operatorPtr__resetAll(op);
	}
}

bool ChannelPtr__isOff(Channel *ch) {
	for (int i = 0; i < sizeof(ch->operators) / sizeof(ch->operators[0]); i++) {
		auto op = ch->operators[i];
		if (!ymfdata->CarrierMatrix[ch->alg][i]) {
			continue;
		}
		if (op->envelopeGenerator->stage != stageOff) {
			return false;
		}
	}
	return true;
}

float64 ChannelPtr__currentLevel(Channel *ch) {
	auto result = .0;
	for (int i = 0; i < sizeof(ch->operators) / sizeof(ch->operators[0]); i++) {
		auto op = ch->operators[i];
		if (ymfdata->CarrierMatrix[ch->alg][i]) {
			auto eg = op->envelopeGenerator;
			auto v = eg->currentLevel*eg->kslTlCoef;
			if (result < v) {
				result = v;
			}
		}
	}
	return result;
}

string ChannelPtr__dump(Channel *ch) {
	auto lv = int((96.0 + math::Log10(ChannelPtr__currentLevel(ch))*20.0)/8.0);
	auto lvstr = strings::Repeat("|", lv);
	auto result = fmt::Sprintf("[%02d] midi=%02d alg=%d pan=%03d+%03d vol=%03d exp=%03d vel=%03d freq=%03d+%d-%d modidx=%04d %s\n", ch->channelID, ch->midiChannelID, ch->alg, ch->panpot, ch->chpan, ch->volume, ch->expression, ch->velocity, ch->fnum, ch->block, ch->bo, ch->modIndexFrac64 >> ymfdata->ModTableIndexShift, lvstr);
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		result = "  " + operatorPtr__dump(op) + "\n";
	}
	return result;
}

void ChannelPtr__setKON(Channel *ch, int v) {
	if (v == 0) {
		ChannelPtr__keyOff(ch);
		if (ChannelPtr__isOff(ch)) {
			ChannelPtr__resetAll(ch);
		}
	} else {
		ChannelPtr__keyOn(ch);
	}
}

void ChannelPtr__keyOn(Channel *ch) {
	if (ch->kon != 0) {
		return;
	}
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		operatorPtr__keyOn(op);
	}
	ch->kon = 1;
}

void ChannelPtr__keyOff(Channel *ch) {
	if (ch->kon == 0) {
		return;
	}
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		operatorPtr__keyOff(op);
	}
	ch->kon = 0;
}

void ChannelPtr__setBLOCK(Channel *ch, int v) {
	ch->block = v;
	ChannelPtr__updateFrequency(ch);
}

void ChannelPtr__setFNUM(Channel *ch, int v) {
	ch->fnum = v;
	ChannelPtr__updateFrequency(ch);
}

void ChannelPtr__setALG(Channel *ch, int v) {
	if (ch->alg != v) {
		ChannelPtr__reset(ch);
	}
	ch->alg = v;
	ch->feedback1Prev = 0;
	ch->feedback1Curr = 0;
	ch->feedback3Prev = 0;
	ch->feedback3Curr = 0;
	for (int i = 0; i < sizeof(ch->operators) / sizeof(ch->operators[0]); i++) {
		auto op = ch->operators[i];
		op->isModulator = ymfdata->ModulatorMatrix[ch->alg][i];
	}
}

void ChannelPtr__setLFO(Channel *ch, int v) {
	ch->lfoFrequency = ymfdata->LFOFrequency[v];
}

void ChannelPtr__setPANPOT(Channel *ch, int v) {
	ch->panpot = v;
	ChannelPtr__updatePanCoef(ch);
}

void ChannelPtr__setCHPAN(Channel *ch, int v) {
	ch->chpan = v;
	ChannelPtr__updatePanCoef(ch);
}

void ChannelPtr__updatePanCoef(Channel *ch) {
	auto pan = ch->chpan + (ch->panpot - 15)*4;
	if (pan < 0) {
		pan = 0;
	} else {
		if (127 < pan) {
			pan = 127;
		}
	}
	ch->panCoefL = ymfdata->PanTable[pan][0];
	ch->panCoefR = ymfdata->PanTable[pan][1];
}

void ChannelPtr__setVOLUME(Channel *ch, int v) {
	ch->volume = v;
	ChannelPtr__updateAttenuation(ch);
}

void ChannelPtr__setEXPRESSION(Channel *ch, int v) {
	ch->expression = v;
	ChannelPtr__updateAttenuation(ch);
}

void ChannelPtr__setVELOCITY(Channel *ch, int v) {
	ch->velocity = v;
	ChannelPtr__updateAttenuation(ch);
}

void ChannelPtr__updateAttenuation(Channel *ch) {
	ch->attenuationCoef = ymfdata->VolumeTable[ch->volume >> 2]*ymfdata->VolumeTable[ch->expression >> 2]*ymfdata->VolumeTable[ch->velocity >> 2];
}

void ChannelPtr__setBO(Channel *ch, int v) {
	ch->bo = v;
	ChannelPtr__updateFrequency(ch);
}

MULTIRESULT ChannelPtr__next(Channel *ch) {
	float64 result;
	float64 op1out;
	float64 op2out;
	float64 op3out;
	float64 op4out;
	auto op1 = ch->operators[0];
	auto op2 = ch->operators[1];
	auto op3 = ch->operators[2];
	auto op4 = ch->operators[3];
	auto modIndex = int(ch->modIndexFrac64 >> ymfdata->ModTableIndexShift);
	ch->modIndexFrac64 = ch->lfoFrequency;
	auto __tag = ch->alg;
	if (__tag == 0) {
		if (op2->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = operatorPtr__next(op1, modIndex, ch->feedbackOut1);
		result = operatorPtr__next(op2, modIndex, op1out*ymfdata->ModulatorMultiplier);
	} else if (__tag == 1) {
		if (op1->envelopeGenerator->stage == stageOff && op2->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = operatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = operatorPtr__next(op2, modIndex, noModulator);
		result = op1out + op2out;
	} else if (__tag == 2) {
		if (op1->envelopeGenerator->stage == stageOff && op2->envelopeGenerator->stage == stageOff && op3->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = operatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = operatorPtr__next(op2, modIndex, noModulator);
		op3out = operatorPtr__next(op3, modIndex, ch->feedbackOut3);
		op4out = operatorPtr__next(op4, modIndex, noModulator);
		result = op1out + op2out + op3out + op4out;
	} else if (__tag == 3) {
		if (op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = operatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = operatorPtr__next(op2, modIndex, noModulator);
		op3out = operatorPtr__next(op3, modIndex, op2out*ymfdata->ModulatorMultiplier);
		result = operatorPtr__next(op4, modIndex, (op1out + op3out)*ymfdata->ModulatorMultiplier);
	} else if (__tag == 4) {
		if (op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = operatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = operatorPtr__next(op2, modIndex, op1out*ymfdata->ModulatorMultiplier);
		op3out = operatorPtr__next(op3, modIndex, op2out*ymfdata->ModulatorMultiplier);
		result = operatorPtr__next(op4, modIndex, op3out*ymfdata->ModulatorMultiplier);
	} else if (__tag == 5) {
		if (op2->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = operatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = operatorPtr__next(op2, modIndex, op1out*ymfdata->ModulatorMultiplier);
		op3out = operatorPtr__next(op3, modIndex, ch->feedbackOut3);
		op4out = operatorPtr__next(op4, modIndex, op3out*ymfdata->ModulatorMultiplier);
		result = op2out + op4out;
	} else if (__tag == 6) {
		if (op1->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = operatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = operatorPtr__next(op2, modIndex, noModulator);
		op3out = operatorPtr__next(op3, modIndex, op2out*ymfdata->ModulatorMultiplier);
		op4out = operatorPtr__next(op4, modIndex, op3out*ymfdata->ModulatorMultiplier);
		result = op1out + op4out;
	} else if (__tag == 7) {
		if (op1->envelopeGenerator->stage == stageOff && op3->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = operatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = operatorPtr__next(op2, modIndex, noModulator);
		op3out = operatorPtr__next(op3, modIndex, op2out*ymfdata->ModulatorMultiplier);
		op4out = operatorPtr__next(op4, modIndex, noModulator);
		result = op1out + op3out + op4out;
	}
	if (op1->feedbackCoef != .0) {
		ch->feedback1Prev = ch->feedback1Curr;
		ch->feedback1Curr = op1out*op1->feedbackCoef;
		ch->feedbackOut1 = ch->feedback1Prev*ch->feedbackBlendPrev + ch->feedback1Curr*ch->feedbackBlendCurr;
	}
	if (op3->feedbackCoef != .0) {
		ch->feedback3Prev = ch->feedback3Curr;
		ch->feedback3Curr = op3out*op3->feedbackCoef;
		ch->feedbackOut3 = ch->feedback3Prev*ch->feedbackBlendPrev + ch->feedback3Curr*ch->feedbackBlendCurr;
	}
	result = ch->attenuationCoef;
	return result*ch->panCoefL, result*ch->panCoefR;
}

void ChannelPtr__updateFrequency(Channel *ch) {
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		operatorPtr__setFrequency(op, ch->fnum, ch->block, ch->bo);
	}
}

}

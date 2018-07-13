#include "sim.h"

namespace sim {




std::shared_ptr<Channel> newChannel(int channelID, std::shared_ptr<Chip> chip) {
	auto ch = __ptr((const Channel){
		channelID,
		int(0),
		chip,
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		{} /*gopkg.in/but80/fmfm.core.v1/ymf/ymfdata.Frac64*/,
		{} /*gopkg.in/but80/fmfm.core.v1/ymf/ymfdata.Frac64*/,
		float64(.0),
		float64(.0),
		{} /*[4]*gopkg.in/but80/fmfm.core.v1/sim.fmOperator*/,
	});
	ch->feedbackBlendCurr = .5*ymfdata::SampleRate/chip->sampleRate;
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

void ChannelPtr__reset(std::shared_ptr<Channel> ch) {
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

void ChannelPtr__resetAll(std::shared_ptr<Channel> ch) {
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
		fmOperatorPtr__resetAll(op);
	}
}

bool ChannelPtr__isOff(std::shared_ptr<Channel> ch) {
	for (int i = 0; i < sizeof(ch->operators) / sizeof(ch->operators[0]); i++) {
		auto op = ch->operators[i];
		if (!ymfdata::CarrierMatrix[ch->alg][i]) {
			continue;
		}
		if (op->envelopeGenerator->stage != stageOff) {
			return false;
		}
	}
	return true;
}

float64 ChannelPtr__currentLevel(std::shared_ptr<Channel> ch) {
	auto result = .0;
	for (int i = 0; i < sizeof(ch->operators) / sizeof(ch->operators[0]); i++) {
		auto op = ch->operators[i];
		if (ymfdata::CarrierMatrix[ch->alg][i]) {
			auto eg = op->envelopeGenerator;
			auto v = eg->currentLevel*eg->kslTlCoef;
			if (result < v) {
				result = v;
			}
		}
	}
	return result;
}

string ChannelPtr__dump(std::shared_ptr<Channel> ch) {
	auto lv = int((96.0 + math::Log10(ChannelPtr__currentLevel(ch))*20.0)/8.0);
	auto lvstr = strings::Repeat(string("|"), lv);
	auto result = fmt::Sprintf(string("[%02d] midi=%02d alg=%d pan=%03d+%03d vol=%03d exp=%03d vel=%03d freq=%03d+%d-%d modidx=%04d %s\n"), ch->channelID, ch->midiChannelID, ch->alg, ch->panpot, ch->chpan, ch->volume, ch->expression, ch->velocity, ch->fnum, ch->block, ch->bo, ch->modIndexFrac64 >> ymfdata::ModTableIndexShift, lvstr.c_str());
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		result = string("  ") + fmOperatorPtr__dump(op) + string("\n");
	}
	return result;
}

void ChannelPtr__setKON(std::shared_ptr<Channel> ch, int v) {
	if (v == 0) {
		ChannelPtr__keyOff(ch);
		if (ChannelPtr__isOff(ch)) {
			ChannelPtr__resetAll(ch);
		}
	} else {
		ChannelPtr__keyOn(ch);
	}
}

void ChannelPtr__keyOn(std::shared_ptr<Channel> ch) {
	if (ch->kon != 0) {
		return;
	}
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		fmOperatorPtr__keyOn(op);
	}
	ch->kon = 1;
}

void ChannelPtr__keyOff(std::shared_ptr<Channel> ch) {
	if (ch->kon == 0) {
		return;
	}
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		fmOperatorPtr__keyOff(op);
	}
	ch->kon = 0;
}

void ChannelPtr__setBLOCK(std::shared_ptr<Channel> ch, int v) {
	ch->block = v;
	ChannelPtr__updateFrequency(ch);
}

void ChannelPtr__setFNUM(std::shared_ptr<Channel> ch, int v) {
	ch->fnum = v;
	ChannelPtr__updateFrequency(ch);
}

void ChannelPtr__setALG(std::shared_ptr<Channel> ch, int v) {
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
		op->isModulator = ymfdata::ModulatorMatrix[ch->alg][i];
	}
}

void ChannelPtr__setLFO(std::shared_ptr<Channel> ch, int v) {
	ch->lfoFrequency = ymfdata::LFOFrequency[v];
}

void ChannelPtr__setPANPOT(std::shared_ptr<Channel> ch, int v) {
	ch->panpot = v;
	ChannelPtr__updatePanCoef(ch);
}

void ChannelPtr__setCHPAN(std::shared_ptr<Channel> ch, int v) {
	ch->chpan = v;
	ChannelPtr__updatePanCoef(ch);
}

void ChannelPtr__updatePanCoef(std::shared_ptr<Channel> ch) {
	auto pan = ch->chpan + (ch->panpot - 15)*4;
	if (pan < 0) {
		pan = 0;
	} else {
		if (127 < pan) {
			pan = 127;
		}
	}
	ch->panCoefL = ymfdata::PanTable[pan][0];
	ch->panCoefR = ymfdata::PanTable[pan][1];
}

void ChannelPtr__setVOLUME(std::shared_ptr<Channel> ch, int v) {
	ch->volume = v;
	ChannelPtr__updateAttenuation(ch);
}

void ChannelPtr__setEXPRESSION(std::shared_ptr<Channel> ch, int v) {
	ch->expression = v;
	ChannelPtr__updateAttenuation(ch);
}

void ChannelPtr__setVELOCITY(std::shared_ptr<Channel> ch, int v) {
	ch->velocity = v;
	ChannelPtr__updateAttenuation(ch);
}

void ChannelPtr__updateAttenuation(std::shared_ptr<Channel> ch) {
	ch->attenuationCoef = ymfdata::VolumeTable[ch->volume >> 2]*ymfdata::VolumeTable[ch->expression >> 2]*ymfdata::VolumeTable[ch->velocity >> 2];
}

void ChannelPtr__setBO(std::shared_ptr<Channel> ch, int v) {
	ch->bo = v;
	ChannelPtr__updateFrequency(ch);
}

ChannelPtr__next__result ChannelPtr__next(std::shared_ptr<Channel> ch) {
	ChannelPtr__next__result __result; // multi-result
	float64 result;
	float64 op1out;
	float64 op2out;
	float64 op3out;
	float64 op4out;
	auto op1 = ch->operators[0];
	auto op2 = ch->operators[1];
	auto op3 = ch->operators[2];
	auto op4 = ch->operators[3];
	auto modIndex = int(ch->modIndexFrac64 >> ymfdata::ModTableIndexShift);
	ch->modIndexFrac64 = ch->lfoFrequency;
	switch (ch->alg) {
	case 0:
		if (op2->envelopeGenerator->stage == stageOff) {
			__result.r0 = 0;
			__result.r1 = 0;
			return __result;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		result = fmOperatorPtr__next(op2, modIndex, op1out*ymfdata::ModulatorMultiplier);
		break;
	case 1:
		if (op1->envelopeGenerator->stage == stageOff && op2->envelopeGenerator->stage == stageOff) {
			__result.r0 = 0;
			__result.r1 = 0;
			return __result;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, noModulator);
		result = op1out + op2out;
		break;
	case 2:
		if (op1->envelopeGenerator->stage == stageOff && op2->envelopeGenerator->stage == stageOff && op3->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			__result.r0 = 0;
			__result.r1 = 0;
			return __result;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, noModulator);
		op3out = fmOperatorPtr__next(op3, modIndex, ch->feedbackOut3);
		op4out = fmOperatorPtr__next(op4, modIndex, noModulator);
		result = op1out + op2out + op3out + op4out;
		break;
	case 3:
		if (op4->envelopeGenerator->stage == stageOff) {
			__result.r0 = 0;
			__result.r1 = 0;
			return __result;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, noModulator);
		op3out = fmOperatorPtr__next(op3, modIndex, op2out*ymfdata::ModulatorMultiplier);
		result = fmOperatorPtr__next(op4, modIndex, (op1out + op3out)*ymfdata::ModulatorMultiplier);
		break;
	case 4:
		if (op4->envelopeGenerator->stage == stageOff) {
			__result.r0 = 0;
			__result.r1 = 0;
			return __result;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, op1out*ymfdata::ModulatorMultiplier);
		op3out = fmOperatorPtr__next(op3, modIndex, op2out*ymfdata::ModulatorMultiplier);
		result = fmOperatorPtr__next(op4, modIndex, op3out*ymfdata::ModulatorMultiplier);
		break;
	case 5:
		if (op2->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			__result.r0 = 0;
			__result.r1 = 0;
			return __result;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, op1out*ymfdata::ModulatorMultiplier);
		op3out = fmOperatorPtr__next(op3, modIndex, ch->feedbackOut3);
		op4out = fmOperatorPtr__next(op4, modIndex, op3out*ymfdata::ModulatorMultiplier);
		result = op2out + op4out;
		break;
	case 6:
		if (op1->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			__result.r0 = 0;
			__result.r1 = 0;
			return __result;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, noModulator);
		op3out = fmOperatorPtr__next(op3, modIndex, op2out*ymfdata::ModulatorMultiplier);
		op4out = fmOperatorPtr__next(op4, modIndex, op3out*ymfdata::ModulatorMultiplier);
		result = op1out + op4out;
		break;
	case 7:
		if (op1->envelopeGenerator->stage == stageOff && op3->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			__result.r0 = 0;
			__result.r1 = 0;
			return __result;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, noModulator);
		op3out = fmOperatorPtr__next(op3, modIndex, op2out*ymfdata::ModulatorMultiplier);
		op4out = fmOperatorPtr__next(op4, modIndex, noModulator);
		result = op1out + op3out + op4out;
		break;
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
	__result.r0 = result*ch->panCoefL;
	__result.r1 = result*ch->panCoefR;
	return __result;
}

void ChannelPtr__updateFrequency(std::shared_ptr<Channel> ch) {
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		fmOperatorPtr__setFrequency(op, ch->fnum, ch->block, ch->bo);
	}
}



// NewChip は、新しい Chip を作成します。
std::shared_ptr<Chip> NewChip(float64 sampleRate, float64 totalLevel, int dumpMIDIChannel) {
	auto chip = __ptr((const Chip){
		{} /*sync.Mutex*/,
		sampleRate,
		totalLevel,
		dumpMIDIChannel,
		make((std::shared_ptr<Channel> *)NULL, ymfdata::ChannelCount),
		make((float64 *)NULL, 2),
	});
	ChipPtr__initChannels(chip);
	return chip;
}

int debugDumpCount = 0;

// Next は、次のサンプルを生成し、その左右それぞれの振幅を返します。
ChipPtr__Next__result ChipPtr__Next(std::shared_ptr<Chip> chip) {
	ChipPtr__Next__result __result; // multi-result
	float64 l;
	float64 r;
	for (int _ = 0; _ < (int)chip->channels.size(); _++) {
		auto channel = chip->channels[_];
		sync::Mutex__Lock(chip->Mutex);
		auto __tuple = ChannelPtr__next(channel);
		auto cl = __tuple.r0;
		auto cr = __tuple.r1;
		sync::Mutex__Unlock(chip->Mutex);
		l = cl;
		r = cr;
	}
	auto v = math::Pow(10, chip->totalLevel/20);
	if (0 <= chip->dumpMIDIChannel) {
		debugDumpCount++;
		if (int(chip->sampleRate/ymfdata::DebugDumpFPS) <= debugDumpCount) {
			debugDumpCount = 0;
			auto toDump = make((std::shared_ptr<Channel> *)NULL, 0);
			for (int _ = 0; _ < (int)chip->channels.size(); _++) {
				auto ch = chip->channels[_];
				if (ch->midiChannelID == chip->dumpMIDIChannel && epsilon < ChannelPtr__currentLevel(ch)) {
					toDump = append(toDump, ch);
				}
			}
			if (0 < len(toDump)) {
				sort::Slice(toDump, [](int i, int j) -> bool  {
					return [41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4203f0fe0), Lbrack:11969, Index:(*ast.Ident)(0xc4203f1000), Rbrack:11971})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4203f0fe0), Lbrack:11969, Index:(*ast.Ident)(0xc4203f1000), Rbrack:11971})[0mChannelPtr__currentLevel(toDump[i]) < [41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4203f1060), Lbrack:11996, Index:(*ast.Ident)(0xc4203f1080), Rbrack:11998})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4203f1060), Lbrack:11996, Index:(*ast.Ident)(0xc4203f1080), Rbrack:11998})[0mChannelPtr__currentLevel(toDump[j]);
				});
				for (int _ = 0; _ < (int)toDump.size(); _++) {
					auto ch = toDump[_];
					fmt::Print(ChannelPtr__dump(ch));
				}
				fmt::Println(string("------------------------------"));
			}
		}
	}
	__result.r0 = l*v;
	__result.r1 = r*v;
	return __result;
}

void ChipPtr__initChannels(std::shared_ptr<Chip> chip) {
	chip->channels = make((std::shared_ptr<Channel> *)NULL, ymfdata::ChannelCount);
	for (int i = 0; i < (int)chip->channels.size(); i++) {
		chip->channels[i] = newChannel(i, chip);
	}
}




string stage__String(stage s) {
	switch (s) {
	case stageOff:
		return string("-");
		break;
	case stageAttack:
		return string("A");
		break;
	case stageDecay:
		return string("D");
		break;
	case stageSustain:
		return string("S");
		break;
	case stageRelease:
		return string("R");
		break;
	default:
		return string("?");
		break;
	}
}



std::shared_ptr<envelopeGenerator> newEnvelopeGenerator(float64 sampleRate) {
	auto eg = __ptr((const envelopeGenerator){
		sampleRate,
		{} /*gopkg.in/but80/fmfm.core.v1/sim.stage*/,
		{} /*bool*/,
		int(0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
		float64(.0),
	});
	envelopeGeneratorPtr__resetAll(eg);
	return eg;
}

void envelopeGeneratorPtr__reset(std::shared_ptr<envelopeGenerator> eg) {
	eg->currentLevel = .0;
	eg->stage = stageOff;
}

void envelopeGeneratorPtr__resetAll(std::shared_ptr<envelopeGenerator> eg) {
	eg->eam = false;
	eg->dam = 0;
	eg->sustainLevel = .0;
	envelopeGeneratorPtr__setTotalLevel(eg, 63);
	envelopeGeneratorPtr__setKeyScalingLevel(eg, 0, 0, 1, 0);
	envelopeGeneratorPtr__reset(eg);
}

void envelopeGeneratorPtr__setActualSustainLevel(std::shared_ptr<envelopeGenerator> eg, int sl) {
	if (sl == 0x0f) {
		eg->sustainLevel = 0;
	} else {
		auto slDB = -3.0*float64(sl);
		eg->sustainLevel = math::Pow(10.0, slDB/20.0);
	}
}

void envelopeGeneratorPtr__setTotalLevel(std::shared_ptr<envelopeGenerator> eg, int tl) {
	if (63 <= tl) {
		eg->tlCoef = .0;
		eg->kslTlCoef = .0;
		return;
	}
	auto tlDB = float64(tl)*-0.75;
	eg->tlCoef = math::Pow(10.0, tlDB/20.0);
	eg->kslTlCoef = eg->kslCoef*eg->tlCoef;
}

void envelopeGeneratorPtr__setKeyScalingLevel(std::shared_ptr<envelopeGenerator> eg, int fnum, int block, int bo, int ksl) {
	auto blkbo = block + 1 - bo;
	if (blkbo < 0) {
		blkbo = 0;
	} else {
		if (7 < blkbo) {
			blkbo = 7;
		}
	}
	eg->kslCoef = ymfdata::KSLTable[ksl][blkbo][fnum >> 5];
	eg->kslTlCoef = eg->kslCoef*eg->tlCoef;
}

void envelopeGeneratorPtr__setActualAR(std::shared_ptr<envelopeGenerator> eg, int attackRate, int ksr, int keyScaleNumber) {
	if (attackRate <= 0) {
		eg->arDiffPerSample = .0;
		return;
	}
	auto ksn = (keyScaleNumber >> 1) + (keyScaleNumber & 1);
	auto sec = attackTimeSecAt1[ksr][ksn]/float64(uint(1) << uint(attackRate - 1));
	eg->arDiffPerSample = 1.0/(sec*eg->sampleRate);
}

void envelopeGeneratorPtr__setActualDR(std::shared_ptr<envelopeGenerator> eg, int dr, int ksr, int keyScaleNumber) {
	if (dr == 0) {
		eg->drCoefPerSample = 1.0;
	} else {
		auto dbPerSecAt4 = decayDBPerSecAt4[ksr][keyScaleNumber]/2.0;
		auto dbPerSample = dbPerSecAt4*float64(uint(1) << uint(dr))/16.0/eg->sampleRate;
		eg->drCoefPerSample = math::Pow(10, -dbPerSample/10);
	}
}

void envelopeGeneratorPtr__setActualSR(std::shared_ptr<envelopeGenerator> eg, int sr, int ksr, int keyScaleNumber) {
	if (sr == 0) {
		eg->srCoefPerSample = 1.0;
	} else {
		auto dbPerSecAt4 = decayDBPerSecAt4[ksr][keyScaleNumber]/2.0;
		auto dbPerSample = dbPerSecAt4*float64(uint(1) << uint(sr))/16.0/eg->sampleRate;
		eg->srCoefPerSample = math::Pow(10, -dbPerSample/10);
	}
}

void envelopeGeneratorPtr__setActualRR(std::shared_ptr<envelopeGenerator> eg, int rr, int ksr, int keyScaleNumber) {
	if (rr == 0) {
		eg->rrCoefPerSample = 1.0;
	} else {
		auto dbPerSecAt4 = decayDBPerSecAt4[ksr][keyScaleNumber]/2.0;
		auto dbPerSample = dbPerSecAt4*float64(uint(1) << uint(rr))/16.0/eg->sampleRate;
		eg->rrCoefPerSample = math::Pow(10, -dbPerSample/10);
	}
}

float64 envelopeGeneratorPtr__getEnvelope(std::shared_ptr<envelopeGenerator> eg, int tremoloIndex) {
	switch (eg->stage) {
	case stageAttack:
		eg->currentLevel = eg->arDiffPerSample;
		if (eg->currentLevel < 1.0) {
			break;
		}
		eg->currentLevel = 1.0;
		eg->stage = stageDecay;
		// fallthrough;
	case stageDecay:
		if (eg->sustainLevel < eg->currentLevel) {
			eg->currentLevel = eg->drCoefPerSample;
			break;
		}
		eg->stage = stageSustain;
		// fallthrough;
	case stageSustain:
		if (epsilon < eg->currentLevel) {
			eg->currentLevel = eg->srCoefPerSample;
		} else {
			eg->stage = stageOff;
		}
		break;
		break;
	case stageRelease:
		if (epsilon < eg->currentLevel) {
			eg->currentLevel = eg->rrCoefPerSample;
		} else {
			eg->currentLevel = .0;
			eg->stage = stageOff;
		}
		break;
		break;
	}
	auto result = eg->currentLevel;
	if (eg->eam) {
		result = ymfdata::TremoloTable[eg->dam][tremoloIndex];
	}
	return result*eg->kslTlCoef;
}

void envelopeGeneratorPtr__keyOn(std::shared_ptr<envelopeGenerator> eg) {
	eg->stage = stageAttack;
}

void envelopeGeneratorPtr__keyOff(std::shared_ptr<envelopeGenerator> eg) {
	if (eg->stage != stageOff) {
		eg->stage = stageRelease;
	}
}

float64 decayDBPerSecAt4[16][2] = {
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

float64 attackTimeSecAt1[9][2] = {
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



std::shared_ptr<fmOperator> newOperator(int channelID, int operatorIndex, std::shared_ptr<Chip> chip) {
	return __ptr((const fmOperator){
		false,
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		int(0),
		float64(.0),
		int(0),
		int(0),
		int(0),
		1,
		newEnvelopeGenerator(chip->sampleRate),
		chip,
		channelID,
		operatorIndex,
		newPhaseGenerator(chip->sampleRate),
	});
}

void fmOperatorPtr__reset(std::shared_ptr<fmOperator> op) {
	phaseGeneratorPtr__reset(op->phaseGenerator);
	envelopeGeneratorPtr__reset(op->envelopeGenerator);
}

void fmOperatorPtr__resetAll(std::shared_ptr<fmOperator> op) {
	op->bo = 1;
	phaseGeneratorPtr__resetAll(op->phaseGenerator);
	envelopeGeneratorPtr__resetAll(op->envelopeGenerator);
}

string fmOperatorPtr__dump(std::shared_ptr<fmOperator> op) {
	auto eg = op->envelopeGenerator;
	auto pg = op->phaseGenerator;
	auto lvdb = math::Log10(eg->currentLevel)*20.0;
	auto lv = int((96.0 + lvdb)/8.0);
	if (lv < 0) {
		lv = 0;
	}
	auto lvstr = strings::Repeat(string("|"), lv);
	auto cm = string("C");
	if (op->isModulator) {
		cm = string("M");
	}
	auto am = string("-");
	if (eg->eam) {
		am = fmt::Sprintf(string("%d"), eg->dam);
	}
	auto vb = string("-");
	if (pg->evb) {
		vb = fmt::Sprintf(string("%d"), pg->dvb);
	}
	auto phase = pg->phaseFrac64 >> ymfdata::WaveformIndexShift;
	auto phstr = [41mobjectOf[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:18927, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc420410420)})[0m[41mCallExpr[0m[31m<<nil>>(<nil>)[0mbyte (string("        "));
	phstr[phase >> (ymfdata::WaveformLenBits - 3)] = string.byte("|");
	return fmt::Sprintf(string("%d: %s mul=%02d ws=%02d adssr=%02d,%02d,%02d,%02d,%02d tl=%f am=%s vb=%s dt=%d ksr=%d fb=%3.2f ksn=%02d ksl=%f st=%s ph=%s lv=%-03d %s"), op->operatorIndex, cm.c_str(), op->mult, op->ws, op->ar, op->dr, op->sl, op->sr, op->rr, eg->tlCoef, am.c_str(), vb.c_str(), op->dt, op->ksr, op->feedbackCoef, op->keyScaleNumber, eg->kslCoef, stage__String(eg->stage), string(phstr), int(math::Floor(lvdb)), lvstr.c_str());
}

void fmOperatorPtr__setEAM(std::shared_ptr<fmOperator> op, int v) {
	op->envelopeGenerator->eam = v != 0;
}

void fmOperatorPtr__setEVB(std::shared_ptr<fmOperator> op, int v) {
	op->phaseGenerator->evb = v != 0;
}

void fmOperatorPtr__setDAM(std::shared_ptr<fmOperator> op, int v) {
	op->envelopeGenerator->dam = v;
}

void fmOperatorPtr__setDVB(std::shared_ptr<fmOperator> op, int v) {
	op->phaseGenerator->dvb = v;
}

void fmOperatorPtr__setDT(std::shared_ptr<fmOperator> op, int v) {
	op->dt = v;
	fmOperatorPtr__updateFrequency(op);
}

void fmOperatorPtr__setKSR(std::shared_ptr<fmOperator> op, int v) {
	op->ksr = v;
	fmOperatorPtr__updateEnvelope(op);
}

void fmOperatorPtr__setMULT(std::shared_ptr<fmOperator> op, int v) {
	op->mult = v;
	fmOperatorPtr__updateFrequency(op);
}

void fmOperatorPtr__setKSL(std::shared_ptr<fmOperator> op, int v) {
	op->ksl = v;
	envelopeGeneratorPtr__setKeyScalingLevel(op->envelopeGenerator, op->fnum, op->block, op->bo, op->ksl);
}

void fmOperatorPtr__setTL(std::shared_ptr<fmOperator> op, int v) {
	envelopeGeneratorPtr__setTotalLevel(op->envelopeGenerator, v);
}

void fmOperatorPtr__setAR(std::shared_ptr<fmOperator> op, int v) {
	op->ar = v;
	envelopeGeneratorPtr__setActualAR(op->envelopeGenerator, op->ar, op->ksr, op->keyScaleNumber);
}

void fmOperatorPtr__setDR(std::shared_ptr<fmOperator> op, int v) {
	op->dr = v;
	envelopeGeneratorPtr__setActualDR(op->envelopeGenerator, op->dr, op->ksr, op->keyScaleNumber);
}

void fmOperatorPtr__setSL(std::shared_ptr<fmOperator> op, int v) {
	op->sl = v;
	envelopeGeneratorPtr__setActualSustainLevel(op->envelopeGenerator, op->sl);
}

void fmOperatorPtr__setSR(std::shared_ptr<fmOperator> op, int v) {
	op->sr = v;
	envelopeGeneratorPtr__setActualSR(op->envelopeGenerator, op->sr, op->ksr, op->keyScaleNumber);
}

void fmOperatorPtr__setRR(std::shared_ptr<fmOperator> op, int v) {
	op->rr = v;
	envelopeGeneratorPtr__setActualRR(op->envelopeGenerator, op->rr, op->ksr, op->keyScaleNumber);
}

void fmOperatorPtr__setXOF(std::shared_ptr<fmOperator> op, int v) {
	op->xof = v;
}

void fmOperatorPtr__setWS(std::shared_ptr<fmOperator> op, int v) {
	op->ws = v;
}

void fmOperatorPtr__setFB(std::shared_ptr<fmOperator> op, int v) {
	op->feedbackCoef = ymfdata::FeedbackTable[v];
}

float64 fmOperatorPtr__next(std::shared_ptr<fmOperator> op, int modIndex, float64 modulator) {
	auto phaseFrac64 = phaseGeneratorPtr__getPhase(op->phaseGenerator, modIndex);
	if (op->envelopeGenerator->stage == stageOff) {
		return 0;
	}
	auto envelope = envelopeGeneratorPtr__getEnvelope(op->envelopeGenerator, modIndex);
	auto sampleIndex = uint64(phaseFrac64) >> ymfdata::WaveformIndexShift;
	sampleIndex = uint64((modulator + ymfdata::WaveformLen)*ymfdata::WaveformLen);
	return ymfdata::Waveforms[op->ws][sampleIndex & 1023]*envelope;
}

void fmOperatorPtr__keyOn(std::shared_ptr<fmOperator> op) {
	if (0 < op->ar) {
		envelopeGeneratorPtr__keyOn(op->envelopeGenerator);
	} else {
		op->envelopeGenerator->stage = stageOff;
	}
}

void fmOperatorPtr__keyOff(std::shared_ptr<fmOperator> op) {
	if (op->xof == 0) {
		envelopeGeneratorPtr__keyOff(op->envelopeGenerator);
	}
}

void fmOperatorPtr__setFrequency(std::shared_ptr<fmOperator> op, int fnum, int blk, int bo) {
	op->keyScaleNumber = (blk + 1 - bo)*2 + (fnum >> 9);
	if (op->keyScaleNumber < 0) {
		op->keyScaleNumber = 0;
	} else {
		if (15 < op->keyScaleNumber) {
			op->keyScaleNumber = 15;
		}
	}
	op->fnum = fnum;
	op->block = blk;
	op->bo = bo;
	fmOperatorPtr__updateFrequency(op);
	fmOperatorPtr__updateEnvelope(op);
	envelopeGeneratorPtr__setKeyScalingLevel(op->envelopeGenerator, op->fnum, op->block, op->bo, op->ksl);
}

void fmOperatorPtr__updateFrequency(std::shared_ptr<fmOperator> op) {
	phaseGeneratorPtr__setFrequency(op->phaseGenerator, op->fnum, op->block, op->bo, op->mult, op->dt);
}

void fmOperatorPtr__updateEnvelope(std::shared_ptr<fmOperator> op) {
	envelopeGeneratorPtr__setActualAR(op->envelopeGenerator, op->ar, op->ksr, op->keyScaleNumber);
	envelopeGeneratorPtr__setActualDR(op->envelopeGenerator, op->dr, op->ksr, op->keyScaleNumber);
	envelopeGeneratorPtr__setActualSR(op->envelopeGenerator, op->sr, op->ksr, op->keyScaleNumber);
	envelopeGeneratorPtr__setActualRR(op->envelopeGenerator, op->rr, op->ksr, op->keyScaleNumber);
}



std::shared_ptr<phaseGenerator> newPhaseGenerator(float64 sampleRate) {
	auto pg = __ptr((const phaseGenerator){
		sampleRate,
		{} /*bool*/,
		int(0),
		{} /*gopkg.in/but80/fmfm.core.v1/ymf/ymfdata.Frac64*/,
		{} /*gopkg.in/but80/fmfm.core.v1/ymf/ymfdata.Frac64*/,
	});
	phaseGeneratorPtr__reset(pg);
	return pg;
}

void phaseGeneratorPtr__reset(std::shared_ptr<phaseGenerator> pg) {
	pg->phaseFrac64 = 0;
}

void phaseGeneratorPtr__resetAll(std::shared_ptr<phaseGenerator> pg) {
	pg->evb = false;
	pg->dvb = 0;
	pg->phaseIncrementFrac64 = 0;
	phaseGeneratorPtr__reset(pg);
}

void phaseGeneratorPtr__setFrequency(std::shared_ptr<phaseGenerator> pg, int fnum, int block, int bo, int mult, int dt) {
	auto baseFrequency = float64(fnum << uint(block + 3 - bo))/(16.0*ymfdata::FNUMCoef);
	auto ksn = block << 1 | fnum >> 9;
	auto operatorFrequency = baseFrequency + ymfdata::DTCoef[dt][ksn];
	pg->phaseIncrementFrac64 = ymfdata::FloatToFrac64(operatorFrequency/pg->sampleRate);
	pg->phaseIncrementFrac64 = ymfdata::Frac64__MulUint64(pg->phaseIncrementFrac64, ymfdata::MultTable2[mult]);
	pg->phaseIncrementFrac64 = 1;
}

ymfdata::Frac64 phaseGeneratorPtr__getPhase(std::shared_ptr<phaseGenerator> pg, int vibratoIndex) {
	if (pg->evb) {
		pg->phaseFrac64 = ymfdata::Frac64__MulInt32Frac32(pg->phaseIncrementFrac64, ymfdata::VibratoTableInt32Frac32[pg->dvb][vibratoIndex]);
	} else {
		pg->phaseFrac64 = pg->phaseIncrementFrac64;
	}
	return pg->phaseFrac64;
}




// NewRegisters は、新しい Registers を作成します。
std::shared_ptr<Registers> NewRegisters(std::shared_ptr<Chip> chip) {
	return __ptr((const Registers){
		chip,
	});
}

// WriteOperator は、オペレータレジスタに値を書き込みます。
void RegistersPtr__WriteOperator(std::shared_ptr<Registers> regs, int channel, int operatorIndex, ymf::OpRegister offset, int v) {
	sync::Mutex__Lock(regs->chip->Mutex);
	defer([]{
		sync::Mutex__Unlock(regs->chip->Mutex);
	});
	switch (offset) {
	case ymf::EAM:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204276e0), Lbrack:24780, Index:(*ast.Ident)(0xc420427700), Rbrack:24794})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204276e0), Lbrack:24780, Index:(*ast.Ident)(0xc420427700), Rbrack:24794})[0mfmOperatorPtr__setEAM(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::EVB:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204278c0), Lbrack:24860, Index:(*ast.Ident)(0xc4204278e0), Rbrack:24874})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204278c0), Lbrack:24860, Index:(*ast.Ident)(0xc4204278e0), Rbrack:24874})[0mfmOperatorPtr__setEVB(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::DAM:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420427ac0), Lbrack:24940, Index:(*ast.Ident)(0xc420427ae0), Rbrack:24954})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420427ac0), Lbrack:24940, Index:(*ast.Ident)(0xc420427ae0), Rbrack:24954})[0mfmOperatorPtr__setDAM(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::DVB:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420427ca0), Lbrack:25020, Index:(*ast.Ident)(0xc420427cc0), Rbrack:25034})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420427ca0), Lbrack:25020, Index:(*ast.Ident)(0xc420427cc0), Rbrack:25034})[0mfmOperatorPtr__setDVB(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::DT:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420427e80), Lbrack:25099, Index:(*ast.Ident)(0xc420427ea0), Rbrack:25113})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420427e80), Lbrack:25099, Index:(*ast.Ident)(0xc420427ea0), Rbrack:25113})[0mfmOperatorPtr__setDT(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::KSR:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430060), Lbrack:25178, Index:(*ast.Ident)(0xc420430080), Rbrack:25192})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430060), Lbrack:25178, Index:(*ast.Ident)(0xc420430080), Rbrack:25192})[0mfmOperatorPtr__setKSR(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::MULT:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430240), Lbrack:25259, Index:(*ast.Ident)(0xc420430260), Rbrack:25273})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430240), Lbrack:25259, Index:(*ast.Ident)(0xc420430260), Rbrack:25273})[0mfmOperatorPtr__setMULT(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::KSL:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430420), Lbrack:25340, Index:(*ast.Ident)(0xc420430440), Rbrack:25354})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430420), Lbrack:25340, Index:(*ast.Ident)(0xc420430440), Rbrack:25354})[0mfmOperatorPtr__setKSL(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::TL:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430600), Lbrack:25419, Index:(*ast.Ident)(0xc420430620), Rbrack:25433})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430600), Lbrack:25419, Index:(*ast.Ident)(0xc420430620), Rbrack:25433})[0mfmOperatorPtr__setTL(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::AR:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204307e0), Lbrack:25497, Index:(*ast.Ident)(0xc420430800), Rbrack:25511})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204307e0), Lbrack:25497, Index:(*ast.Ident)(0xc420430800), Rbrack:25511})[0mfmOperatorPtr__setAR(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::DR:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204309c0), Lbrack:25575, Index:(*ast.Ident)(0xc4204309e0), Rbrack:25589})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204309c0), Lbrack:25575, Index:(*ast.Ident)(0xc4204309e0), Rbrack:25589})[0mfmOperatorPtr__setDR(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::SL:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430ba0), Lbrack:25653, Index:(*ast.Ident)(0xc420430bc0), Rbrack:25667})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430ba0), Lbrack:25653, Index:(*ast.Ident)(0xc420430bc0), Rbrack:25667})[0mfmOperatorPtr__setSL(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::SR:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430d80), Lbrack:25731, Index:(*ast.Ident)(0xc420430da0), Rbrack:25745})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430d80), Lbrack:25731, Index:(*ast.Ident)(0xc420430da0), Rbrack:25745})[0mfmOperatorPtr__setSR(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::RR:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430f60), Lbrack:25809, Index:(*ast.Ident)(0xc420430f80), Rbrack:25823})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420430f60), Lbrack:25809, Index:(*ast.Ident)(0xc420430f80), Rbrack:25823})[0mfmOperatorPtr__setRR(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::XOF:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420431140), Lbrack:25888, Index:(*ast.Ident)(0xc420431160), Rbrack:25902})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420431140), Lbrack:25888, Index:(*ast.Ident)(0xc420431160), Rbrack:25902})[0mfmOperatorPtr__setXOF(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::WS:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420431320), Lbrack:25967, Index:(*ast.Ident)(0xc420431340), Rbrack:25981})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420431320), Lbrack:25967, Index:(*ast.Ident)(0xc420431340), Rbrack:25981})[0mfmOperatorPtr__setWS(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	case ymf::FB:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420431500), Lbrack:26045, Index:(*ast.Ident)(0xc420431520), Rbrack:26059})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420431500), Lbrack:26045, Index:(*ast.Ident)(0xc420431520), Rbrack:26059})[0mfmOperatorPtr__setFB(regs->chip->channels[channel]->operators[operatorIndex], v);
		break;
	}
}

// WriteTL は、TLレジスタに値を書き込みます。
void RegistersPtr__WriteTL(std::shared_ptr<Registers> regs, int channel, int operatorIndex, int tlCarrier, int tlModulator) {
	sync::Mutex__Lock(regs->chip->Mutex);
	defer([]{
		sync::Mutex__Unlock(regs->chip->Mutex);
	});
	if (regs->chip->channels[channel]->operators[operatorIndex]->isModulator) {
		RegistersPtr__WriteOperator(regs, channel, operatorIndex, ymf::TL, tlModulator);
	} else {
		RegistersPtr__WriteOperator(regs, channel, operatorIndex, ymf::TL, tlCarrier);
	}
}

// DebugSetMIDIChannel は、チャンネルを使用しているMIDIチャンネル番号をデバッグ用にセットします。
void RegistersPtr__DebugSetMIDIChannel(std::shared_ptr<Registers> regs, int channel, int midiChannel) {
	sync::Mutex__Lock(regs->chip->Mutex);
	defer([]{
		sync::Mutex__Unlock(regs->chip->Mutex);
	});
	regs->chip->channels[channel]->midiChannelID = midiChannel;
}

// WriteChannel は、チャンネルレジスタに値を書き込みます。
void RegistersPtr__WriteChannel(std::shared_ptr<Registers> regs, int channel, ymf::ChRegister offset, int v) {
	sync::Mutex__Lock(regs->chip->Mutex);
	defer([]{
		sync::Mutex__Unlock(regs->chip->Mutex);
	});
	switch (offset) {
	case ymf::KON:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434660), Lbrack:27089, Index:(*ast.Ident)(0xc420434680), Rbrack:27097})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434660), Lbrack:27089, Index:(*ast.Ident)(0xc420434680), Rbrack:27097})[0mChannelPtr__setKON(regs->chip->channels[channel], v);
		break;
	case ymf::BLOCK:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204347e0), Lbrack:27146, Index:(*ast.Ident)(0xc420434800), Rbrack:27154})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204347e0), Lbrack:27146, Index:(*ast.Ident)(0xc420434800), Rbrack:27154})[0mChannelPtr__setBLOCK(regs->chip->channels[channel], v);
		break;
	case ymf::FNUM:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434980), Lbrack:27204, Index:(*ast.Ident)(0xc4204349a0), Rbrack:27212})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434980), Lbrack:27204, Index:(*ast.Ident)(0xc4204349a0), Rbrack:27212})[0mChannelPtr__setFNUM(regs->chip->channels[channel], v);
		break;
	case ymf::ALG:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434b00), Lbrack:27260, Index:(*ast.Ident)(0xc420434b20), Rbrack:27268})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434b00), Lbrack:27260, Index:(*ast.Ident)(0xc420434b20), Rbrack:27268})[0mChannelPtr__setALG(regs->chip->channels[channel], v);
		break;
	case ymf::LFO:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434c80), Lbrack:27315, Index:(*ast.Ident)(0xc420434ca0), Rbrack:27323})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434c80), Lbrack:27315, Index:(*ast.Ident)(0xc420434ca0), Rbrack:27323})[0mChannelPtr__setLFO(regs->chip->channels[channel], v);
		break;
	case ymf::PANPOT:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434e00), Lbrack:27373, Index:(*ast.Ident)(0xc420434e20), Rbrack:27381})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434e00), Lbrack:27373, Index:(*ast.Ident)(0xc420434e20), Rbrack:27381})[0mChannelPtr__setPANPOT(regs->chip->channels[channel], v);
		break;
	case ymf::CHPAN:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434f80), Lbrack:27433, Index:(*ast.Ident)(0xc420434fa0), Rbrack:27441})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420434f80), Lbrack:27433, Index:(*ast.Ident)(0xc420434fa0), Rbrack:27441})[0mChannelPtr__setCHPAN(regs->chip->channels[channel], v);
		break;
	case ymf::VOLUME:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420435100), Lbrack:27493, Index:(*ast.Ident)(0xc420435120), Rbrack:27501})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420435100), Lbrack:27493, Index:(*ast.Ident)(0xc420435120), Rbrack:27501})[0mChannelPtr__setVOLUME(regs->chip->channels[channel], v);
		break;
	case ymf::EXPRESSION:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420435280), Lbrack:27558, Index:(*ast.Ident)(0xc4204352a0), Rbrack:27566})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420435280), Lbrack:27558, Index:(*ast.Ident)(0xc4204352a0), Rbrack:27566})[0mChannelPtr__setEXPRESSION(regs->chip->channels[channel], v);
		break;
	case ymf::VELOCITY:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420435400), Lbrack:27625, Index:(*ast.Ident)(0xc420435420), Rbrack:27633})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420435400), Lbrack:27625, Index:(*ast.Ident)(0xc420435420), Rbrack:27633})[0mChannelPtr__setVELOCITY(regs->chip->channels[channel], v);
		break;
	case ymf::BO:
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420435580), Lbrack:27684, Index:(*ast.Ident)(0xc4204355a0), Rbrack:27692})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420435580), Lbrack:27684, Index:(*ast.Ident)(0xc4204355a0), Rbrack:27692})[0mChannelPtr__setBO(regs->chip->channels[channel], v);
		break;
	case ymf::RESET:
		if (v != 0) {
			[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420435740), Lbrack:27755, Index:(*ast.Ident)(0xc420435760), Rbrack:27763})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420435740), Lbrack:27755, Index:(*ast.Ident)(0xc420435760), Rbrack:27763})[0mChannelPtr__resetAll(regs->chip->channels[channel]);
		}
		break;
	}
}

}

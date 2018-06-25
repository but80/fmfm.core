#include "sim.h"

namespace sim {


#define noModulator (0)


Channel *newChannel(int channelID, Chip *chip) {
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
		fmOperatorPtr__resetAll(op);
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
		result = "  " + fmOperatorPtr__dump(op) + "\n";
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
		fmOperatorPtr__keyOn(op);
	}
	ch->kon = 1;
}

void ChannelPtr__keyOff(Channel *ch) {
	if (ch->kon == 0) {
		return;
	}
	for (int _ = 0; _ < sizeof(ch->operators) / sizeof(ch->operators[0]); _++) {
		auto op = ch->operators[_];
		fmOperatorPtr__keyOff(op);
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

class nextResult {
	float64 r0;
	float64 r1;
};
nextResult ChannelPtr__next(Channel *ch) {
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
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		result = fmOperatorPtr__next(op2, modIndex, op1out*ymfdata->ModulatorMultiplier);
	} else if (__tag == 1) {
		if (op1->envelopeGenerator->stage == stageOff && op2->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, noModulator);
		result = op1out + op2out;
	} else if (__tag == 2) {
		if (op1->envelopeGenerator->stage == stageOff && op2->envelopeGenerator->stage == stageOff && op3->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, noModulator);
		op3out = fmOperatorPtr__next(op3, modIndex, ch->feedbackOut3);
		op4out = fmOperatorPtr__next(op4, modIndex, noModulator);
		result = op1out + op2out + op3out + op4out;
	} else if (__tag == 3) {
		if (op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, noModulator);
		op3out = fmOperatorPtr__next(op3, modIndex, op2out*ymfdata->ModulatorMultiplier);
		result = fmOperatorPtr__next(op4, modIndex, (op1out + op3out)*ymfdata->ModulatorMultiplier);
	} else if (__tag == 4) {
		if (op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, op1out*ymfdata->ModulatorMultiplier);
		op3out = fmOperatorPtr__next(op3, modIndex, op2out*ymfdata->ModulatorMultiplier);
		result = fmOperatorPtr__next(op4, modIndex, op3out*ymfdata->ModulatorMultiplier);
	} else if (__tag == 5) {
		if (op2->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, op1out*ymfdata->ModulatorMultiplier);
		op3out = fmOperatorPtr__next(op3, modIndex, ch->feedbackOut3);
		op4out = fmOperatorPtr__next(op4, modIndex, op3out*ymfdata->ModulatorMultiplier);
		result = op2out + op4out;
	} else if (__tag == 6) {
		if (op1->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, noModulator);
		op3out = fmOperatorPtr__next(op3, modIndex, op2out*ymfdata->ModulatorMultiplier);
		op4out = fmOperatorPtr__next(op4, modIndex, op3out*ymfdata->ModulatorMultiplier);
		result = op1out + op4out;
	} else if (__tag == 7) {
		if (op1->envelopeGenerator->stage == stageOff && op3->envelopeGenerator->stage == stageOff && op4->envelopeGenerator->stage == stageOff) {
			return 0, 0;
		}
		op1out = fmOperatorPtr__next(op1, modIndex, ch->feedbackOut1);
		op2out = fmOperatorPtr__next(op2, modIndex, noModulator);
		op3out = fmOperatorPtr__next(op3, modIndex, op2out*ymfdata->ModulatorMultiplier);
		op4out = fmOperatorPtr__next(op4, modIndex, noModulator);
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
		fmOperatorPtr__setFrequency(op, ch->fnum, ch->block, ch->bo);
	}
}



// NewChip は、新しい Chip を作成します。
Chip *NewChip(float64 sampleRate, float64 totalLevel, int dumpMIDIChannel) {
	auto chip = __ptr((const Chip){
		{} /*sync.Mutex*/,
		sampleRate,
		totalLevel,
		dumpMIDIChannel,
		make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:11072, Len:ast.Expr(nil), Elt:(*ast.StarExpr)(0xc4203ddd40)})[0m, ymfdata->ChannelCount),
		make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:11131, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc4203dde20)})[0m, 2),
	});
	ChipPtr__initChannels(chip);
	return chip;
}

auto debugDumpCount = 0;

// Next は、次のサンプルを生成し、その左右それぞれの振幅を返します。
class NextResult {
	float64 r0;
	float64 r1;
};
NextResult ChipPtr__Next(Chip *chip) {
	float64 l, float64 r;
	for (int _ = 0; _ < (int)chip->channels.size(); _++) {
		auto channel = chip->channels[_];
		sync::Mutex__Lock(chip->Mutex);
		auto cl, cr = ChannelPtr__next(channel);
		sync::Mutex__Unlock(chip->Mutex);
		l = cl;
		r = cr;
	}
	auto v = math::Pow(10, chip->totalLevel/20);
	if (0 <= chip->dumpMIDIChannel) {
		debugDumpCount++;
		if (int(chip->sampleRate/ymfdata->DebugDumpFPS) <= debugDumpCount) {
			debugDumpCount = 0;
			auto toDump = {};
			for (int _ = 0; _ < (int)chip->channels.size(); _++) {
				auto ch = chip->channels[_];
				if (ch->midiChannelID == chip->dumpMIDIChannel && epsilon < ChannelPtr__currentLevel(ch)) {
					toDump = append(toDump, ch);
				}
			}
			if (0 < len(toDump)) {
				sort::Slice(toDump, 				bool UNKNOWN(int i, int j) {
					return [41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4203e4d00), Lbrack:11962, Index:(*ast.Ident)(0xc4203e4d20), Rbrack:11964})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4203e4d00), Lbrack:11962, Index:(*ast.Ident)(0xc4203e4d20), Rbrack:11964})[0mChannelPtr__currentLevel(toDump[i]) < [41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4203e4d80), Lbrack:11989, Index:(*ast.Ident)(0xc4203e4da0), Rbrack:11991})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4203e4d80), Lbrack:11989, Index:(*ast.Ident)(0xc4203e4da0), Rbrack:11991})[0mChannelPtr__currentLevel(toDump[j]);
				});
				for (int _ = 0; _ < (int)toDump.size(); _++) {
					auto ch = toDump[_];
					fmt::Print(ChannelPtr__dump(ch));
				}
				fmt::Println("------------------------------");
			}
		}
	}
	return l*v, r*v;
}

void ChipPtr__initChannels(Chip *chip) {
	chip->channels = make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:12223, Len:ast.Expr(nil), Elt:(*ast.StarExpr)(0xc4203e5260)})[0m, ymfdata->ChannelCount);
	for (int i = 0; i < (int)chip->channels.size(); i++) {
		chip->channels[i] = newChannel(i, chip);
	}
}



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



fmOperator *newOperator(int channelID, int operatorIndex, Chip *chip) {
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

void fmOperatorPtr__reset(fmOperator *op) {
	phaseGeneratorPtr__reset(op->phaseGenerator);
	envelopeGeneratorPtr__reset(op->envelopeGenerator);
}

void fmOperatorPtr__resetAll(fmOperator *op) {
	op->bo = 1;
	phaseGeneratorPtr__resetAll(op->phaseGenerator);
	envelopeGeneratorPtr__resetAll(op->envelopeGenerator);
}

string fmOperatorPtr__dump(fmOperator *op) {
	auto eg = op->envelopeGenerator;
	auto pg = op->phaseGenerator;
	auto lvdb = math::Log10(eg->currentLevel)*20.0;
	auto lv = int((96.0 + lvdb)/8.0);
	if (lv < 0) {
		lv = 0;
	}
	auto lvstr = strings::Repeat("|", lv);
	auto cm = "C";
	if (op->isModulator) {
		cm = "M";
	}
	auto am = "-";
	if (eg->eam) {
		am = fmt::Sprintf("%d", eg->dam);
	}
	auto vb = "-";
	if (pg->evb) {
		vb = fmt::Sprintf("%d", pg->dvb);
	}
	auto phase = pg->phaseFrac64 >> ymfdata->WaveformIndexShift;
	auto phstr = [41mobjectOf[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:18920, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc420406140)})[0m[41mCallExpr[0m[31m<<nil>>(<nil>)[0m[41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:18920, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc420406140)})[0m("        ");
	phstr[phase >> (ymfdata->WaveformLenBits - 3)] = string.byte("|");
	return fmt::Sprintf("%d: %s mul=%02d ws=%02d adssr=%02d,%02d,%02d,%02d,%02d tl=%f am=%s vb=%s dt=%d ksr=%d fb=%3.2f ksn=%02d ksl=%f st=%s ph=%s lv=%-03d %s", op->operatorIndex, cm, op->mult, op->ws, op->ar, op->dr, op->sl, op->sr, op->rr, eg->tlCoef, am, vb, op->dt, op->ksr, op->feedbackCoef, op->keyScaleNumber, eg->kslCoef, stage__String(eg->stage), string(phstr), int(math::Floor(lvdb)), lvstr);
}

void fmOperatorPtr__setEAM(fmOperator *op, int v) {
	op->envelopeGenerator->eam = v != 0;
}

void fmOperatorPtr__setEVB(fmOperator *op, int v) {
	op->phaseGenerator->evb = v != 0;
}

void fmOperatorPtr__setDAM(fmOperator *op, int v) {
	op->envelopeGenerator->dam = v;
}

void fmOperatorPtr__setDVB(fmOperator *op, int v) {
	op->phaseGenerator->dvb = v;
}

void fmOperatorPtr__setDT(fmOperator *op, int v) {
	op->dt = v;
	fmOperatorPtr__updateFrequency(op);
}

void fmOperatorPtr__setKSR(fmOperator *op, int v) {
	op->ksr = v;
	fmOperatorPtr__updateEnvelope(op);
}

void fmOperatorPtr__setMULT(fmOperator *op, int v) {
	op->mult = v;
	fmOperatorPtr__updateFrequency(op);
}

void fmOperatorPtr__setKSL(fmOperator *op, int v) {
	op->ksl = v;
	envelopeGeneratorPtr__setKeyScalingLevel(op->envelopeGenerator, op->fnum, op->block, op->bo, op->ksl);
}

void fmOperatorPtr__setTL(fmOperator *op, int v) {
	envelopeGeneratorPtr__setTotalLevel(op->envelopeGenerator, v);
}

void fmOperatorPtr__setAR(fmOperator *op, int v) {
	op->ar = v;
	envelopeGeneratorPtr__setActualAR(op->envelopeGenerator, op->ar, op->ksr, op->keyScaleNumber);
}

void fmOperatorPtr__setDR(fmOperator *op, int v) {
	op->dr = v;
	envelopeGeneratorPtr__setActualDR(op->envelopeGenerator, op->dr, op->ksr, op->keyScaleNumber);
}

void fmOperatorPtr__setSL(fmOperator *op, int v) {
	op->sl = v;
	envelopeGeneratorPtr__setActualSustainLevel(op->envelopeGenerator, op->sl);
}

void fmOperatorPtr__setSR(fmOperator *op, int v) {
	op->sr = v;
	envelopeGeneratorPtr__setActualSR(op->envelopeGenerator, op->sr, op->ksr, op->keyScaleNumber);
}

void fmOperatorPtr__setRR(fmOperator *op, int v) {
	op->rr = v;
	envelopeGeneratorPtr__setActualRR(op->envelopeGenerator, op->rr, op->ksr, op->keyScaleNumber);
}

void fmOperatorPtr__setXOF(fmOperator *op, int v) {
	op->xof = v;
}

void fmOperatorPtr__setWS(fmOperator *op, int v) {
	op->ws = v;
}

void fmOperatorPtr__setFB(fmOperator *op, int v) {
	op->feedbackCoef = ymfdata->FeedbackTable[v];
}

float64 fmOperatorPtr__next(fmOperator *op, int modIndex, float64 modulator) {
	auto phaseFrac64 = phaseGeneratorPtr__getPhase(op->phaseGenerator, modIndex);
	if (op->envelopeGenerator->stage == stageOff) {
		return 0;
	}
	auto envelope = envelopeGeneratorPtr__getEnvelope(op->envelopeGenerator, modIndex);
	auto sampleIndex = uint64(phaseFrac64) >> ymfdata->WaveformIndexShift;
	sampleIndex = uint64((modulator + ymfdata->WaveformLen)*ymfdata->WaveformLen);
	return ymfdata->Waveforms[op->ws][sampleIndex & 1023]*envelope;
}

void fmOperatorPtr__keyOn(fmOperator *op) {
	if (0 < op->ar) {
		envelopeGeneratorPtr__keyOn(op->envelopeGenerator);
	} else {
		op->envelopeGenerator->stage = stageOff;
	}
}

void fmOperatorPtr__keyOff(fmOperator *op) {
	if (op->xof == 0) {
		envelopeGeneratorPtr__keyOff(op->envelopeGenerator);
	}
}

void fmOperatorPtr__setFrequency(fmOperator *op, int fnum, int blk, int bo) {
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

void fmOperatorPtr__updateFrequency(fmOperator *op) {
	phaseGeneratorPtr__setFrequency(op->phaseGenerator, op->fnum, op->block, op->bo, op->mult, op->dt);
}

void fmOperatorPtr__updateEnvelope(fmOperator *op) {
	envelopeGeneratorPtr__setActualAR(op->envelopeGenerator, op->ar, op->ksr, op->keyScaleNumber);
	envelopeGeneratorPtr__setActualDR(op->envelopeGenerator, op->dr, op->ksr, op->keyScaleNumber);
	envelopeGeneratorPtr__setActualSR(op->envelopeGenerator, op->sr, op->ksr, op->keyScaleNumber);
	envelopeGeneratorPtr__setActualRR(op->envelopeGenerator, op->rr, op->ksr, op->keyScaleNumber);
}



phaseGenerator *newPhaseGenerator(float64 sampleRate) {
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
	pg->phaseIncrementFrac64 = ymfdata::Frac64__MulUint64(pg->phaseIncrementFrac64, ymfdata->MultTable2[mult]);
	pg->phaseIncrementFrac64 = 1;
}

ymfdata::Frac64 phaseGeneratorPtr__getPhase(phaseGenerator *pg, int vibratoIndex) {
	if (pg->evb) {
		pg->phaseFrac64 = ymfdata::Frac64__MulInt32Frac32(pg->phaseIncrementFrac64, ymfdata->VibratoTableInt32Frac32[pg->dvb][vibratoIndex]);
	} else {
		pg->phaseFrac64 = pg->phaseIncrementFrac64;
	}
	return pg->phaseFrac64;
}



ymf::Registers _ = __ptr((const Registers){
});

// NewRegisters は、新しい Registers を作成します。
Registers *NewRegisters(Chip *chip) {
	return __ptr((const Registers){
		chip,
	});
}

// WriteOperator は、オペレータレジスタに値を書き込みます。
void RegistersPtr__WriteOperator(Registers *regs, int channel, int operatorIndex, ymf::OpRegister offset, int v) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:24671, Call:(*ast.CallExpr)(0xc420416ac0)})[0m
	auto __tag = offset;
	if (__tag == ymf->EAM) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420419400), Lbrack:24773, Index:(*ast.Ident)(0xc420419420), Rbrack:24787})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420419400), Lbrack:24773, Index:(*ast.Ident)(0xc420419420), Rbrack:24787})[0mfmOperatorPtr__setEAM(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->EVB) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204195e0), Lbrack:24853, Index:(*ast.Ident)(0xc420419600), Rbrack:24867})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204195e0), Lbrack:24853, Index:(*ast.Ident)(0xc420419600), Rbrack:24867})[0mfmOperatorPtr__setEVB(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DAM) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204197e0), Lbrack:24933, Index:(*ast.Ident)(0xc420419800), Rbrack:24947})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204197e0), Lbrack:24933, Index:(*ast.Ident)(0xc420419800), Rbrack:24947})[0mfmOperatorPtr__setDAM(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DVB) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204199c0), Lbrack:25013, Index:(*ast.Ident)(0xc4204199e0), Rbrack:25027})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204199c0), Lbrack:25013, Index:(*ast.Ident)(0xc4204199e0), Rbrack:25027})[0mfmOperatorPtr__setDVB(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DT) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420419ba0), Lbrack:25092, Index:(*ast.Ident)(0xc420419bc0), Rbrack:25106})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420419ba0), Lbrack:25092, Index:(*ast.Ident)(0xc420419bc0), Rbrack:25106})[0mfmOperatorPtr__setDT(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->KSR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420419d80), Lbrack:25171, Index:(*ast.Ident)(0xc420419da0), Rbrack:25185})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420419d80), Lbrack:25171, Index:(*ast.Ident)(0xc420419da0), Rbrack:25185})[0mfmOperatorPtr__setKSR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->MULT) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420419f60), Lbrack:25252, Index:(*ast.Ident)(0xc420419f80), Rbrack:25266})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420419f60), Lbrack:25252, Index:(*ast.Ident)(0xc420419f80), Rbrack:25266})[0mfmOperatorPtr__setMULT(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->KSL) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424140), Lbrack:25333, Index:(*ast.Ident)(0xc420424160), Rbrack:25347})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424140), Lbrack:25333, Index:(*ast.Ident)(0xc420424160), Rbrack:25347})[0mfmOperatorPtr__setKSL(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->TL) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424320), Lbrack:25412, Index:(*ast.Ident)(0xc420424340), Rbrack:25426})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424320), Lbrack:25412, Index:(*ast.Ident)(0xc420424340), Rbrack:25426})[0mfmOperatorPtr__setTL(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->AR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424500), Lbrack:25490, Index:(*ast.Ident)(0xc420424520), Rbrack:25504})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424500), Lbrack:25490, Index:(*ast.Ident)(0xc420424520), Rbrack:25504})[0mfmOperatorPtr__setAR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204246e0), Lbrack:25568, Index:(*ast.Ident)(0xc420424700), Rbrack:25582})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204246e0), Lbrack:25568, Index:(*ast.Ident)(0xc420424700), Rbrack:25582})[0mfmOperatorPtr__setDR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->SL) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204248c0), Lbrack:25646, Index:(*ast.Ident)(0xc4204248e0), Rbrack:25660})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204248c0), Lbrack:25646, Index:(*ast.Ident)(0xc4204248e0), Rbrack:25660})[0mfmOperatorPtr__setSL(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->SR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424aa0), Lbrack:25724, Index:(*ast.Ident)(0xc420424ac0), Rbrack:25738})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424aa0), Lbrack:25724, Index:(*ast.Ident)(0xc420424ac0), Rbrack:25738})[0mfmOperatorPtr__setSR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->RR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424c80), Lbrack:25802, Index:(*ast.Ident)(0xc420424ca0), Rbrack:25816})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424c80), Lbrack:25802, Index:(*ast.Ident)(0xc420424ca0), Rbrack:25816})[0mfmOperatorPtr__setRR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->XOF) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424e60), Lbrack:25881, Index:(*ast.Ident)(0xc420424e80), Rbrack:25895})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420424e60), Lbrack:25881, Index:(*ast.Ident)(0xc420424e80), Rbrack:25895})[0mfmOperatorPtr__setXOF(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->WS) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420425040), Lbrack:25960, Index:(*ast.Ident)(0xc420425060), Rbrack:25974})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420425040), Lbrack:25960, Index:(*ast.Ident)(0xc420425060), Rbrack:25974})[0mfmOperatorPtr__setWS(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->FB) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420425220), Lbrack:26038, Index:(*ast.Ident)(0xc420425240), Rbrack:26052})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420425220), Lbrack:26038, Index:(*ast.Ident)(0xc420425240), Rbrack:26052})[0mfmOperatorPtr__setFB(regs->chip->channels[channel]->operators[operatorIndex], v);
	}
}

// WriteTL は、TLレジスタに値を書き込みます。
void RegistersPtr__WriteTL(Registers *regs, int channel, int operatorIndex, int tlCarrier, int tlModulator) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:26245, Call:(*ast.CallExpr)(0xc420417500)})[0m
	if (regs->chip->channels[channel]->operators[operatorIndex]->isModulator) {
		RegistersPtr__WriteOperator(regs, channel, operatorIndex, ymf->TL, tlModulator);
	} else {
		RegistersPtr__WriteOperator(regs, channel, operatorIndex, ymf->TL, tlCarrier);
	}
}

// DebugSetMIDIChannel は、チャンネルを使用しているMIDIチャンネル番号をデバッグ用にセットします。
void RegistersPtr__DebugSetMIDIChannel(Registers *regs, int channel, int midiChannel) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:26722, Call:(*ast.CallExpr)(0xc4204177c0)})[0m
	regs->chip->channels[channel]->midiChannelID = midiChannel;
}

// WriteChannel は、チャンネルレジスタに値を書き込みます。
void RegistersPtr__WriteChannel(Registers *regs, int channel, ymf::ChRegister offset, int v) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:26999, Call:(*ast.CallExpr)(0xc4204179c0)})[0m
	auto __tag = offset;
	if (__tag == ymf->KON) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428380), Lbrack:27082, Index:(*ast.Ident)(0xc4204283a0), Rbrack:27090})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428380), Lbrack:27082, Index:(*ast.Ident)(0xc4204283a0), Rbrack:27090})[0mChannelPtr__setKON(regs->chip->channels[channel], v);
	} else if (__tag == ymf->BLOCK) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428500), Lbrack:27139, Index:(*ast.Ident)(0xc420428520), Rbrack:27147})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428500), Lbrack:27139, Index:(*ast.Ident)(0xc420428520), Rbrack:27147})[0mChannelPtr__setBLOCK(regs->chip->channels[channel], v);
	} else if (__tag == ymf->FNUM) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204286a0), Lbrack:27197, Index:(*ast.Ident)(0xc4204286c0), Rbrack:27205})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204286a0), Lbrack:27197, Index:(*ast.Ident)(0xc4204286c0), Rbrack:27205})[0mChannelPtr__setFNUM(regs->chip->channels[channel], v);
	} else if (__tag == ymf->ALG) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428820), Lbrack:27253, Index:(*ast.Ident)(0xc420428840), Rbrack:27261})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428820), Lbrack:27253, Index:(*ast.Ident)(0xc420428840), Rbrack:27261})[0mChannelPtr__setALG(regs->chip->channels[channel], v);
	} else if (__tag == ymf->LFO) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204289a0), Lbrack:27308, Index:(*ast.Ident)(0xc4204289c0), Rbrack:27316})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204289a0), Lbrack:27308, Index:(*ast.Ident)(0xc4204289c0), Rbrack:27316})[0mChannelPtr__setLFO(regs->chip->channels[channel], v);
	} else if (__tag == ymf->PANPOT) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428b20), Lbrack:27366, Index:(*ast.Ident)(0xc420428b40), Rbrack:27374})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428b20), Lbrack:27366, Index:(*ast.Ident)(0xc420428b40), Rbrack:27374})[0mChannelPtr__setPANPOT(regs->chip->channels[channel], v);
	} else if (__tag == ymf->CHPAN) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428ca0), Lbrack:27426, Index:(*ast.Ident)(0xc420428cc0), Rbrack:27434})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428ca0), Lbrack:27426, Index:(*ast.Ident)(0xc420428cc0), Rbrack:27434})[0mChannelPtr__setCHPAN(regs->chip->channels[channel], v);
	} else if (__tag == ymf->VOLUME) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428e20), Lbrack:27486, Index:(*ast.Ident)(0xc420428e40), Rbrack:27494})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428e20), Lbrack:27486, Index:(*ast.Ident)(0xc420428e40), Rbrack:27494})[0mChannelPtr__setVOLUME(regs->chip->channels[channel], v);
	} else if (__tag == ymf->EXPRESSION) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428fa0), Lbrack:27551, Index:(*ast.Ident)(0xc420428fc0), Rbrack:27559})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420428fa0), Lbrack:27551, Index:(*ast.Ident)(0xc420428fc0), Rbrack:27559})[0mChannelPtr__setEXPRESSION(regs->chip->channels[channel], v);
	} else if (__tag == ymf->VELOCITY) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420429120), Lbrack:27618, Index:(*ast.Ident)(0xc420429140), Rbrack:27626})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420429120), Lbrack:27618, Index:(*ast.Ident)(0xc420429140), Rbrack:27626})[0mChannelPtr__setVELOCITY(regs->chip->channels[channel], v);
	} else if (__tag == ymf->BO) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204292a0), Lbrack:27677, Index:(*ast.Ident)(0xc4204292c0), Rbrack:27685})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4204292a0), Lbrack:27677, Index:(*ast.Ident)(0xc4204292c0), Rbrack:27685})[0mChannelPtr__setBO(regs->chip->channels[channel], v);
	} else if (__tag == ymf->RESET) {
		if (v != 0) {
			[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420429460), Lbrack:27748, Index:(*ast.Ident)(0xc420429480), Rbrack:27756})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420429460), Lbrack:27748, Index:(*ast.Ident)(0xc420429480), Rbrack:27756})[0mChannelPtr__resetAll(regs->chip->channels[channel]);
		}
	}
}

}

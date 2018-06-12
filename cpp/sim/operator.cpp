#pragma once
#include "./operator.h"

namespace sim {



operator *newOperator(int channelID, int operatorIndex, Chip *chip) {
	return __ptr((const operator){
		chip: chip,
		channelID: channelID,
		operatorIndex: operatorIndex,
		phaseGenerator: newPhaseGenerator(chip->sampleRate),
		envelopeGenerator: newEnvelopeGenerator(chip->sampleRate),
		isModulator: false,
		bo: 1,
	});
}

void operatorPtr__reset(operator *op) {
	phaseGeneratorPtr__reset(op->phaseGenerator);
	envelopeGeneratorPtr__reset(op->envelopeGenerator);
}

void operatorPtr__resetAll(operator *op) {
	op->bo = 1;
	phaseGeneratorPtr__resetAll(op->phaseGenerator);
	envelopeGeneratorPtr__resetAll(op->envelopeGenerator);
}

string operatorPtr__dump(operator *op) {
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
	auto phstr = [41mobjectOf[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:18906, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc42011bb80)})[0m[41mCallExpr[0m[31m<<nil>>(<nil>)[0m[41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:18906, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc42011bb80)})[0m("        ");
	phstr[phase >> (ymfdata->WaveformLenBits - 3)] = string.byte("|");
	return fmt::Sprintf("%d: %s mul=%02d ws=%02d adssr=%02d,%02d,%02d,%02d,%02d tl=%f am=%s vb=%s dt=%d ksr=%d fb=%3.2f ksn=%02d ksl=%f st=%s ph=%s lv=%-03d %s", op->operatorIndex, cm, op->mult, op->ws, op->ar, op->dr, op->sl, op->sr, op->rr, eg->tlCoef, am, vb, op->dt, op->ksr, op->feedbackCoef, op->keyScaleNumber, eg->kslCoef, stage__String(eg->stage), string(phstr), int(math::Floor(lvdb)), lvstr);
}

void operatorPtr__setEAM(operator *op, int v) {
	op->envelopeGenerator->eam = v != 0;
}

void operatorPtr__setEVB(operator *op, int v) {
	op->phaseGenerator->evb = v != 0;
}

void operatorPtr__setDAM(operator *op, int v) {
	op->envelopeGenerator->dam = v;
}

void operatorPtr__setDVB(operator *op, int v) {
	op->phaseGenerator->dvb = v;
}

void operatorPtr__setDT(operator *op, int v) {
	op->dt = v;
	operatorPtr__updateFrequency(op);
}

void operatorPtr__setKSR(operator *op, int v) {
	op->ksr = v;
	operatorPtr__updateEnvelope(op);
}

void operatorPtr__setMULT(operator *op, int v) {
	op->mult = v;
	operatorPtr__updateFrequency(op);
}

void operatorPtr__setKSL(operator *op, int v) {
	op->ksl = v;
	envelopeGeneratorPtr__setKeyScalingLevel(op->envelopeGenerator, op->fnum, op->block, op->bo, op->ksl);
}

void operatorPtr__setTL(operator *op, int v) {
	envelopeGeneratorPtr__setTotalLevel(op->envelopeGenerator, v);
}

void operatorPtr__setAR(operator *op, int v) {
	op->ar = v;
	envelopeGeneratorPtr__setActualAR(op->envelopeGenerator, op->ar, op->ksr, op->keyScaleNumber);
}

void operatorPtr__setDR(operator *op, int v) {
	op->dr = v;
	envelopeGeneratorPtr__setActualDR(op->envelopeGenerator, op->dr, op->ksr, op->keyScaleNumber);
}

void operatorPtr__setSL(operator *op, int v) {
	op->sl = v;
	envelopeGeneratorPtr__setActualSustainLevel(op->envelopeGenerator, op->sl);
}

void operatorPtr__setSR(operator *op, int v) {
	op->sr = v;
	envelopeGeneratorPtr__setActualSR(op->envelopeGenerator, op->sr, op->ksr, op->keyScaleNumber);
}

void operatorPtr__setRR(operator *op, int v) {
	op->rr = v;
	envelopeGeneratorPtr__setActualRR(op->envelopeGenerator, op->rr, op->ksr, op->keyScaleNumber);
}

void operatorPtr__setXOF(operator *op, int v) {
	op->xof = v;
}

void operatorPtr__setWS(operator *op, int v) {
	op->ws = v;
}

void operatorPtr__setFB(operator *op, int v) {
	op->feedbackCoef = ymfdata->FeedbackTable[v];
}

float64 operatorPtr__next(operator *op, int modIndex, float64 modulator) {
	auto phaseFrac64 = phaseGeneratorPtr__getPhase(op->phaseGenerator, modIndex);
	if (op->envelopeGenerator->stage == stageOff) {
		return 0;
	}
	auto envelope = envelopeGeneratorPtr__getEnvelope(op->envelopeGenerator, modIndex);
	auto sampleIndex = uint64(phaseFrac64) >> ymfdata->WaveformIndexShift;
	sampleIndex = uint64((modulator + ymfdata->WaveformLen)*ymfdata->WaveformLen);
	return ymfdata->Waveforms[op->ws][sampleIndex & 1023]*envelope;
}

void operatorPtr__keyOn(operator *op) {
	if (0 < op->ar) {
		envelopeGeneratorPtr__keyOn(op->envelopeGenerator);
	} else {
		op->envelopeGenerator->stage = stageOff;
	}
}

void operatorPtr__keyOff(operator *op) {
	if (op->xof == 0) {
		envelopeGeneratorPtr__keyOff(op->envelopeGenerator);
	}
}

void operatorPtr__setFrequency(operator *op, int fnum, int blk, int bo) {
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
	operatorPtr__updateFrequency(op);
	operatorPtr__updateEnvelope(op);
	envelopeGeneratorPtr__setKeyScalingLevel(op->envelopeGenerator, op->fnum, op->block, op->bo, op->ksl);
}

void operatorPtr__updateFrequency(operator *op) {
	phaseGeneratorPtr__setFrequency(op->phaseGenerator, op->fnum, op->block, op->bo, op->mult, op->dt);
}

void operatorPtr__updateEnvelope(operator *op) {
	envelopeGeneratorPtr__setActualAR(op->envelopeGenerator, op->ar, op->ksr, op->keyScaleNumber);
	envelopeGeneratorPtr__setActualDR(op->envelopeGenerator, op->dr, op->ksr, op->keyScaleNumber);
	envelopeGeneratorPtr__setActualSR(op->envelopeGenerator, op->sr, op->ksr, op->keyScaleNumber);
	envelopeGeneratorPtr__setActualRR(op->envelopeGenerator, op->rr, op->ksr, op->keyScaleNumber);
}

}

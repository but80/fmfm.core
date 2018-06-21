#pragma once
#include "go2cpp.h"
namespace sim {

#include "fmt.h"
#include "math.h"
#include "strings.h"
#include "ymf/ymfdata.h"
namespace ymfdata = ymf::ymfdata;
typedef struct __Channel {
	int channelID;
	int midiChannelID;
	Chip *chip;
	int fnum;
	int kon;
	int block;
	int alg;
	int panpot;
	int chpan;
	int volume;
	int expression;
	int velocity;
	int bo;
	float64 feedbackBlendPrev;
	float64 feedbackBlendCurr;
	float64 feedback1Prev;
	float64 feedback1Curr;
	float64 feedback3Prev;
	float64 feedback3Curr;
	float64 feedbackOut1;
	float64 feedbackOut3;
	float64 attenuationCoef;
	gopkg_in::but80::fmfm_core_v1::ymf::ymfdata::Frac64 modIndexFrac64;
	gopkg_in::but80::fmfm_core_v1::ymf::ymfdata::Frac64 lfoFrequency;
	float64 panCoefL;
	float64 panCoefR;
	operator *operators[4];
} Channel;
Channel *newChannel(int channelID, Chip *chip);
void ChannelPtr__reset(Channel *ch);
void ChannelPtr__resetAll(Channel *ch);
bool ChannelPtr__isOff(Channel *ch);
float64 ChannelPtr__currentLevel(Channel *ch);
string ChannelPtr__dump(Channel *ch);
void ChannelPtr__setKON(Channel *ch, int v);
void ChannelPtr__keyOn(Channel *ch);
void ChannelPtr__keyOff(Channel *ch);
void ChannelPtr__setBLOCK(Channel *ch, int v);
void ChannelPtr__setFNUM(Channel *ch, int v);
void ChannelPtr__setALG(Channel *ch, int v);
void ChannelPtr__setLFO(Channel *ch, int v);
void ChannelPtr__setPANPOT(Channel *ch, int v);
void ChannelPtr__setCHPAN(Channel *ch, int v);
void ChannelPtr__updatePanCoef(Channel *ch);
void ChannelPtr__setVOLUME(Channel *ch, int v);
void ChannelPtr__setEXPRESSION(Channel *ch, int v);
void ChannelPtr__setVELOCITY(Channel *ch, int v);
void ChannelPtr__updateAttenuation(Channel *ch);
void ChannelPtr__setBO(Channel *ch, int v);
MULTIRESULT ChannelPtr__next(Channel *ch);
void ChannelPtr__updateFrequency(Channel *ch);
}
#pragma once
#include "go2cpp.h"
namespace sim {

#include "fmt.h"
#include "math.h"
#include "sort.h"
#include "sync.h"
#include "ymf/ymfdata.h"
namespace ymfdata = ymf::ymfdata;
typedef struct __Chip {
	sync::Mutex Mutex;
	float64 sampleRate;
	float64 totalLevel;
	int dumpMIDIChannel;
	std::vector<Channel*> channels;
	std::vector<float64> currentOutput;
} Chip;
Chip *NewChip(float64 sampleRate, float64 totalLevel, int dumpMIDIChannel);
MULTIRESULT ChipPtr__Next(Chip *chip);
void ChipPtr__initChannels(Chip *chip);
}
#pragma once
#include "go2cpp.h"
namespace sim {

#include "math.h"
#include "ymf/ymfdata.h"
namespace ymfdata = ymf::ymfdata;
typedef int stage;
string stage__String(stage s);
typedef struct __envelopeGenerator {
	float64 sampleRate;
	stage stage;
	bool eam;
	int dam;
	float64 arDiffPerSample;
	float64 drCoefPerSample;
	float64 srCoefPerSample;
	float64 rrCoefPerSample;
	float64 kslCoef;
	float64 tlCoef;
	float64 kslTlCoef;
	float64 sustainLevel;
	float64 currentLevel;
} envelopeGenerator;
envelopeGenerator *newEnvelopeGenerator(float64 sampleRate);
void envelopeGeneratorPtr__reset(envelopeGenerator *eg);
void envelopeGeneratorPtr__resetAll(envelopeGenerator *eg);
void envelopeGeneratorPtr__setActualSustainLevel(envelopeGenerator *eg, int sl);
void envelopeGeneratorPtr__setTotalLevel(envelopeGenerator *eg, int tl);
void envelopeGeneratorPtr__setKeyScalingLevel(envelopeGenerator *eg, int fnum, int block, int bo, int ksl);
void envelopeGeneratorPtr__setActualAR(envelopeGenerator *eg, int attackRate, int ksr, int keyScaleNumber);
void envelopeGeneratorPtr__setActualDR(envelopeGenerator *eg, int dr, int ksr, int keyScaleNumber);
void envelopeGeneratorPtr__setActualSR(envelopeGenerator *eg, int sr, int ksr, int keyScaleNumber);
void envelopeGeneratorPtr__setActualRR(envelopeGenerator *eg, int rr, int ksr, int keyScaleNumber);
float64 envelopeGeneratorPtr__getEnvelope(envelopeGenerator *eg, int tremoloIndex);
void envelopeGeneratorPtr__keyOn(envelopeGenerator *eg);
void envelopeGeneratorPtr__keyOff(envelopeGenerator *eg);
}
#pragma once
#include "go2cpp.h"
namespace sim {

#include "fmt.h"
#include "math.h"
#include "strings.h"
#include "ymf/ymfdata.h"
namespace ymfdata = ymf::ymfdata;
typedef struct __operator {
	bool isModulator;
	int dt;
	int ksr;
	int mult;
	int ksl;
	int ar;
	int dr;
	int sl;
	int sr;
	int rr;
	int xof;
	int ws;
	float64 feedbackCoef;
	int keyScaleNumber;
	int fnum;
	int block;
	int bo;
	envelopeGenerator *envelopeGenerator;
	Chip *chip;
	int channelID;
	int operatorIndex;
	phaseGenerator *phaseGenerator;
} operator;
operator *newOperator(int channelID, int operatorIndex, Chip *chip);
void operatorPtr__reset(operator *op);
void operatorPtr__resetAll(operator *op);
string operatorPtr__dump(operator *op);
void operatorPtr__setEAM(operator *op, int v);
void operatorPtr__setEVB(operator *op, int v);
void operatorPtr__setDAM(operator *op, int v);
void operatorPtr__setDVB(operator *op, int v);
void operatorPtr__setDT(operator *op, int v);
void operatorPtr__setKSR(operator *op, int v);
void operatorPtr__setMULT(operator *op, int v);
void operatorPtr__setKSL(operator *op, int v);
void operatorPtr__setTL(operator *op, int v);
void operatorPtr__setAR(operator *op, int v);
void operatorPtr__setDR(operator *op, int v);
void operatorPtr__setSL(operator *op, int v);
void operatorPtr__setSR(operator *op, int v);
void operatorPtr__setRR(operator *op, int v);
void operatorPtr__setXOF(operator *op, int v);
void operatorPtr__setWS(operator *op, int v);
void operatorPtr__setFB(operator *op, int v);
float64 operatorPtr__next(operator *op, int modIndex, float64 modulator);
void operatorPtr__keyOn(operator *op);
void operatorPtr__keyOff(operator *op);
void operatorPtr__setFrequency(operator *op, int fnum, int blk, int bo);
void operatorPtr__updateFrequency(operator *op);
void operatorPtr__updateEnvelope(operator *op);
}
#pragma once
#include "go2cpp.h"
namespace sim {

#include "ymf/ymfdata.h"
namespace ymfdata = ymf::ymfdata;
typedef struct __phaseGenerator {
	float64 sampleRate;
	bool evb;
	int dvb;
	gopkg_in::but80::fmfm_core_v1::ymf::ymfdata::Frac64 phaseFrac64;
	gopkg_in::but80::fmfm_core_v1::ymf::ymfdata::Frac64 phaseIncrementFrac64;
} phaseGenerator;
phaseGenerator *newPhaseGenerator(float64 sampleRate);
void phaseGeneratorPtr__reset(phaseGenerator *pg);
void phaseGeneratorPtr__resetAll(phaseGenerator *pg);
void phaseGeneratorPtr__setFrequency(phaseGenerator *pg, int fnum, int block, int bo, int mult, int dt);
gopkg_in::but80::fmfm_core_v1::ymf::ymfdata::Frac64 phaseGeneratorPtr__getPhase(phaseGenerator *pg, int vibratoIndex);
}
#pragma once
#include "go2cpp.h"
namespace sim {

#include "ymf.h"
typedef struct __Registers {
	Chip *chip;
} Registers;
Registers *NewRegisters(Chip *chip);
void RegistersPtr__WriteOperator(Registers *regs, int channel, int operatorIndex, gopkg_in::but80::fmfm_core_v1::ymf::OpRegister offset, int v);
void RegistersPtr__WriteTL(Registers *regs, int channel, int operatorIndex, int tlCarrier, int tlModulator);
void RegistersPtr__DebugSetMIDIChannel(Registers *regs, int channel, int midiChannel);
void RegistersPtr__WriteChannel(Registers *regs, int channel, gopkg_in::but80::fmfm_core_v1::ymf::ChRegister offset, int v);
}

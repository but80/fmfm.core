// namespace sim
#pragma once
#include "go2cpp.h"
#include "fmt.h"
#include "math.h"
#include "strings.h"
#include "ymf/ymfdata.h"
namespace ymfdata = ymf::ymfdata;
#include "sort.h"
#include "sync.h"
#include "ymf.h"
namespace sim {

class Channel;
class Chip;
typedef int stage;
class envelopeGenerator;
class fmOperator;
class phaseGenerator;
class Registers;
const auto noModulator = 0;
/* map[string]string{"fmt":"fmt", "math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata"} */
class Channel {
public:
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
	ymfdata::Frac64 modIndexFrac64;
	ymfdata::Frac64 lfoFrequency;
	float64 panCoefL;
	float64 panCoefR;
	fmOperator *operators[4];
};
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
class nextResult {
	float64 r0;
	float64 r1;
};
nextResult ChannelPtr__next(Channel *ch);
void ChannelPtr__updateFrequency(Channel *ch);
/* map[string]string{"sort":"sort", "sync":"sync", "fmt":"fmt", "math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata"} */
class Chip {
public:
	sync::Mutex Mutex;
	float64 sampleRate;
	float64 totalLevel;
	int dumpMIDIChannel;
	std::vector<Channel*> channels;
	std::vector<float64> currentOutput;
};
Chip *NewChip(float64 sampleRate, float64 totalLevel, int dumpMIDIChannel);
class NextResult {
	float64 r0;
	float64 r1;
};
NextResult ChipPtr__Next(Chip *chip);
void ChipPtr__initChannels(Chip *chip);
/* map[string]string{"sort":"sort", "sync":"sync", "fmt":"fmt", "math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata"} */
const stage stageOff = 0;
const auto stageAttack = 1;
const auto stageDecay = 2;
const auto stageSustain = 3;
const auto stageRelease = 4;
string stage__String(stage s);
const auto epsilon = 1.0/32768.0;
/* map[string]string{"fmt":"fmt", "math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata", "sort":"sort", "sync":"sync"} */
class envelopeGenerator {
public:
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
};
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
/* map[string]string{"gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata", "sort":"sort", "sync":"sync", "fmt":"fmt", "math":"math", "strings":"strings"} */
class fmOperator {
public:
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
};
fmOperator *newOperator(int channelID, int operatorIndex, Chip *chip);
void fmOperatorPtr__reset(fmOperator *op);
void fmOperatorPtr__resetAll(fmOperator *op);
string fmOperatorPtr__dump(fmOperator *op);
void fmOperatorPtr__setEAM(fmOperator *op, int v);
void fmOperatorPtr__setEVB(fmOperator *op, int v);
void fmOperatorPtr__setDAM(fmOperator *op, int v);
void fmOperatorPtr__setDVB(fmOperator *op, int v);
void fmOperatorPtr__setDT(fmOperator *op, int v);
void fmOperatorPtr__setKSR(fmOperator *op, int v);
void fmOperatorPtr__setMULT(fmOperator *op, int v);
void fmOperatorPtr__setKSL(fmOperator *op, int v);
void fmOperatorPtr__setTL(fmOperator *op, int v);
void fmOperatorPtr__setAR(fmOperator *op, int v);
void fmOperatorPtr__setDR(fmOperator *op, int v);
void fmOperatorPtr__setSL(fmOperator *op, int v);
void fmOperatorPtr__setSR(fmOperator *op, int v);
void fmOperatorPtr__setRR(fmOperator *op, int v);
void fmOperatorPtr__setXOF(fmOperator *op, int v);
void fmOperatorPtr__setWS(fmOperator *op, int v);
void fmOperatorPtr__setFB(fmOperator *op, int v);
float64 fmOperatorPtr__next(fmOperator *op, int modIndex, float64 modulator);
void fmOperatorPtr__keyOn(fmOperator *op);
void fmOperatorPtr__keyOff(fmOperator *op);
void fmOperatorPtr__setFrequency(fmOperator *op, int fnum, int blk, int bo);
void fmOperatorPtr__updateFrequency(fmOperator *op);
void fmOperatorPtr__updateEnvelope(fmOperator *op);
/* map[string]string{"math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata", "sort":"sort", "sync":"sync", "fmt":"fmt"} */
class phaseGenerator {
public:
	float64 sampleRate;
	bool evb;
	int dvb;
	ymfdata::Frac64 phaseFrac64;
	ymfdata::Frac64 phaseIncrementFrac64;
};
phaseGenerator *newPhaseGenerator(float64 sampleRate);
void phaseGeneratorPtr__reset(phaseGenerator *pg);
void phaseGeneratorPtr__resetAll(phaseGenerator *pg);
void phaseGeneratorPtr__setFrequency(phaseGenerator *pg, int fnum, int block, int bo, int mult, int dt);
ymfdata::Frac64 phaseGeneratorPtr__getPhase(phaseGenerator *pg, int vibratoIndex);
/* map[string]string{"fmt":"fmt", "math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata", "sort":"sort", "sync":"sync", "gopkg.in/but80/fmfm.core.v1/ymf":"ymf"} */
class Registers {
public:
	Chip *chip;
};
Registers *NewRegisters(Chip *chip);
void RegistersPtr__WriteOperator(Registers *regs, int channel, int operatorIndex, ymf::OpRegister offset, int v);
void RegistersPtr__WriteTL(Registers *regs, int channel, int operatorIndex, int tlCarrier, int tlModulator);
void RegistersPtr__DebugSetMIDIChannel(Registers *regs, int channel, int midiChannel);
void RegistersPtr__WriteChannel(Registers *regs, int channel, ymf::ChRegister offset, int v);
}

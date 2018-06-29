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

struct Channel;
struct Chip;
typedef int stage;
struct envelopeGenerator;
struct fmOperator;
struct phaseGenerator;
struct Registers;
const auto noModulator = 0;
/* map[string]string{"fmt":"fmt", "math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata"} */
struct Channel {
	int channelID;
	int midiChannelID;
	std::shared_ptr<Chip> chip;
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
	std::shared_ptr<fmOperator> operators[4];
};
std::shared_ptr<Channel> newChannel(int channelID, std::shared_ptr<Chip> chip);
void ChannelPtr__reset(std::shared_ptr<Channel> ch);
void ChannelPtr__resetAll(std::shared_ptr<Channel> ch);
bool ChannelPtr__isOff(std::shared_ptr<Channel> ch);
float64 ChannelPtr__currentLevel(std::shared_ptr<Channel> ch);
string ChannelPtr__dump(std::shared_ptr<Channel> ch);
void ChannelPtr__setKON(std::shared_ptr<Channel> ch, int v);
void ChannelPtr__keyOn(std::shared_ptr<Channel> ch);
void ChannelPtr__keyOff(std::shared_ptr<Channel> ch);
void ChannelPtr__setBLOCK(std::shared_ptr<Channel> ch, int v);
void ChannelPtr__setFNUM(std::shared_ptr<Channel> ch, int v);
void ChannelPtr__setALG(std::shared_ptr<Channel> ch, int v);
void ChannelPtr__setLFO(std::shared_ptr<Channel> ch, int v);
void ChannelPtr__setPANPOT(std::shared_ptr<Channel> ch, int v);
void ChannelPtr__setCHPAN(std::shared_ptr<Channel> ch, int v);
void ChannelPtr__updatePanCoef(std::shared_ptr<Channel> ch);
void ChannelPtr__setVOLUME(std::shared_ptr<Channel> ch, int v);
void ChannelPtr__setEXPRESSION(std::shared_ptr<Channel> ch, int v);
void ChannelPtr__setVELOCITY(std::shared_ptr<Channel> ch, int v);
void ChannelPtr__updateAttenuation(std::shared_ptr<Channel> ch);
void ChannelPtr__setBO(std::shared_ptr<Channel> ch, int v);
struct ChannelPtr__next__result {
	float64 r0;
	float64 r1;
};
ChannelPtr__next__result ChannelPtr__next(std::shared_ptr<Channel> ch);
void ChannelPtr__updateFrequency(std::shared_ptr<Channel> ch);
/* map[string]string{"strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata", "sort":"sort", "sync":"sync", "fmt":"fmt", "math":"math"} */
struct Chip {
	sync::Mutex Mutex;
	float64 sampleRate;
	float64 totalLevel;
	int dumpMIDIChannel;
	std::vector<std::shared_ptr<Channel>> channels;
	std::vector<float64> currentOutput;
};
std::shared_ptr<Chip> NewChip(float64 sampleRate, float64 totalLevel, int dumpMIDIChannel);
extern int debugDumpCount;
struct ChipPtr__Next__result {
	float64 r0;
	float64 r1;
};
ChipPtr__Next__result ChipPtr__Next(std::shared_ptr<Chip> chip);
void ChipPtr__initChannels(std::shared_ptr<Chip> chip);
/* map[string]string{"sync":"sync", "fmt":"fmt", "math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata", "sort":"sort"} */
const stage stageOff = 0;
const auto stageAttack = 1;
const auto stageDecay = 2;
const auto stageSustain = 3;
const auto stageRelease = 4;
string stage__String(stage s);
const auto epsilon = 1.0/32768.0;
/* map[string]string{"sync":"sync", "fmt":"fmt", "math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata", "sort":"sort"} */
struct envelopeGenerator {
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
std::shared_ptr<envelopeGenerator> newEnvelopeGenerator(float64 sampleRate);
void envelopeGeneratorPtr__reset(std::shared_ptr<envelopeGenerator> eg);
void envelopeGeneratorPtr__resetAll(std::shared_ptr<envelopeGenerator> eg);
void envelopeGeneratorPtr__setActualSustainLevel(std::shared_ptr<envelopeGenerator> eg, int sl);
void envelopeGeneratorPtr__setTotalLevel(std::shared_ptr<envelopeGenerator> eg, int tl);
void envelopeGeneratorPtr__setKeyScalingLevel(std::shared_ptr<envelopeGenerator> eg, int fnum, int block, int bo, int ksl);
void envelopeGeneratorPtr__setActualAR(std::shared_ptr<envelopeGenerator> eg, int attackRate, int ksr, int keyScaleNumber);
void envelopeGeneratorPtr__setActualDR(std::shared_ptr<envelopeGenerator> eg, int dr, int ksr, int keyScaleNumber);
void envelopeGeneratorPtr__setActualSR(std::shared_ptr<envelopeGenerator> eg, int sr, int ksr, int keyScaleNumber);
void envelopeGeneratorPtr__setActualRR(std::shared_ptr<envelopeGenerator> eg, int rr, int ksr, int keyScaleNumber);
float64 envelopeGeneratorPtr__getEnvelope(std::shared_ptr<envelopeGenerator> eg, int tremoloIndex);
void envelopeGeneratorPtr__keyOn(std::shared_ptr<envelopeGenerator> eg);
void envelopeGeneratorPtr__keyOff(std::shared_ptr<envelopeGenerator> eg);
extern float64 decayDBPerSecAt4[16][2];
extern float64 attackTimeSecAt1[9][2];
/* map[string]string{"math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata", "sort":"sort", "sync":"sync", "fmt":"fmt"} */
struct fmOperator {
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
	std::shared_ptr<envelopeGenerator> envelopeGenerator;
	std::shared_ptr<Chip> chip;
	int channelID;
	int operatorIndex;
	std::shared_ptr<phaseGenerator> phaseGenerator;
};
std::shared_ptr<fmOperator> newOperator(int channelID, int operatorIndex, std::shared_ptr<Chip> chip);
void fmOperatorPtr__reset(std::shared_ptr<fmOperator> op);
void fmOperatorPtr__resetAll(std::shared_ptr<fmOperator> op);
string fmOperatorPtr__dump(std::shared_ptr<fmOperator> op);
void fmOperatorPtr__setEAM(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setEVB(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setDAM(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setDVB(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setDT(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setKSR(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setMULT(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setKSL(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setTL(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setAR(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setDR(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setSL(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setSR(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setRR(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setXOF(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setWS(std::shared_ptr<fmOperator> op, int v);
void fmOperatorPtr__setFB(std::shared_ptr<fmOperator> op, int v);
float64 fmOperatorPtr__next(std::shared_ptr<fmOperator> op, int modIndex, float64 modulator);
void fmOperatorPtr__keyOn(std::shared_ptr<fmOperator> op);
void fmOperatorPtr__keyOff(std::shared_ptr<fmOperator> op);
void fmOperatorPtr__setFrequency(std::shared_ptr<fmOperator> op, int fnum, int blk, int bo);
void fmOperatorPtr__updateFrequency(std::shared_ptr<fmOperator> op);
void fmOperatorPtr__updateEnvelope(std::shared_ptr<fmOperator> op);
/* map[string]string{"sort":"sort", "sync":"sync", "fmt":"fmt", "math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata"} */
struct phaseGenerator {
	float64 sampleRate;
	bool evb;
	int dvb;
	ymfdata::Frac64 phaseFrac64;
	ymfdata::Frac64 phaseIncrementFrac64;
};
std::shared_ptr<phaseGenerator> newPhaseGenerator(float64 sampleRate);
void phaseGeneratorPtr__reset(std::shared_ptr<phaseGenerator> pg);
void phaseGeneratorPtr__resetAll(std::shared_ptr<phaseGenerator> pg);
void phaseGeneratorPtr__setFrequency(std::shared_ptr<phaseGenerator> pg, int fnum, int block, int bo, int mult, int dt);
ymfdata::Frac64 phaseGeneratorPtr__getPhase(std::shared_ptr<phaseGenerator> pg, int vibratoIndex);
/* map[string]string{"math":"math", "strings":"strings", "gopkg.in/but80/fmfm.core.v1/ymf/ymfdata":"ymfdata", "sort":"sort", "sync":"sync", "gopkg.in/but80/fmfm.core.v1/ymf":"ymf", "fmt":"fmt"} */
struct Registers {
	std::shared_ptr<Chip> chip;
};
std::shared_ptr<Registers> NewRegisters(std::shared_ptr<Chip> chip);
void RegistersPtr__WriteOperator(std::shared_ptr<Registers> regs, int channel, int operatorIndex, ymf::OpRegister offset, int v);
void RegistersPtr__WriteTL(std::shared_ptr<Registers> regs, int channel, int operatorIndex, int tlCarrier, int tlModulator);
void RegistersPtr__DebugSetMIDIChannel(std::shared_ptr<Registers> regs, int channel, int midiChannel);
void RegistersPtr__WriteChannel(std::shared_ptr<Registers> regs, int channel, ymf::ChRegister offset, int v);
}

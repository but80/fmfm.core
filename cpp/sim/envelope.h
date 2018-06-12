#include "math.h"
#include "ymf/ymfdata.h"
namespace ymfdata = ymf::ymfdata;
// *ast.Ident
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

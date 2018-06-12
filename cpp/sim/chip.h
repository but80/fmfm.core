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
	vector<Channel*> channels;
	vector<float64> currentOutput;
} Chip;

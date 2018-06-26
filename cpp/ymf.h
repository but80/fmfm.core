// namespace ymf
#pragma once
#include "go2cpp.h"
namespace ymf {

typedef int OpRegister;
typedef int ChRegister;
// *ast.InterfaceType
/* map[string]string{} */
const OpRegister EAM = 0;
const auto EVB = 1;
const auto DAM = 2;
const auto DVB = 3;
const auto DT = 4;
const auto KSL = 5;
const auto KSR = 6;
const auto WS = 7;
const auto MULT = 8;
const auto FB = 9;
const auto AR = 10;
const auto DR = 11;
const auto SL = 12;
const auto SR = 13;
const auto RR = 14;
const auto TL = 15;
const auto XOF = 16;
/* map[string]string{} */
const ChRegister KON = 0;
const auto BLOCK = 1;
const auto FNUM = 2;
const auto ALG = 3;
const auto LFO = 4;
const auto PANPOT = 5;
const auto CHPAN = 6;
const auto VOLUME = 7;
const auto EXPRESSION = 8;
const auto VELOCITY = 9;
const auto BO = 10;
const auto RESET = 11;
/* map[string]string{} */
}

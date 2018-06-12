#pragma once
#include "./register.h"

namespace sim {



gopkg_in::but80::fmfm_core_v1::ymf::Registers _ = __ptr((const Registers){
});

// NewRegisters は、新しい Registers を作成します。
Registers *NewRegisters(Chip *chip) {
	return __ptr((const Registers){
		chip: chip,
	});
}

// WriteOperator は、オペレータレジスタに値を書き込みます。
void RegistersPtr__WriteOperator(Registers *regs, int channel, int operatorIndex, gopkg_in::but80::fmfm_core_v1::ymf::OpRegister offset, int v) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:24611, Call:(*ast.CallExpr)(0xc420130e80)})[0m
	auto __tag = offset;
	if (__tag == ymf->EAM) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420136e40), Lbrack:24713, Index:(*ast.Ident)(0xc420136e60), Rbrack:24727})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420136e40), Lbrack:24713, Index:(*ast.Ident)(0xc420136e60), Rbrack:24727})[0moperatorPtr__setEAM(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->EVB) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137020), Lbrack:24793, Index:(*ast.Ident)(0xc420137040), Rbrack:24807})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137020), Lbrack:24793, Index:(*ast.Ident)(0xc420137040), Rbrack:24807})[0moperatorPtr__setEVB(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DAM) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137220), Lbrack:24873, Index:(*ast.Ident)(0xc420137240), Rbrack:24887})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137220), Lbrack:24873, Index:(*ast.Ident)(0xc420137240), Rbrack:24887})[0moperatorPtr__setDAM(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DVB) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137400), Lbrack:24953, Index:(*ast.Ident)(0xc420137420), Rbrack:24967})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137400), Lbrack:24953, Index:(*ast.Ident)(0xc420137420), Rbrack:24967})[0moperatorPtr__setDVB(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DT) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201375e0), Lbrack:25032, Index:(*ast.Ident)(0xc420137600), Rbrack:25046})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201375e0), Lbrack:25032, Index:(*ast.Ident)(0xc420137600), Rbrack:25046})[0moperatorPtr__setDT(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->KSR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201377c0), Lbrack:25111, Index:(*ast.Ident)(0xc4201377e0), Rbrack:25125})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201377c0), Lbrack:25111, Index:(*ast.Ident)(0xc4201377e0), Rbrack:25125})[0moperatorPtr__setKSR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->MULT) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201379a0), Lbrack:25192, Index:(*ast.Ident)(0xc4201379c0), Rbrack:25206})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201379a0), Lbrack:25192, Index:(*ast.Ident)(0xc4201379c0), Rbrack:25206})[0moperatorPtr__setMULT(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->KSL) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137b80), Lbrack:25273, Index:(*ast.Ident)(0xc420137ba0), Rbrack:25287})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137b80), Lbrack:25273, Index:(*ast.Ident)(0xc420137ba0), Rbrack:25287})[0moperatorPtr__setKSL(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->TL) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137d60), Lbrack:25352, Index:(*ast.Ident)(0xc420137d80), Rbrack:25366})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137d60), Lbrack:25352, Index:(*ast.Ident)(0xc420137d80), Rbrack:25366})[0moperatorPtr__setTL(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->AR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137f40), Lbrack:25430, Index:(*ast.Ident)(0xc420137f60), Rbrack:25444})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420137f40), Lbrack:25430, Index:(*ast.Ident)(0xc420137f60), Rbrack:25444})[0moperatorPtr__setAR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420140120), Lbrack:25508, Index:(*ast.Ident)(0xc420140140), Rbrack:25522})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420140120), Lbrack:25508, Index:(*ast.Ident)(0xc420140140), Rbrack:25522})[0moperatorPtr__setDR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->SL) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420140300), Lbrack:25586, Index:(*ast.Ident)(0xc420140320), Rbrack:25600})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420140300), Lbrack:25586, Index:(*ast.Ident)(0xc420140320), Rbrack:25600})[0moperatorPtr__setSL(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->SR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201404e0), Lbrack:25664, Index:(*ast.Ident)(0xc420140500), Rbrack:25678})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201404e0), Lbrack:25664, Index:(*ast.Ident)(0xc420140500), Rbrack:25678})[0moperatorPtr__setSR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->RR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201406c0), Lbrack:25742, Index:(*ast.Ident)(0xc4201406e0), Rbrack:25756})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201406c0), Lbrack:25742, Index:(*ast.Ident)(0xc4201406e0), Rbrack:25756})[0moperatorPtr__setRR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->XOF) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201408a0), Lbrack:25821, Index:(*ast.Ident)(0xc4201408c0), Rbrack:25835})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201408a0), Lbrack:25821, Index:(*ast.Ident)(0xc4201408c0), Rbrack:25835})[0moperatorPtr__setXOF(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->WS) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420140a80), Lbrack:25900, Index:(*ast.Ident)(0xc420140aa0), Rbrack:25914})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420140a80), Lbrack:25900, Index:(*ast.Ident)(0xc420140aa0), Rbrack:25914})[0moperatorPtr__setWS(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->FB) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420140c60), Lbrack:25978, Index:(*ast.Ident)(0xc420140c80), Rbrack:25992})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420140c60), Lbrack:25978, Index:(*ast.Ident)(0xc420140c80), Rbrack:25992})[0moperatorPtr__setFB(regs->chip->channels[channel]->operators[operatorIndex], v);
	}
}

// WriteTL は、TLレジスタに値を書き込みます。
void RegistersPtr__WriteTL(Registers *regs, int channel, int operatorIndex, int tlCarrier, int tlModulator) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:26185, Call:(*ast.CallExpr)(0xc4201318c0)})[0m
	if (regs->chip->channels[channel]->operators[operatorIndex]->isModulator) {
		RegistersPtr__WriteOperator(regs, channel, operatorIndex, ymf->TL, tlModulator);
	} else {
		RegistersPtr__WriteOperator(regs, channel, operatorIndex, ymf->TL, tlCarrier);
	}
}

// DebugSetMIDIChannel は、チャンネルを使用しているMIDIチャンネル番号をデバッグ用にセットします。
void RegistersPtr__DebugSetMIDIChannel(Registers *regs, int channel, int midiChannel) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:26662, Call:(*ast.CallExpr)(0xc420131b80)})[0m
	regs->chip->channels[channel]->midiChannelID = midiChannel;
}

// WriteChannel は、チャンネルレジスタに値を書き込みます。
void RegistersPtr__WriteChannel(Registers *regs, int channel, gopkg_in::but80::fmfm_core_v1::ymf::ChRegister offset, int v) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:26939, Call:(*ast.CallExpr)(0xc420131d80)})[0m
	auto __tag = offset;
	if (__tag == ymf->KON) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420141dc0), Lbrack:27022, Index:(*ast.Ident)(0xc420141de0), Rbrack:27030})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420141dc0), Lbrack:27022, Index:(*ast.Ident)(0xc420141de0), Rbrack:27030})[0mChannelPtr__setKON(regs->chip->channels[channel], v);
	} else if (__tag == ymf->BLOCK) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420141f40), Lbrack:27079, Index:(*ast.Ident)(0xc420141f60), Rbrack:27087})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420141f40), Lbrack:27079, Index:(*ast.Ident)(0xc420141f60), Rbrack:27087})[0mChannelPtr__setBLOCK(regs->chip->channels[channel], v);
	} else if (__tag == ymf->FNUM) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201420e0), Lbrack:27137, Index:(*ast.Ident)(0xc420142100), Rbrack:27145})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201420e0), Lbrack:27137, Index:(*ast.Ident)(0xc420142100), Rbrack:27145})[0mChannelPtr__setFNUM(regs->chip->channels[channel], v);
	} else if (__tag == ymf->ALG) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142260), Lbrack:27193, Index:(*ast.Ident)(0xc420142280), Rbrack:27201})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142260), Lbrack:27193, Index:(*ast.Ident)(0xc420142280), Rbrack:27201})[0mChannelPtr__setALG(regs->chip->channels[channel], v);
	} else if (__tag == ymf->LFO) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201423e0), Lbrack:27248, Index:(*ast.Ident)(0xc420142400), Rbrack:27256})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201423e0), Lbrack:27248, Index:(*ast.Ident)(0xc420142400), Rbrack:27256})[0mChannelPtr__setLFO(regs->chip->channels[channel], v);
	} else if (__tag == ymf->PANPOT) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142560), Lbrack:27306, Index:(*ast.Ident)(0xc420142580), Rbrack:27314})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142560), Lbrack:27306, Index:(*ast.Ident)(0xc420142580), Rbrack:27314})[0mChannelPtr__setPANPOT(regs->chip->channels[channel], v);
	} else if (__tag == ymf->CHPAN) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201426e0), Lbrack:27366, Index:(*ast.Ident)(0xc420142700), Rbrack:27374})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201426e0), Lbrack:27366, Index:(*ast.Ident)(0xc420142700), Rbrack:27374})[0mChannelPtr__setCHPAN(regs->chip->channels[channel], v);
	} else if (__tag == ymf->VOLUME) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142860), Lbrack:27426, Index:(*ast.Ident)(0xc420142880), Rbrack:27434})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142860), Lbrack:27426, Index:(*ast.Ident)(0xc420142880), Rbrack:27434})[0mChannelPtr__setVOLUME(regs->chip->channels[channel], v);
	} else if (__tag == ymf->EXPRESSION) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201429e0), Lbrack:27491, Index:(*ast.Ident)(0xc420142a00), Rbrack:27499})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201429e0), Lbrack:27491, Index:(*ast.Ident)(0xc420142a00), Rbrack:27499})[0mChannelPtr__setEXPRESSION(regs->chip->channels[channel], v);
	} else if (__tag == ymf->VELOCITY) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142b60), Lbrack:27558, Index:(*ast.Ident)(0xc420142b80), Rbrack:27566})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142b60), Lbrack:27558, Index:(*ast.Ident)(0xc420142b80), Rbrack:27566})[0mChannelPtr__setVELOCITY(regs->chip->channels[channel], v);
	} else if (__tag == ymf->BO) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142ce0), Lbrack:27617, Index:(*ast.Ident)(0xc420142d00), Rbrack:27625})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142ce0), Lbrack:27617, Index:(*ast.Ident)(0xc420142d00), Rbrack:27625})[0mChannelPtr__setBO(regs->chip->channels[channel], v);
	} else if (__tag == ymf->RESET) {
		if (v != 0) {
			[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142ea0), Lbrack:27688, Index:(*ast.Ident)(0xc420142ec0), Rbrack:27696})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420142ea0), Lbrack:27688, Index:(*ast.Ident)(0xc420142ec0), Rbrack:27696})[0mChannelPtr__resetAll(regs->chip->channels[channel]);
		}
	}
}

}

#include "register.h"

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
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:24611, Call:(*ast.CallExpr)(0xc420138f00)})[0m
	auto __tag = offset;
	if (__tag == ymf->EAM) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013ee60), Lbrack:24713, Index:(*ast.Ident)(0xc42013ee80), Rbrack:24727})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013ee60), Lbrack:24713, Index:(*ast.Ident)(0xc42013ee80), Rbrack:24727})[0moperatorPtr__setEAM(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->EVB) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f040), Lbrack:24793, Index:(*ast.Ident)(0xc42013f060), Rbrack:24807})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f040), Lbrack:24793, Index:(*ast.Ident)(0xc42013f060), Rbrack:24807})[0moperatorPtr__setEVB(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DAM) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f240), Lbrack:24873, Index:(*ast.Ident)(0xc42013f260), Rbrack:24887})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f240), Lbrack:24873, Index:(*ast.Ident)(0xc42013f260), Rbrack:24887})[0moperatorPtr__setDAM(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DVB) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f420), Lbrack:24953, Index:(*ast.Ident)(0xc42013f440), Rbrack:24967})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f420), Lbrack:24953, Index:(*ast.Ident)(0xc42013f440), Rbrack:24967})[0moperatorPtr__setDVB(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DT) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f600), Lbrack:25032, Index:(*ast.Ident)(0xc42013f620), Rbrack:25046})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f600), Lbrack:25032, Index:(*ast.Ident)(0xc42013f620), Rbrack:25046})[0moperatorPtr__setDT(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->KSR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f7e0), Lbrack:25111, Index:(*ast.Ident)(0xc42013f800), Rbrack:25125})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f7e0), Lbrack:25111, Index:(*ast.Ident)(0xc42013f800), Rbrack:25125})[0moperatorPtr__setKSR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->MULT) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f9c0), Lbrack:25192, Index:(*ast.Ident)(0xc42013f9e0), Rbrack:25206})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013f9c0), Lbrack:25192, Index:(*ast.Ident)(0xc42013f9e0), Rbrack:25206})[0moperatorPtr__setMULT(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->KSL) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013fba0), Lbrack:25273, Index:(*ast.Ident)(0xc42013fbc0), Rbrack:25287})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013fba0), Lbrack:25273, Index:(*ast.Ident)(0xc42013fbc0), Rbrack:25287})[0moperatorPtr__setKSL(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->TL) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013fd80), Lbrack:25352, Index:(*ast.Ident)(0xc42013fda0), Rbrack:25366})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013fd80), Lbrack:25352, Index:(*ast.Ident)(0xc42013fda0), Rbrack:25366})[0moperatorPtr__setTL(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->AR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013ff60), Lbrack:25430, Index:(*ast.Ident)(0xc42013ff80), Rbrack:25444})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42013ff60), Lbrack:25430, Index:(*ast.Ident)(0xc42013ff80), Rbrack:25444})[0moperatorPtr__setAR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->DR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420148140), Lbrack:25508, Index:(*ast.Ident)(0xc420148160), Rbrack:25522})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420148140), Lbrack:25508, Index:(*ast.Ident)(0xc420148160), Rbrack:25522})[0moperatorPtr__setDR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->SL) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420148320), Lbrack:25586, Index:(*ast.Ident)(0xc420148340), Rbrack:25600})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420148320), Lbrack:25586, Index:(*ast.Ident)(0xc420148340), Rbrack:25600})[0moperatorPtr__setSL(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->SR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420148500), Lbrack:25664, Index:(*ast.Ident)(0xc420148520), Rbrack:25678})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420148500), Lbrack:25664, Index:(*ast.Ident)(0xc420148520), Rbrack:25678})[0moperatorPtr__setSR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->RR) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201486e0), Lbrack:25742, Index:(*ast.Ident)(0xc420148700), Rbrack:25756})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201486e0), Lbrack:25742, Index:(*ast.Ident)(0xc420148700), Rbrack:25756})[0moperatorPtr__setRR(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->XOF) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201488c0), Lbrack:25821, Index:(*ast.Ident)(0xc4201488e0), Rbrack:25835})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc4201488c0), Lbrack:25821, Index:(*ast.Ident)(0xc4201488e0), Rbrack:25835})[0moperatorPtr__setXOF(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->WS) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420148aa0), Lbrack:25900, Index:(*ast.Ident)(0xc420148ac0), Rbrack:25914})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420148aa0), Lbrack:25900, Index:(*ast.Ident)(0xc420148ac0), Rbrack:25914})[0moperatorPtr__setWS(regs->chip->channels[channel]->operators[operatorIndex], v);
	} else if (__tag == ymf->FB) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420148c80), Lbrack:25978, Index:(*ast.Ident)(0xc420148ca0), Rbrack:25992})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420148c80), Lbrack:25978, Index:(*ast.Ident)(0xc420148ca0), Rbrack:25992})[0moperatorPtr__setFB(regs->chip->channels[channel]->operators[operatorIndex], v);
	}
}

// WriteTL は、TLレジスタに値を書き込みます。
void RegistersPtr__WriteTL(Registers *regs, int channel, int operatorIndex, int tlCarrier, int tlModulator) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:26185, Call:(*ast.CallExpr)(0xc420139940)})[0m
	if (regs->chip->channels[channel]->operators[operatorIndex]->isModulator) {
		RegistersPtr__WriteOperator(regs, channel, operatorIndex, ymf->TL, tlModulator);
	} else {
		RegistersPtr__WriteOperator(regs, channel, operatorIndex, ymf->TL, tlCarrier);
	}
}

// DebugSetMIDIChannel は、チャンネルを使用しているMIDIチャンネル番号をデバッグ用にセットします。
void RegistersPtr__DebugSetMIDIChannel(Registers *regs, int channel, int midiChannel) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:26662, Call:(*ast.CallExpr)(0xc420139c00)})[0m
	regs->chip->channels[channel]->midiChannelID = midiChannel;
}

// WriteChannel は、チャンネルレジスタに値を書き込みます。
void RegistersPtr__WriteChannel(Registers *regs, int channel, gopkg_in::but80::fmfm_core_v1::ymf::ChRegister offset, int v) {
	sync::Mutex__Lock(regs->chip->Mutex);
	[41mstmt[0m[31m<*ast.DeferStmt>(&ast.DeferStmt{Defer:26939, Call:(*ast.CallExpr)(0xc420139e00)})[0m
	auto __tag = offset;
	if (__tag == ymf->KON) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420149de0), Lbrack:27022, Index:(*ast.Ident)(0xc420149e00), Rbrack:27030})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420149de0), Lbrack:27022, Index:(*ast.Ident)(0xc420149e00), Rbrack:27030})[0mChannelPtr__setKON(regs->chip->channels[channel], v);
	} else if (__tag == ymf->BLOCK) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420149f60), Lbrack:27079, Index:(*ast.Ident)(0xc420149f80), Rbrack:27087})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc420149f60), Lbrack:27079, Index:(*ast.Ident)(0xc420149f80), Rbrack:27087})[0mChannelPtr__setBLOCK(regs->chip->channels[channel], v);
	} else if (__tag == ymf->FNUM) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a100), Lbrack:27137, Index:(*ast.Ident)(0xc42014a120), Rbrack:27145})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a100), Lbrack:27137, Index:(*ast.Ident)(0xc42014a120), Rbrack:27145})[0mChannelPtr__setFNUM(regs->chip->channels[channel], v);
	} else if (__tag == ymf->ALG) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a280), Lbrack:27193, Index:(*ast.Ident)(0xc42014a2a0), Rbrack:27201})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a280), Lbrack:27193, Index:(*ast.Ident)(0xc42014a2a0), Rbrack:27201})[0mChannelPtr__setALG(regs->chip->channels[channel], v);
	} else if (__tag == ymf->LFO) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a400), Lbrack:27248, Index:(*ast.Ident)(0xc42014a420), Rbrack:27256})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a400), Lbrack:27248, Index:(*ast.Ident)(0xc42014a420), Rbrack:27256})[0mChannelPtr__setLFO(regs->chip->channels[channel], v);
	} else if (__tag == ymf->PANPOT) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a580), Lbrack:27306, Index:(*ast.Ident)(0xc42014a5a0), Rbrack:27314})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a580), Lbrack:27306, Index:(*ast.Ident)(0xc42014a5a0), Rbrack:27314})[0mChannelPtr__setPANPOT(regs->chip->channels[channel], v);
	} else if (__tag == ymf->CHPAN) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a700), Lbrack:27366, Index:(*ast.Ident)(0xc42014a720), Rbrack:27374})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a700), Lbrack:27366, Index:(*ast.Ident)(0xc42014a720), Rbrack:27374})[0mChannelPtr__setCHPAN(regs->chip->channels[channel], v);
	} else if (__tag == ymf->VOLUME) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a880), Lbrack:27426, Index:(*ast.Ident)(0xc42014a8a0), Rbrack:27434})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014a880), Lbrack:27426, Index:(*ast.Ident)(0xc42014a8a0), Rbrack:27434})[0mChannelPtr__setVOLUME(regs->chip->channels[channel], v);
	} else if (__tag == ymf->EXPRESSION) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014aa00), Lbrack:27491, Index:(*ast.Ident)(0xc42014aa20), Rbrack:27499})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014aa00), Lbrack:27491, Index:(*ast.Ident)(0xc42014aa20), Rbrack:27499})[0mChannelPtr__setEXPRESSION(regs->chip->channels[channel], v);
	} else if (__tag == ymf->VELOCITY) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014ab80), Lbrack:27558, Index:(*ast.Ident)(0xc42014aba0), Rbrack:27566})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014ab80), Lbrack:27558, Index:(*ast.Ident)(0xc42014aba0), Rbrack:27566})[0mChannelPtr__setVELOCITY(regs->chip->channels[channel], v);
	} else if (__tag == ymf->BO) {
		[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014ad00), Lbrack:27617, Index:(*ast.Ident)(0xc42014ad20), Rbrack:27625})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014ad00), Lbrack:27617, Index:(*ast.Ident)(0xc42014ad20), Rbrack:27625})[0mChannelPtr__setBO(regs->chip->channels[channel], v);
	} else if (__tag == ymf->RESET) {
		if (v != 0) {
			[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014aec0), Lbrack:27688, Index:(*ast.Ident)(0xc42014aee0), Rbrack:27696})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.SelectorExpr)(0xc42014aec0), Lbrack:27688, Index:(*ast.Ident)(0xc42014aee0), Rbrack:27696})[0mChannelPtr__resetAll(regs->chip->channels[channel]);
		}
	}
}

}

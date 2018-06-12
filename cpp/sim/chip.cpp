#pragma once
#include "./chip.h"

namespace sim {



// NewChip は、新しい Chip を作成します。
Chip *NewChip(float64 sampleRate, float64 totalLevel, int dumpMIDIChannel) {
	auto chip = __ptr((const Chip){
		sampleRate: sampleRate,
		totalLevel: totalLevel,
		dumpMIDIChannel: dumpMIDIChannel,
		channels: make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:11070, Len:ast.Expr(nil), Elt:(*ast.StarExpr)(0xc4200f9780)})[0m, ymfdata->ChannelCount),
		currentOutput: make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:11129, Len:ast.Expr(nil), Elt:(*ast.Ident)(0xc4200f9860)})[0m, 2),
	});
	ChipPtr__initChannels(chip);
	return chip;
}

auto debugDumpCount = 0;

// Next は、次のサンプルを生成し、その左右それぞれの振幅を返します。
MULTIRESULT ChipPtr__Next(Chip *chip) {
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
		debugDumpCount++
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
					return [41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4200fe740), Lbrack:11960, Index:(*ast.Ident)(0xc4200fe760), Rbrack:11962})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4200fe740), Lbrack:11960, Index:(*ast.Ident)(0xc4200fe760), Rbrack:11962})[0mChannelPtr__currentLevel(toDump[i]) < [41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4200fe7c0), Lbrack:11987, Index:(*ast.Ident)(0xc4200fe7e0), Rbrack:11989})[0m[41mobjectOf[0m[31m<*ast.IndexExpr>(&ast.IndexExpr{X:(*ast.Ident)(0xc4200fe7c0), Lbrack:11987, Index:(*ast.Ident)(0xc4200fe7e0), Rbrack:11989})[0mChannelPtr__currentLevel(toDump[j]);
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
	chip->channels = make([41mexpr[0m[31m<*ast.ArrayType>(&ast.ArrayType{Lbrack:12221, Len:ast.Expr(nil), Elt:(*ast.StarExpr)(0xc4200feca0)})[0m, ymfdata->ChannelCount);
	for (int i = 0; i < (int)chip->channels.size(); i++) {
		chip->channels[i] = newChannel(i, chip);
	}
}

}

package ymf

// OpRegister は、オペレータパラメータを保持するレジスタの種類を表す型です。
type OpRegister int

const (
	// EAM は、EAM レジスタです。
	EAM OpRegister = iota
	// EVB は、EVB レジスタです。
	EVB
	// DAM は、DAM レジスタです。
	DAM
	// DVB は、DVB レジスタです。
	DVB
	// DT は、DT レジスタです。
	DT
	// KSL は、KSL レジスタです。
	KSL
	// KSR は、KSR レジスタです。
	KSR
	// WS は、WS レジスタです。
	WS
	// MULT は、MULT レジスタです。
	MULT
	// FB は、FB レジスタです。
	FB
	// AR は、AR レジスタです。
	AR
	// DR は、DR レジスタです。
	DR
	// SL は、SL レジスタです。
	SL
	// SR は、SR レジスタです。
	SR
	// RR は、RR レジスタです。
	RR
	// TL は、TL レジスタです。
	TL
	// XOF は、XOF レジスタです。
	XOF
)

// ChRegister は、チャンネルパラメータを保持するレジスタの種類を表す型です。
type ChRegister int

const (
	// KON は、KON レジスタです。
	KON ChRegister = iota
	// BLOCK は、BLOCK レジスタです。
	BLOCK
	// FNUM は、FNUM レジスタです。
	FNUM
	// ALG は、ALG レジスタです。
	ALG
	// LFO は、LFO レジスタです。
	LFO
	// PANPOT は、PANPOT レジスタです。
	PANPOT
	// CHPAN は、CHPAN レジスタです。
	CHPAN
	// VOLUME は、VOLUME レジスタです。
	VOLUME
	// EXPRESSION は、EXPRESSION レジスタです。
	EXPRESSION
	// VELOCITY は、VELOCITY レジスタです。
	VELOCITY
	// BO は、BO レジスタです。
	BO
)

// Registers は、音源チップのレジスタを抽象化したインタフェースです。
type Registers interface {
	// WriteOperator は、オペレータレジスタに値を書き込みます。
	WriteOperator(channel, operatorIndex int, offset OpRegister, v int)
	// WriteTL は、TLレジスタに値を書き込みます。
	WriteTL(channel, operatorIndex int, tlCarrier, tlModulator int)
	// WriteChannel は、チャンネルレジスタに値を書き込みます。
	WriteChannel(channel int, offset ChRegister, v int)
	// DebugSetMIDIChannel は、チャンネルを使用しているMIDIチャンネル番号をデバッグ用にセットします。
	DebugSetMIDIChannel(channel, midiChannel int)
}

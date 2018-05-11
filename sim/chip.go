package sim

import (
	"math"

	"github.com/but80/fmfm/ymf/ymfdata"
)

// Chip は、FM音源チップ全体を表す型です。
type Chip struct {
	// sampleRate は、出力波形の目標サンプルレートです。
	sampleRate float64
	// totalLevel は、出力のトータルな音量[dB]です。
	totalLevel float64
	// channels は、このチップが備える全チャンネルです。
	channels []*Channel

	currentOutput []float64
}

// NewChip は、新しい Chip を作成します。
func NewChip(sampleRate, totalLevel float64) *Chip {
	chip := &Chip{
		sampleRate:    sampleRate,
		totalLevel:    totalLevel,
		channels:      make([]*Channel, ymfdata.ChannelCount),
		currentOutput: make([]float64, 2),
	}
	chip.initChannels()
	return chip
}

// Next は、次のサンプルを生成し、その左右それぞれの振幅を返します。
func (chip *Chip) Next() (float64, float64) {
	var l, r float64
	for _, channel := range chip.channels {
		cl, cr := channel.getChannelOutput()
		l += cl
		r += cr
	}
	v := math.Pow(10, chip.totalLevel/20)
	return l * v, r * v
}

func (chip *Chip) initChannels() {
	chip.channels = make([]*Channel, ymfdata.ChannelCount)
	for i := range chip.channels {
		chip.channels[i] = newChannel4op(i, chip)
	}
}

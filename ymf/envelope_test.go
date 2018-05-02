package ymf

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/but80/fmfm/ymf/ymfdata"
)

func TestEnvelopeGenerator(t *testing.T) {
	threshDB := -30.0
	thresh := math.Pow(10.0, threshDB/20.0)
	gen := newEnvelopeGenerator()
	eam := 0
	dam := 0
	ar := 15
	dr := 15
	sl := 0
	sr := 0
	rr := 4
	ksl := 0
	result := [][]float64{}
	for ksr := 0; ksr < 2; ksr++ {
		r := []float64{}
		for ksn := 0; ksn < 16; ksn++ {
			fnum := (ksn & 1) * 256
			block := ksn >> 1
			gen.setKeyScalingLevel(fnum, block, ksl)
			gen.setActualAttackRate(ar, ksr, ksn)
			gen.setActualDR(dr, ksr, ksn)
			gen.setActualSustainLevel(sl)
			gen.setActualSR(sr, ksr, ksn)
			gen.setActualRR(rr, ksr, ksn)
			gen.keyOn()
			n := int(.1 * ymfdata.SampleRate)
			i := 0
			for ; i < 60*ymfdata.SampleRate; i++ {
				if i == n {
					gen.keyOff()
				}
				v := gen.getEnvelope(eam, dam, 0)
				if n < i && v <= thresh {
					break
				}
			}
			i -= n
			secPerDb := float64(i) / float64(ymfdata.SampleRate) / (.0 - threshDB)
			dbPerSec := 1.0 / secPerDb
			r = append(r, math.Floor(dbPerSec))
		}
		result = append(result, r)
	}

	assert.Equal(t, [][]float64{
		{17, 17, 17, 17, 17, 22, 22, 22, 22, 26, 26, 26, 26, 31, 31, 31},
		{17, 22, 22, 31, 31, 44, 44, 62, 62, 89, 89, 125, 125, 179, 179, 250},
	}, result)
}

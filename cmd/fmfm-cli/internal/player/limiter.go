package player

import (
	"math"
)

// Insertion は、インサーションエフェクトを抽象化したインタフェースです。
type Insertion interface {
	// Next は、次のサンプルを生成し、その左右それぞれの振幅を返します。
	Next(l, r float64) (float64, float64)
}

// Limiter は、インサーションエフェクト「リミッター」です。
type Limiter struct {
	sampleRate  float64
	attack      float64
	attackInv   float64
	release     float64
	threshold   float64
	thresholdDB float64
	attenuation float64
	buffer      [][2]float64
	bufferPos   int
}

// NewLimiter は、新しい Limiter を作成します。
func NewLimiter(sampleRate float64) *Limiter {
	lim := &Limiter{
		sampleRate: sampleRate,
	}
	return lim.SetThreshold(-3.0).SetLookAhead(.005).SetAttack(.005).SetRelease(.02)
}

// SetThreshold は、スレッショルドレベル [dB] を設定します。
func (lim *Limiter) SetThreshold(v float64) *Limiter {
	lim.threshold = math.Pow(10, v/20.0)
	lim.thresholdDB = v
	return lim
}

// SetLookAhead は、先読み時間 [秒] を設定します。
func (lim *Limiter) SetLookAhead(v float64) *Limiter {
	n := int(math.Ceil(lim.sampleRate * v))
	lim.buffer = make([][2]float64, n)
	lim.bufferPos = 0
	return lim
}

// SetAttack は、アタックタイムを設定します。
func (lim *Limiter) SetAttack(sec float64) *Limiter {
	lim.attack = lim.timeToMultiplier(sec)
	lim.attackInv = 1.0 - lim.attack
	return lim
}

// SetRelease は、リリースタイムを設定します。
func (lim *Limiter) SetRelease(sec float64) *Limiter {
	lim.release = lim.timeToMultiplier(sec)
	return lim
}

func (lim *Limiter) timeToMultiplier(sec float64) float64 {
	n := sec * lim.sampleRate
	return math.Pow(0.1/0.9, 1/n) // result ^ n = 0.1/0.9
	// return math.Exp(-0.9542 / n)
}

// Next は、次のサンプルを生成し、その左右それぞれの振幅を返します。
func (lim *Limiter) Next(l, r float64) (float64, float64) {
	lim.buffer[lim.bufferPos][0] = l
	lim.buffer[lim.bufferPos][1] = r
	lim.bufferPos = (lim.bufferPos + 1) % len(lim.buffer)
	v := math.Max(math.Abs(l), math.Abs(r))
	l = lim.buffer[lim.bufferPos][0]
	r = lim.buffer[lim.bufferPos][1]
	if .0 <= lim.thresholdDB {
		return l, r
	}
	if lim.threshold <= v {
		db := 20.0 * math.Log10(v)
		target := lim.thresholdDB - db
		lim.attenuation = lim.attack*lim.attenuation + lim.attackInv*target
	} else {
		lim.attenuation *= lim.release
	}
	a := math.Pow(10.0, lim.attenuation/20.0)
	return l * a, r * a
}

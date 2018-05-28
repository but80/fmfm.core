package player

import (
	"math"
)

type Insertion interface {
	Next(l, r float64) (float64, float64)
}

type Limiter struct {
	sampleRate  float64
	attack      float64
	attackInv   float64
	release     float64
	threshold   float64
	attenuation float64
	buffer      [][2]float64
	bufferPos   int
}

func NewLimiter(sampleRate float64) *Limiter {
	lim := &Limiter{
		sampleRate: sampleRate,
		threshold:  -3.0,
	}
	return lim.SetLookAhead(.005).SetAttack(.005).SetRelease(.02)
}

func (lim *Limiter) SetLookAhead(v float64) *Limiter {
	n := int(math.Ceil(lim.sampleRate * v))
	lim.buffer = make([][2]float64, n)
	lim.bufferPos = 0
	return lim
}

func (lim *Limiter) SetThreshold(v float64) *Limiter {
	lim.threshold = v
	return lim
}

func (lim *Limiter) SetAttack(sec float64) *Limiter {
	lim.attack = lim.timeToMultiplier(sec)
	lim.attackInv = 1.0 - lim.attack
	return lim
}

func (lim *Limiter) SetRelease(sec float64) *Limiter {
	lim.release = lim.timeToMultiplier(sec)
	return lim
}

func (lim *Limiter) timeToMultiplier(sec float64) float64 {
	n := sec * lim.sampleRate
	return math.Pow(0.1/0.9, 1/n) // result ^ n = 0.1/0.9
	// return math.Exp(-0.9542 / n)
}

func (lim *Limiter) Next(l, r float64) (float64, float64) {
	lim.buffer[lim.bufferPos][0] = l
	lim.buffer[lim.bufferPos][1] = r
	lim.bufferPos = (lim.bufferPos + 1) % len(lim.buffer)
	v := math.Max(l, r)
	l = lim.buffer[lim.bufferPos][0]
	r = lim.buffer[lim.bufferPos][1]
	if .0 <= lim.threshold {
		return l, r
	}
	db := 20.0 * math.Log10(v)
	if lim.threshold <= db {
		target := lim.threshold - db
		lim.attenuation = lim.attack*lim.attenuation + lim.attackInv*target
	} else {
		lim.attenuation *= lim.release
	}
	a := math.Pow(10.0, lim.attenuation/20.0)
	return l * a, r * a
}

import phaseGenerator from './phase-generator'
import { YMF_DATA } from './data'

describe('phaseGenerator', () => {

  it('generates phase detuned by DT', () => {
    function round (v: number): number {
      return Math.round(v * 100) / 100
    }

    function measure (fnum: number, block: number, mult: number, dt: number): number {
      const gen = new phaseGenerator()
      gen.setFrequency(fnum, block, mult, dt)
      gen.keyOn()
      let dPhase = gen.getPhase(0, 0, 0)
      return dPhase * YMF_DATA.sampleRate
    }

    for (let mult = 1; mult <= 8; mult++) {
      for (let dt = 1; dt < 4; dt++) {
        const r = []
        for (let ksn = 0; ksn < 16; ksn++) {
          const fnum = (ksn & 1) * 512 + 256
          const block = ksn >> 1
          const freq0 = measure(fnum, block, mult, 0)
          const freq = measure(fnum, block, mult, dt)
          const dFreq = freq - freq0
          r.push(round(dFreq))
        }
        expect(r).toEqual(YMF_DATA.dtCoef[dt].map(v => round(v * mult)))
      }
    }
  })

})

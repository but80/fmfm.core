import EnvelopeGenerator from './envelope-generator'
import { YMF_DATA } from './data'

describe('EnvelopeGenerator', () => {

  it('generates envelope', () => {
    const threshDB = -30
    const thresh = Math.pow(10, threshDB / 20)
    const gen = new EnvelopeGenerator()
    const eam = 0
    const dam = 0
    const ar = 15
    const dr = 15
    const sl = 0
    const sr = 0
    const rr = 4
    const ksl = 0
    const result = []
    for (let ksr = 0; ksr < 2; ksr++) {
      const r = []
      for (let ksn = 0; ksn < 16; ksn++) {
        const fnum = (ksn & 1) * 256
        const block = ksn >> 1
        gen.setAtennuation(fnum, block, ksl)
        gen.setActualAttackRate(ar, ksr, ksn)
        gen.setActualDR(dr, ksr, ksn)
        gen.setActualSustainLevel(sl)
        gen.setActualSR(sr, ksr, ksn)
        gen.setActualRR(rr, ksr, ksn)
        gen.keyOn()
        let n = 10
        let i = 0
        for (; i < 60 * YMF_DATA.sampleRate; i++) {
          if (i === n) gen.keyOff()
          let e = gen.getEnvelope(eam, dam, 0)
          const v = Math.pow(10, e / 20)
          if (v <= thresh) break
        }
        i -= n
        const secPerDb = i / YMF_DATA.sampleRate / (0 - threshDB)
        const dbPerSec = 1 / secPerDb
        r.push(Math.round(dbPerSec))
      }
      result.push(r)
    }
    expect(result).toEqual([
      [ 17, 17, 17, 17, 17, 22, 22, 22, 22, 26, 26, 26, 26, 31, 31, 31 ],
      [ 17, 22, 22, 31, 31, 44, 44, 62, 62, 89, 89, 125, 125, 179, 179, 250 ]
    ])
  })

})

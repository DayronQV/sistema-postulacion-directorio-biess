import type { Inhabilidades } from '@/services/postulacion'
import {
  PREGUNTAS_INHABILIDADES,
  inhabilidadesCompletas,
  preguntasQueInhabilitan,
} from './inhabilidades'

// Respuestas "buenas": Sí en la 1 y No en el resto.
function buenas(): Inhabilidades {
  const r: Inhabilidades = {}
  for (let n = 1; n <= 13; n++) r[String(n)] = n === 1
  return r
}

describe('inhabilidades', () => {
  it('tiene las 13 preguntas', () => {
    expect(PREGUNTAS_INHABILIDADES).toHaveLength(13)
  })

  it('con las respuestas esperadas nada inhabilita', () => {
    expect(preguntasQueInhabilitan(buenas())).toEqual([])
  })

  it('un No en la 1 inhabilita', () => {
    expect(preguntasQueInhabilitan({ ...buenas(), '1': false })).toEqual([1])
  })

  it('un Sí en cualquiera de la 2 a la 13 inhabilita', () => {
    expect(preguntasQueInhabilitan({ ...buenas(), '5': true, '13': true })).toEqual([5, 13])
  })

  it('no avisa por las preguntas sin responder', () => {
    expect(preguntasQueInhabilitan({})).toEqual([])
    expect(preguntasQueInhabilitan(undefined)).toEqual([])
  })

  it('está completa solo con las 13 respondidas', () => {
    expect(inhabilidadesCompletas(buenas())).toBe(true)
    const incompleta = buenas()
    delete incompleta['9']
    expect(inhabilidadesCompletas(incompleta)).toBe(false)
    expect(inhabilidadesCompletas(undefined)).toBe(false)
  })

  it('con la 7 en Sí exige institución y motivo', () => {
    const sinDetalle = { ...buenas(), '7': true }
    expect(inhabilidadesCompletas(sinDetalle)).toBe(false)
    const conDetalle = {
      ...sinDetalle,
      removido_detalle: { institucion: 'SB', cargo: '', fecha_fin: '', motivo: 'Renuncia' },
    }
    expect(inhabilidadesCompletas(conDetalle)).toBe(true)
  })
})

import type { Postulacion } from '@/services/postulacion'
import {
  esNacionalidadEcuatoriana,
  estadoCivilConPareja,
  formatearFecha,
  motivoDeBloqueo,
} from './informacionPersonal'

const con = (estado: Postulacion['estado'], nacionalidad: string) =>
  ({ estado, informacion_personal: { nacionalidad } }) as Pick<
    Postulacion,
    'estado' | 'informacion_personal'
  >

describe('esNacionalidadEcuatoriana', () => {
  it.each([
    'Ecuatoriana',
    'ecuatoriana',
    'ECUATORIANA',
    '  Ecuatoriana  ',
    'Ecuatoriano',
    'Écuatoriana',
  ])('acepta %j', (valor) => {
    expect(esNacionalidadEcuatoriana(valor)).toBe(true)
  })

  it.each(['Colombiana', '', 'Ecuador', 'No ecuatoriana'])('rechaza %j', (valor) => {
    expect(esNacionalidadEcuatoriana(valor)).toBe(false)
  })
})

describe('estadoCivilConPareja', () => {
  it.each(['Casado/a', 'casado/a', 'Unión de hecho', 'union de hecho'])(
    '%j lleva pareja',
    (valor) => {
      expect(estadoCivilConPareja(valor)).toBe(true)
    },
  )

  it.each(['Soltero/a', 'Divorciado/a', 'Viudo/a', ''])('%j no lleva pareja', (valor) => {
    expect(estadoCivilConPareja(valor)).toBe(false)
  })
})

describe('formatearFecha', () => {
  it('convierte AAAA-MM-DD a DD/MM/AAAA sin cambiar el día', () => {
    expect(formatearFecha('1978-11-02')).toBe('02/11/1978')
    expect(formatearFecha('1985-03-14')).toBe('14/03/1985')
  })

  it('deja intacto lo que no es una fecha ISO', () => {
    expect(formatearFecha('')).toBe('')
    expect(formatearFecha('ayer')).toBe('ayer')
  })
})

describe('motivoDeBloqueo', () => {
  it('no bloquea a una persona ecuatoriana con la postulación en borrador', () => {
    expect(motivoDeBloqueo(con('borrador', 'Ecuatoriana'))).toBeNull()
    expect(motivoDeBloqueo(con('enviada', 'ECUATORIANA'))).toBeNull()
  })

  it('bloquea por nacionalidad aunque la postulación siga en borrador', () => {
    expect(motivoDeBloqueo(con('borrador', 'Colombiana'))).toMatch(/nacionalidad ecuatoriana/)
    expect(motivoDeBloqueo(con('cerrada', 'Colombiana'))).toMatch(/«Colombiana»/)
  })

  it('bloquea una postulación cerrada de una persona ecuatoriana con un motivo genérico', () => {
    expect(motivoDeBloqueo(con('cerrada', 'Ecuatoriana'))).toMatch(/cerrada/)
  })
})

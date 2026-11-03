import type { Fase } from '@/services/adminFases'
import { minutosDeVentana, ventanaMasCortaQueElExamen } from './ventana'

const fase = (inicio: string | null, fin: string | null, nombre = 'Evaluación'): Fase => ({
  id: 3,
  nombre,
  orden: 3,
  publicada: true,
  habilita_formulario: false,
  fecha_inicio: inicio,
  fecha_fin: fin,
})

const DE_10_A_11 = fase('2026-11-10T10:00:00Z', '2026-11-10T11:00:00Z')

describe('minutosDeVentana', () => {
  it('calcula los minutos entre la fecha de inicio y la de fin', () => {
    expect(minutosDeVentana(DE_10_A_11)).toBe(60)
  })

  it('sin alguna de las dos fechas no hay ventana que medir', () => {
    expect(minutosDeVentana(fase(null, '2026-11-10T11:00:00Z'))).toBeNull()
    expect(minutosDeVentana(fase('2026-11-10T10:00:00Z', null))).toBeNull()
    expect(minutosDeVentana(undefined)).toBeNull()
  })

  it('ignora fechas inválidas o con el fin antes del inicio', () => {
    expect(minutosDeVentana(fase('no-es-fecha', '2026-11-10T11:00:00Z'))).toBeNull()
    expect(minutosDeVentana(fase('2026-11-10T11:00:00Z', '2026-11-10T10:00:00Z'))).toBeNull()
  })
})

describe('ventanaMasCortaQueElExamen', () => {
  it('devuelve los minutos de la ventana si es más corta que la duración', () => {
    expect(ventanaMasCortaQueElExamen([DE_10_A_11], 90)).toBe(60)
  })

  it('no avisa si la ventana alcanza (igual o mayor)', () => {
    expect(ventanaMasCortaQueElExamen([DE_10_A_11], 60)).toBeNull()
    expect(ventanaMasCortaQueElExamen([DE_10_A_11], 30)).toBeNull()
  })

  it('solo mira la fase Evaluación', () => {
    const otra = fase('2026-11-10T10:00:00Z', '2026-11-10T10:10:00Z', 'Méritos')
    expect(ventanaMasCortaQueElExamen([otra], 90)).toBeNull()
  })

  it('sin fases o sin duración válida no avisa', () => {
    expect(ventanaMasCortaQueElExamen(null, 90)).toBeNull()
    expect(ventanaMasCortaQueElExamen([DE_10_A_11], null)).toBeNull()
  })
})

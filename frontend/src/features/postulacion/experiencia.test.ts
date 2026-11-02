import { calcularDuracion, sumarDuraciones, textoDuracion } from './experiencia'

describe('calculadora de experiencia', () => {
  it.each([
    ['2024-06-15', '2024-06-15', { anios: 0, meses: 0, dias: 0 }],
    ['2024-01-31', '2024-02-29', { anios: 0, meses: 0, dias: 29 }],
    ['2020-02-29', '2021-02-28', { anios: 0, meses: 11, dias: 30 }],
    ['2018-01-10', '2020-03-15', { anios: 2, meses: 2, dias: 5 }],
  ])('calcula %s a %s', (inicio, fin, esperado) => {
    expect(calcularDuracion(inicio, fin)).toEqual(esperado)
  })

  it('normaliza y presenta el total', () => {
    const total = sumarDuraciones([
      { anios: 1, meses: 11, dias: 20 },
      { anios: 0, meses: 2, dias: 15 },
    ])
    expect(total).toEqual({ anios: 2, meses: 2, dias: 5 })
    expect(textoDuracion(total)).toBe('2 años, 2 meses y 5 días')
  })
})

import { analizarCsv, erroresDeCarga, leerCsv, validarPregunta } from './preguntas'

const valida = { enunciado: '¿Capital?', opciones: ['Quito', 'Cuenca'], respuesta_correcta: 0 }

describe('validarPregunta', () => {
  it('acepta una pregunta válida', () => {
    expect(validarPregunta(valida)).toEqual({})
  })

  it('exige enunciado y respuesta correcta', () => {
    const errores = validarPregunta({ ...valida, enunciado: '  ', respuesta_correcta: null })
    expect(errores.enunciado).toBeDefined()
    expect(errores.respuesta_correcta).toBeDefined()
  })

  it('exige de 2 a 8 opciones', () => {
    expect(validarPregunta({ ...valida, opciones: ['Sola'] }).opciones).toBeDefined()
    const nueve = Array.from({ length: 9 }, (_, i) => `Opción ${i}`)
    expect(validarPregunta({ ...valida, opciones: nueve }).opciones).toBeDefined()
  })

  it('rechaza opciones vacías o repetidas, indicando cuál', () => {
    const errores = validarPregunta({ ...valida, opciones: ['Quito', ' ', 'Quito'] })
    expect(errores['opciones[1]']).toBeDefined()
    expect(errores['opciones[2]']).toBeDefined()
    expect(errores['opciones[0]']).toBeUndefined()
  })

  it('la respuesta correcta debe estar dentro de las opciones', () => {
    expect(validarPregunta({ ...valida, respuesta_correcta: 2 }).respuesta_correcta).toBeDefined()
    expect(validarPregunta({ ...valida, respuesta_correcta: -1 }).respuesta_correcta).toBeDefined()
  })

  it('limita el largo del enunciado y de cada opción', () => {
    expect(validarPregunta({ ...valida, enunciado: 'a'.repeat(1001) }).enunciado).toBeDefined()
    expect(
      validarPregunta({ ...valida, opciones: ['a'.repeat(501), 'b'] })['opciones[0]'],
    ).toBeDefined()
  })
})

describe('leerCsv', () => {
  it('lee comillas, comillas escapadas y saltos de línea dentro de una celda', () => {
    const filas = leerCsv('a,b\n"uno, dos","dice ""hola""\nlínea 2"\n')
    expect(filas).toEqual([
      ['a', 'b'],
      ['uno, dos', 'dice "hola"\nlínea 2'],
    ])
  })

  it('acepta punto y coma (Excel en español), BOM y fin de línea de Windows', () => {
    expect(leerCsv('﻿a;b\r\n1;2\r\n')).toEqual([
      ['a', 'b'],
      ['1', '2'],
    ])
  })

  it('ignora las líneas vacías', () => {
    expect(leerCsv('a,b\n\n1,2\n\n')).toHaveLength(2)
  })
})

describe('analizarCsv', () => {
  const cabecera = 'enunciado,opciones,respuesta_correcta\n'

  it('lee filas válidas con su número de línea (la cabecera es la 1)', () => {
    const { filas, errorGeneral } = analizarCsv(
      `${cabecera}"¿Capital?",Guayaquil|Quito|Cuenca,1\n"¿Moneda?",Dólar|Sucre,0\n`,
    )
    expect(errorGeneral).toBeUndefined()
    expect(filas.map((f) => f.linea)).toEqual([2, 3])
    expect(filas[0]).toMatchObject({
      enunciado: '¿Capital?',
      opciones: ['Guayaquil', 'Quito', 'Cuenca'],
      respuesta_correcta: 1,
      errores: [],
    })
  })

  it('acepta las columnas en cualquier orden', () => {
    const { filas } = analizarCsv('respuesta_correcta;opciones;enunciado\n1;Sí|No;¿Verdad?\n')
    expect(filas[0]).toMatchObject({ enunciado: '¿Verdad?', respuesta_correcta: 1 })
  })

  it('marca los errores de cada fila', () => {
    const { filas } = analizarCsv(`${cabecera}"¿Una?",Sola,0\n"¿Otra?",A|B,5\n`)
    expect(filas[0].errores.length).toBeGreaterThan(0)
    expect(filas[1].errores.length).toBeGreaterThan(0)
  })

  it('una respuesta correcta que no es un número es un error de la fila', () => {
    const { filas } = analizarCsv(`${cabecera}"¿Una?",A|B,uno\n`)
    expect(filas[0].errores).toContain('La respuesta correcta debe ser un número entero')
  })

  it('avisa si falta una columna, el archivo está vacío o no tiene preguntas', () => {
    expect(analizarCsv('enunciado,opciones\nx,A|B\n').errorGeneral).toMatch(/respuesta_correcta/)
    expect(analizarCsv('').errorGeneral).toMatch(/vacío/)
    expect(analizarCsv(cabecera).errorGeneral).toMatch(/no tiene preguntas/)
  })

  it('limita a 500 preguntas por carga', () => {
    const filas = Array.from({ length: 501 }, (_, i) => `"¿P${i}?",A|B,0`).join('\n')
    expect(analizarCsv(`${cabecera}${filas}\n`).errorGeneral).toMatch(/500/)
  })
})

describe('erroresDeCarga', () => {
  it('traduce las claves del servidor a líneas legibles', () => {
    expect(
      erroresDeCarga({
        'filas[4].respuesta_correcta': 'Fuera de las opciones',
        'preguntas[0].opciones[1]': 'Repetida',
        archivo: 'No está en UTF-8',
      }),
    ).toEqual([
      'Línea 4 (respuesta_correcta): Fuera de las opciones',
      'Pregunta 1 (opciones[1]): Repetida',
      'archivo: No está en UTF-8',
    ])
  })

  it('sin campos devuelve una lista vacía', () => {
    expect(erroresDeCarga(undefined)).toEqual([])
  })
})

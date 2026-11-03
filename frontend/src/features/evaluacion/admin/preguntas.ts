// Reglas del banco de preguntas (docs/api.md, Banco de preguntas y configuración). El servidor
// las vuelve a validar; aquí se aplican para avisar antes de enviar y para la vista previa
// de la carga masiva.

export const MAX_ENUNCIADO = 1000
export const MIN_OPCIONES = 2
export const MAX_OPCIONES = 8
export const MAX_OPCION = 500
export const MAX_BYTES_CSV = 1024 * 1024
export const MAX_PREGUNTAS_POR_CARGA = 500

export interface BorradorPregunta {
  enunciado: string
  opciones: string[]
  respuesta_correcta: number | null
}

// Errores por campo: `enunciado`, `opciones` (de la lista), `opciones[2]` y `respuesta_correcta`.
export function validarPregunta(p: BorradorPregunta): Record<string, string> {
  const errores: Record<string, string> = {}
  const enunciado = p.enunciado.trim()
  if (!enunciado) errores.enunciado = 'Escriba el enunciado'
  else if ([...enunciado].length > MAX_ENUNCIADO)
    errores.enunciado = `Máximo ${MAX_ENUNCIADO} caracteres`

  if (p.opciones.length < MIN_OPCIONES || p.opciones.length > MAX_OPCIONES) {
    errores.opciones = `Debe tener de ${MIN_OPCIONES} a ${MAX_OPCIONES} opciones`
  }
  const vistas = new Map<string, number>()
  p.opciones.forEach((opcion, i) => {
    const texto = opcion.trim()
    if (!texto) errores[`opciones[${i}]`] = 'La opción no puede estar vacía'
    else if ([...texto].length > MAX_OPCION)
      errores[`opciones[${i}]`] = `Máximo ${MAX_OPCION} caracteres`
    // El servidor compara sin distinguir mayúsculas.
    else if (vistas.has(texto.toLowerCase())) errores[`opciones[${i}]`] = 'La opción está repetida'
    else vistas.set(texto.toLowerCase(), i)
  })

  const correcta = p.respuesta_correcta
  if (correcta === null) errores.respuesta_correcta = 'Marque la opción correcta'
  else if (!Number.isInteger(correcta) || correcta < 0 || correcta >= p.opciones.length)
    errores.respuesta_correcta = 'La respuesta correcta debe ser una de las opciones'
  return errores
}

// Lee texto CSV en filas de celdas: comillas dobles con "" como comilla escapada, saltos de
// línea dentro de comillas y separador `,` o `;` (el que use la cabecera, como Excel en español).
export function leerCsv(texto: string): string[][] {
  const limpio = texto.replace(/^\uFEFF/, '')
  const primera = limpio.split(/\r?\n/, 1)[0] ?? ''
  const separador = primera.split(';').length > primera.split(',').length ? ';' : ','
  const filas: string[][] = []
  let fila: string[] = []
  let celda = ''
  let entreComillas = false
  for (let i = 0; i < limpio.length; i++) {
    const c = limpio[i]
    if (entreComillas) {
      if (c === '"' && limpio[i + 1] === '"') {
        celda += '"'
        i++
      } else if (c === '"') entreComillas = false
      else celda += c
    } else if (c === '"') entreComillas = true
    else if (c === separador) {
      fila.push(celda)
      celda = ''
    } else if (c === '\n' || c === '\r') {
      if (c === '\r' && limpio[i + 1] === '\n') i++
      fila.push(celda)
      filas.push(fila)
      fila = []
      celda = ''
    } else celda += c
  }
  if (celda !== '' || fila.length > 0) {
    fila.push(celda)
    filas.push(fila)
  }
  return filas.filter((f) => f.some((celda) => celda.trim() !== ''))
}

export interface FilaCsv extends BorradorPregunta {
  linea: number // número de línea del archivo (la cabecera es la 1)
  errores: string[]
}

export interface AnalisisCsv {
  filas: FilaCsv[]
  errorGeneral?: string
}

const COLUMNAS = ['enunciado', 'opciones', 'respuesta_correcta'] as const

// Vista previa de un CSV con columnas `enunciado,opciones,respuesta_correcta` (en cualquier
// orden, con cabecera); las opciones van en una celda separadas por una barra vertical.
export function analizarCsv(texto: string): AnalisisCsv {
  const filas = leerCsv(texto)
  if (filas.length === 0) return { filas: [], errorGeneral: 'El archivo está vacío' }
  const cabecera = filas[0].map((c) => c.trim().toLowerCase())
  const faltan = COLUMNAS.filter((c) => !cabecera.includes(c))
  if (faltan.length > 0) {
    return {
      filas: [],
      errorGeneral: `Falta la columna ${faltan.join(', ')} en la cabecera`,
    }
  }
  const posicion = Object.fromEntries(COLUMNAS.map((c) => [c, cabecera.indexOf(c)]))
  const datos = filas.slice(1).map((celdas, i): FilaCsv => {
    const opciones = (celdas[posicion.opciones] ?? '').split('|').map((o) => o.trim())
    const bruto = (celdas[posicion.respuesta_correcta] ?? '').trim()
    const numero = bruto === '' ? null : Number(bruto)
    const borrador: BorradorPregunta = {
      enunciado: (celdas[posicion.enunciado] ?? '').trim(),
      opciones,
      respuesta_correcta: numero !== null && Number.isInteger(numero) ? numero : null,
    }
    const errores = Object.values(validarPregunta(borrador))
    if (numero !== null && !Number.isInteger(numero)) {
      errores.push('La respuesta correcta debe ser un número entero')
    }
    return { ...borrador, linea: i + 2, errores }
  })
  if (datos.length === 0) return { filas: [], errorGeneral: 'El archivo no tiene preguntas' }
  if (datos.length > MAX_PREGUNTAS_POR_CARGA) {
    return {
      filas: datos,
      errorGeneral: `Máximo ${MAX_PREGUNTAS_POR_CARGA} preguntas por carga`,
    }
  }
  return { filas: datos }
}

// Convierte los errores `fields` de una carga masiva (`filas[4].respuesta_correcta`) en una
// lista legible, con el número de línea del archivo.
export function erroresDeCarga(fields: Record<string, string> | undefined): string[] {
  return Object.entries(fields ?? {}).map(([campo, mensaje]) => {
    const fila = campo.match(/^filas\[(\d+)\]\.(.+)$/)
    if (fila) return `Línea ${fila[1]} (${fila[2]}): ${mensaje}`
    const pregunta = campo.match(/^preguntas\[(\d+)\]\.(.+)$/)
    if (pregunta) return `Pregunta ${Number(pregunta[1]) + 1} (${pregunta[2]}): ${mensaje}`
    return `${campo}: ${mensaje}`
  })
}

import { http } from './http'

// Banco de preguntas, configuración y resultados de la evaluación teórica en el panel de
// administración (docs/api.md, Evaluación teórica). Todas las rutas son solo para el
// administrador: el analista recibe 403, también al leer.

export interface PreguntaBanco {
  id: number
  enunciado: string
  opciones: string[]
  respuesta_correcta: number // índice desde 0 de la opción correcta
}

export type PreguntaNueva = Omit<PreguntaBanco, 'id'>

export interface PaginaPreguntas {
  items: PreguntaBanco[]
  total: number
  pagina: number
  por_pagina: number
}

export interface FiltrosPreguntas {
  buscar?: string
  pagina?: number
  por_pagina?: number
}

export interface ConfigEvaluacion {
  duracion_minutos: number
  num_preguntas: number
  total_banco: number
  configurada: boolean
}

export type CambiosConfig = Pick<ConfigEvaluacion, 'duracion_minutos' | 'num_preguntas'>

export interface ResultadoEvaluacion {
  intento_id: string
  codigo: string
  cedula: string
  nombres: string
  apellidos: string
  cargo: string
  estado: 'en_curso' | 'finalizado'
  inicio: string
  cierre: string
  finalizado_at: string | null
  nota: number | null
  aciertos: number | null
  total_preguntas: number
}

// GET /admin/evaluacion/preguntas: lista el banco con la respuesta correcta.
export async function listarPreguntas(signal: AbortSignal, filtros: FiltrosPreguntas = {}) {
  const { buscar, pagina = 1, por_pagina = 20 } = filtros
  const { data } = await http.get<PaginaPreguntas>('/admin/evaluacion/preguntas', {
    signal,
    params: { buscar: buscar || undefined, pagina, por_pagina },
  })
  return data
}

// POST /admin/evaluacion/preguntas con una pregunta: 201 con la pregunta creada.
export async function crearPregunta(signal: AbortSignal, pregunta: PreguntaNueva) {
  const { data } = await http.post<PreguntaBanco>('/admin/evaluacion/preguntas', pregunta, {
    signal,
  })
  return data
}

// PUT /admin/evaluacion/preguntas/{id}: la pregunta completa. 409 si ya hay evaluaciones iniciadas.
export async function actualizarPregunta(signal: AbortSignal, id: number, pregunta: PreguntaNueva) {
  const { data } = await http.put<PreguntaBanco>(`/admin/evaluacion/preguntas/${id}`, pregunta, {
    signal,
  })
  return data
}

// DELETE /admin/evaluacion/preguntas/{id}: 204. 409 si ya hay evaluaciones iniciadas o si el
// banco quedaría con menos preguntas de las que pide la configuración.
export async function eliminarPregunta(signal: AbortSignal, id: number) {
  await http.delete(`/admin/evaluacion/preguntas/${id}`, { signal })
}

export interface CargaMasiva {
  creadas: number
}

// POST /admin/evaluacion/preguntas con un CSV (multipart, campo `archivo`): todas o ninguna.
// 400 con un error por fila en `fields` (`filas[4].respuesta_correcta`); 422 si pesa más de 1 MB.
export async function cargarPreguntasCsv(signal: AbortSignal, archivo: File) {
  const cuerpo = new FormData()
  cuerpo.append('archivo', archivo)
  const { data } = await http.post<CargaMasiva>('/admin/evaluacion/preguntas', cuerpo, { signal })
  return data
}

// GET /admin/evaluacion/config. Mientras no se configure: 30 minutos, 10 preguntas y
// `configurada: false`. Cuándo se puede rendir lo fija la fase Evaluación (/admin/fases).
export async function obtenerConfig(signal: AbortSignal) {
  const { data } = await http.get<ConfigEvaluacion>('/admin/evaluacion/config', { signal })
  return data
}

// PUT /admin/evaluacion/config: 422 con `fields.num_preguntas` si supera las del banco.
export async function guardarConfig(signal: AbortSignal, cambios: CambiosConfig) {
  const { data } = await http.put<ConfigEvaluacion>('/admin/evaluacion/config', cambios, {
    signal,
  })
  return data
}

// GET /admin/evaluacion/resultados: notas de todos los postulantes que iniciaron el examen.
export async function listarResultados(signal: AbortSignal) {
  const { data } = await http.get<{ items: ResultadoEvaluacion[] }>(
    '/admin/evaluacion/resultados',
    { signal },
  )
  return data.items
}

// GET /admin/evaluacion/resultados/{id}/reporte: el PDF. 409 si el examen sigue en curso.
export async function descargarReporte(signal: AbortSignal, intentoId: string) {
  const respuesta = await http.get<Blob>(
    `/admin/evaluacion/resultados/${encodeURIComponent(intentoId)}/reporte`,
    { signal, responseType: 'blob' },
  )
  const disposicion = respuesta.headers['content-disposition']
  const nombre =
    typeof disposicion === 'string' ? disposicion.match(/filename="?([^";]+)"?/i)?.[1] : undefined
  return { contenido: respuesta.data, nombre: nombre ?? 'reporte-evaluacion.pdf' }
}

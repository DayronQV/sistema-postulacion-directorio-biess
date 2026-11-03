import { http } from './http'
import type { Postulacion, TipoCargo } from './postulacion'

// Acciones del panel de administración sobre las postulaciones (docs/api.md, Administración).

export interface ResultadoMerito {
  id: string
  supero_merito: boolean
  estado_visible: string // "Evaluación Técnica" si superó la Fase de Méritos
}

export type EstadoPostulacionAdmin = 'borrador' | 'enviada' | 'cerrada'

export interface ResumenPostulacionAdmin {
  id: string
  codigo: string
  cedula: string
  nombres: string
  apellidos: string
  cargo: TipoCargo | ''
  estado: EstadoPostulacionAdmin
  elegible: boolean
}

export interface PaginaPostulacionesAdmin {
  items: ResumenPostulacionAdmin[]
  total: number
  pagina: number
  por_pagina: number
}

export interface FiltrosPostulacionesAdmin {
  cargo?: TipoCargo | ''
  estado?: EstadoPostulacionAdmin | ''
  elegible?: '' | 'true' | 'false'
  pagina?: number
}

export async function listarPostulacionesAdmin(
  signal: AbortSignal,
  filtros: FiltrosPostulacionesAdmin,
) {
  const params = new URLSearchParams()
  if (filtros.cargo) params.set('cargo', filtros.cargo)
  if (filtros.estado) params.set('estado', filtros.estado)
  if (filtros.elegible) params.set('elegible', filtros.elegible)
  if (filtros.pagina && filtros.pagina > 1) params.set('pagina', String(filtros.pagina))
  const consulta = params.size > 0 ? `?${params.toString()}` : ''
  const { data } = await http.get<PaginaPostulacionesAdmin>(`/admin/postulaciones${consulta}`, {
    signal,
  })
  return data
}

export async function obtenerPostulacionAdmin(signal: AbortSignal, id: string) {
  const { data } = await http.get<Postulacion>(`/admin/postulaciones/${encodeURIComponent(id)}`, {
    signal,
  })
  return data
}

function nombreDescarga(contentDisposition: unknown, respaldo = 'certificado.pdf') {
  if (typeof contentDisposition !== 'string') return respaldo
  const utf8 = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i)?.[1]
  if (utf8) {
    try {
      return decodeURIComponent(utf8)
    } catch {
      return respaldo
    }
  }
  return contentDisposition.match(/filename="?([^";]+)"?/i)?.[1] ?? respaldo
}

export async function descargarDocumentoAdmin(
  signal: AbortSignal,
  postulacionId: string,
  documentoId: string,
) {
  const respuesta = await http.get<Blob>(
    `/admin/postulaciones/${encodeURIComponent(postulacionId)}/documentos/${encodeURIComponent(documentoId)}`,
    { signal, responseType: 'blob' },
  )
  return {
    contenido: respuesta.data,
    nombre: nombreDescarga(respuesta.headers['content-disposition']),
  }
}

export type FormatoMatriz = 'xlsx' | 'pdf'

export async function descargarMatrizAdmin(
  signal: AbortSignal,
  formato: FormatoMatriz,
  filtros: FiltrosPostulacionesAdmin,
) {
  const params = new URLSearchParams({ formato })
  if (filtros.cargo) params.set('cargo', filtros.cargo)
  if (filtros.estado) params.set('estado', filtros.estado)
  if (filtros.elegible) params.set('elegible', filtros.elegible)
  const respuesta = await http.get<Blob>(`/admin/matriz?${params.toString()}`, {
    signal,
    responseType: 'blob',
  })
  return {
    contenido: respuesta.data,
    nombre: nombreDescarga(
      respuesta.headers['content-disposition'],
      `matriz-postulantes.${formato}`,
    ),
  }
}

// PUT /admin/postulaciones/{id}/merito: solo administrador y solo postulaciones enviadas
// (403 para el analista, 404 si no existe, 409 si no está enviada).
export async function marcarMerito(signal: AbortSignal, id: string, supero: boolean) {
  const { data } = await http.put<ResultadoMerito>(
    `/admin/postulaciones/${encodeURIComponent(id)}/merito`,
    { supero },
    { signal },
  )
  return data
}

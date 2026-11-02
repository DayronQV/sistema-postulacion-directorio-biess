import { http } from './http'

// Fases del concurso en el panel de administración (docs/api.md, Administración).

export interface Fase {
  id: number
  nombre: string // "Postulación", "Méritos" o "Evaluación"
  orden: number
  publicada: boolean
  habilita_formulario: boolean // solo la fase Postulación lo habilita
  fecha_inicio: string | null // RFC 3339; null es sin límite
  fecha_fin: string | null
}

export type CambiosFase = Pick<
  Fase,
  'publicada' | 'habilita_formulario' | 'fecha_inicio' | 'fecha_fin'
>

export const FASE_POSTULACION = 'Postulación'

// GET /admin/fases: analista y administrador.
export async function listarFases(signal: AbortSignal) {
  const { data } = await http.get<Fase[]>('/admin/fases', { signal })
  return data
}

// PUT /admin/fases/{id}: solo administrador (403 para el analista; 400 con fields si la
// regla no se cumple).
export async function actualizarFase(signal: AbortSignal, id: number, cambios: CambiosFase) {
  const { data } = await http.put<Fase>(`/admin/fases/${id}`, cambios, { signal })
  return data
}

import { http } from './http'

// Acciones del panel de administración sobre las postulaciones (docs/api.md, Administración).

export interface ResultadoMerito {
  id: string
  supero_merito: boolean
  estado_visible: string // "Evaluación Técnica" si superó la Fase de Méritos
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

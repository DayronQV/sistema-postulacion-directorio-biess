import { http } from './http'

export interface DashboardPostulaciones {
  por_cargo: {
    afiliados: number
    jubilados: number
    empleador: number
  }
  por_estado: {
    borrador: number
    enviada: number
    cerrada: number
  }
  total: number
}

export async function obtenerDashboard(signal: AbortSignal) {
  const { data } = await http.get<DashboardPostulaciones>('/admin/dashboard', { signal })
  return data
}

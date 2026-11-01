import axios, { AxiosError } from 'axios'
import { env } from '@/config/env'
import { cerrarSesion, obtenerToken } from '@/store/sesion'
import { ApiError } from './errores'

interface CuerpoError {
  error?: {
    code?: string
    message?: string
    fields?: Record<string, string>
    retry_after_seconds?: number
  }
}

export const http = axios.create({ baseURL: env.apiUrl, timeout: 15000 })

http.interceptors.request.use((config) => {
  const token = obtenerToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Convierte cualquier fallo en ApiError. Un 401 con sesión abierta significa token vencido
// o inválido: se cierra la sesión y las rutas protegidas redirigen solas al login.
http.interceptors.response.use(
  (respuesta) => respuesta,
  (fallo: unknown) => {
    if (axios.isCancel(fallo)) throw fallo
    if (!(fallo instanceof AxiosError) || !fallo.response) {
      throw new ApiError(0, 'NETWORK_ERROR', 'No se pudo conectar con el servidor.')
    }
    const { status, data } = fallo.response
    if (status === 401 && obtenerToken()) cerrarSesion()
    const cuerpo = (data as CuerpoError | undefined)?.error
    throw new ApiError(
      status,
      cuerpo?.code ?? 'INTERNAL_ERROR',
      cuerpo?.message ?? 'Ocurrió un error inesperado.',
      cuerpo?.fields,
      cuerpo?.retry_after_seconds,
    )
  },
)

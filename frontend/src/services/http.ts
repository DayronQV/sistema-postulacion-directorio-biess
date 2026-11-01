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

// Con responseType 'blob' (descargas) el cuerpo del error también llega como Blob: se lee
// como JSON para no perder el mensaje del servidor.
async function leerCuerpo(data: unknown): Promise<CuerpoError | undefined> {
  if (typeof Blob !== 'undefined' && data instanceof Blob) {
    try {
      return JSON.parse(await data.text()) as CuerpoError
    } catch {
      return undefined
    }
  }
  return data as CuerpoError | undefined
}

// Convierte cualquier fallo en ApiError. Un 401 con sesión abierta significa token vencido
// o inválido: se cierra la sesión y las rutas protegidas redirigen solas al login.
http.interceptors.response.use(
  (respuesta) => respuesta,
  async (fallo: unknown) => {
    if (axios.isCancel(fallo)) throw fallo
    if (!(fallo instanceof AxiosError) || !fallo.response) {
      throw new ApiError(0, 'NETWORK_ERROR', 'No se pudo conectar con el servidor.')
    }
    const { status, data } = fallo.response
    if (status === 401 && obtenerToken()) cerrarSesion()
    const cuerpo = (await leerCuerpo(data))?.error
    throw new ApiError(
      status,
      cuerpo?.code ?? 'INTERNAL_ERROR',
      cuerpo?.message ?? 'Ocurrió un error inesperado.',
      cuerpo?.fields,
      cuerpo?.retry_after_seconds,
    )
  },
)

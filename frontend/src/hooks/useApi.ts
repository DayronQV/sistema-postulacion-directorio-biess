import { useCallback, useEffect, useRef, useState } from 'react'
import axios from 'axios'
import { ApiError } from '@/services/errores'

export interface EstadoApi<T> {
  data: T | null
  error: ApiError | null
  cargando: boolean
}

// Ejecuta una llamada del servicio con estados loading/success/error.
// Cancela la petición anterior (y la pendiente al desmontar) con AbortController.
// `llamada` debe ser estable (función de módulo o useCallback) y recibir la señal como primer argumento.
export function useApi<T, A extends unknown[]>(
  llamada: (signal: AbortSignal, ...args: A) => Promise<T>,
) {
  const [estado, setEstado] = useState<EstadoApi<T>>({ data: null, error: null, cargando: false })
  const control = useRef<AbortController | null>(null)

  useEffect(() => () => control.current?.abort(), [])

  const ejecutar = useCallback(
    async (...args: A): Promise<T | undefined> => {
      control.current?.abort()
      const actual = new AbortController()
      control.current = actual
      setEstado((s) => ({ ...s, error: null, cargando: true }))
      try {
        const data = await llamada(actual.signal, ...args)
        if (!actual.signal.aborted) setEstado({ data, error: null, cargando: false })
        return data
      } catch (e) {
        if (axios.isCancel(e) || actual.signal.aborted) return undefined
        const error =
          e instanceof ApiError ? e : new ApiError(0, 'UNKNOWN', 'Ocurrió un error inesperado.')
        setEstado({ data: null, error, cargando: false })
        return undefined
      }
    },
    [llamada],
  )

  return { ...estado, ejecutar }
}

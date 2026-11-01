import { act, renderHook, waitFor } from '@testing-library/react'
import { ApiError } from '@/services/errores'
import { useApi } from './useApi'

describe('useApi', () => {
  it('pasa de cargando a éxito con los datos', async () => {
    const llamada = async (_: AbortSignal, n: number) => n * 2
    const { result } = renderHook(() => useApi(llamada))
    expect(result.current.cargando).toBe(false)

    await act(async () => {
      await result.current.ejecutar(21)
    })
    expect(result.current).toMatchObject({ data: 42, error: null, cargando: false })
  })

  it('expone el ApiError cuando la llamada falla', async () => {
    const llamada = async () => {
      throw new ApiError(422, 'VALIDATION_ERROR', 'Datos inválidos')
    }
    const { result } = renderHook(() => useApi(llamada))
    await act(async () => {
      await result.current.ejecutar()
    })
    expect(result.current.error?.code).toBe('VALIDATION_ERROR')
    expect(result.current.data).toBeNull()
  })

  it('cancela la petición anterior al lanzar otra', async () => {
    const senales: AbortSignal[] = []
    const llamada = (signal: AbortSignal, ...resto: number[]) => {
      void resto
      senales.push(signal)
      return new Promise<number>(() => undefined)
    }
    const { result } = renderHook(() => useApi(llamada))
    act(() => {
      void result.current.ejecutar(1)
    })
    act(() => {
      void result.current.ejecutar(2)
    })
    await waitFor(() => expect(senales[0].aborted).toBe(true))
    expect(senales[1].aborted).toBe(false)
  })

  it('cancela la petición pendiente al desmontar', () => {
    let senal: AbortSignal | undefined
    const llamada = (signal: AbortSignal) => {
      senal = signal
      return new Promise<void>(() => undefined)
    }
    const { result, unmount } = renderHook(() => useApi(llamada))
    act(() => {
      void result.current.ejecutar()
    })
    unmount()
    expect(senal?.aborted).toBe(true)
  })
})

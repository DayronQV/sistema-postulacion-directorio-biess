import { act, renderHook } from '@testing-library/react'
import { vi } from 'vitest'
import { formatoMinutos, useCuentaRegresiva } from './useCuentaRegresiva'

describe('useCuentaRegresiva', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  it('cuenta hacia atrás y termina en 0', () => {
    const { result } = renderHook(() => useCuentaRegresiva())
    expect(result.current.activa).toBe(false)

    act(() => result.current.iniciar(3))
    expect(result.current).toMatchObject({ restante: 3, activa: true })

    act(() => vi.advanceTimersByTime(1000))
    expect(result.current.restante).toBe(2)

    act(() => vi.advanceTimersByTime(2000))
    expect(result.current).toMatchObject({ restante: 0, activa: false })
  })

  it('no se activa con 0 segundos', () => {
    const { result } = renderHook(() => useCuentaRegresiva())
    act(() => result.current.iniciar(0))
    expect(result.current.activa).toBe(false)
  })
})

describe('formatoMinutos', () => {
  it.each([
    [600, '10:00'],
    [540, '09:00'],
    [61, '01:01'],
    [0, '00:00'],
  ])('%i segundos → %s', (s, texto) => {
    expect(formatoMinutos(s)).toBe(texto)
  })
})

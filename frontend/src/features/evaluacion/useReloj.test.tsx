import { act, renderHook } from '@testing-library/react'
import { vi } from 'vitest'
import { formatoReloj, useReloj } from './useReloj'

describe('formatoReloj', () => {
  it('da mm:ss con ceros a la izquierda', () => {
    expect(formatoReloj(0)).toBe('00:00')
    expect(formatoReloj(65)).toBe('01:05')
    expect(formatoReloj(1799)).toBe('29:59')
  })

  it('agrega las horas desde los 60 minutos', () => {
    expect(formatoReloj(3600)).toBe('1:00:00')
    expect(formatoReloj(3725)).toBe('1:02:05')
  })

  it('no baja de cero ni muestra decimales', () => {
    expect(formatoReloj(-5)).toBe('00:00')
    expect(formatoReloj(59.9)).toBe('00:59')
  })
})

describe('useReloj', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  it('no cuenta hasta que el servidor lo sincroniza', () => {
    const { result } = renderHook(() => useReloj(vi.fn()))
    act(() => vi.advanceTimersByTime(5000))

    expect(result.current.restante).toBeNull()
  })

  it('cuenta hacia atrás desde el tiempo que dijo el servidor', () => {
    const { result } = renderHook(() => useReloj(vi.fn()))
    act(() => result.current.sincronizar(90))
    expect(result.current.restante).toBe(90)

    act(() => vi.advanceTimersByTime(30_000))
    expect(result.current.restante).toBe(60)
  })

  it('al llegar a 0 avisa una sola vez', () => {
    const alAgotarse = vi.fn()
    const { result } = renderHook(() => useReloj(alAgotarse))
    act(() => result.current.sincronizar(3))

    act(() => vi.advanceTimersByTime(3000))
    expect(result.current.restante).toBe(0)
    expect(alAgotarse).toHaveBeenCalledTimes(1)

    act(() => vi.advanceTimersByTime(10_000))
    expect(alAgotarse).toHaveBeenCalledTimes(1)
  })

  it('sincronizar de nuevo corrige el tiempo con el del servidor', () => {
    const { result } = renderHook(() => useReloj(vi.fn()))
    act(() => result.current.sincronizar(100))
    act(() => vi.advanceTimersByTime(10_000))
    expect(result.current.restante).toBe(90)

    act(() => result.current.sincronizar(40))
    expect(result.current.restante).toBe(40)
    act(() => vi.advanceTimersByTime(10_000))
    expect(result.current.restante).toBe(30)
  })

  it('detener deja de contar y no avisa', () => {
    const alAgotarse = vi.fn()
    const { result } = renderHook(() => useReloj(alAgotarse))
    act(() => result.current.sincronizar(5))
    act(() => result.current.detener())
    act(() => vi.advanceTimersByTime(10_000))

    expect(result.current.restante).toBeNull()
    expect(alAgotarse).not.toHaveBeenCalled()
  })
})

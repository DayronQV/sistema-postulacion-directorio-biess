import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, vi } from 'vitest'
import { DURACION_TOAST_MS, cerrarToast, mostrarToast, reiniciarToasts, useToasts } from './toasts'

const lista = () => renderHook(() => useToasts()).result.current
const avanzar = (ms: number) => act(() => void vi.advanceTimersByTime(ms))

beforeEach(() => {
  vi.useFakeTimers()
  reiniciarToasts()
})

afterEach(() => {
  reiniciarToasts()
  vi.useRealTimers()
})

describe('toasts', () => {
  it('muestra un aviso de éxito con su mensaje', () => {
    mostrarToast('Guardado')
    expect(lista()).toEqual([expect.objectContaining({ tipo: 'success', mensaje: 'Guardado' })])
  })

  it('se quita solo a los 4 segundos, ni antes ni después', () => {
    expect(DURACION_TOAST_MS).toBe(4000)
    mostrarToast('Guardado')

    avanzar(DURACION_TOAST_MS - 1)
    expect(lista()).toHaveLength(1)
    avanzar(1)
    expect(lista()).toHaveLength(0)
  })

  it('el mismo mensaje no se apila: reinicia su tiempo', () => {
    mostrarToast('Guardado')
    avanzar(3000)
    mostrarToast('Guardado')
    expect(lista()).toHaveLength(1)

    avanzar(3000) // 6 s desde el primero, pero solo 3 s desde el último
    expect(lista()).toHaveLength(1)
    avanzar(1000)
    expect(lista()).toHaveLength(0)
  })

  it('mensajes distintos se apilan, hasta 3: el más antiguo sale primero', () => {
    for (const mensaje of ['Uno', 'Dos', 'Tres', 'Cuatro']) mostrarToast(mensaje)
    expect(lista().map((t) => t.mensaje)).toEqual(['Dos', 'Tres', 'Cuatro'])
  })

  it('un toast que salió por el límite no vuelve a aparecer al vencer su tiempo', () => {
    mostrarToast('Uno')
    for (const mensaje of ['Dos', 'Tres', 'Cuatro']) mostrarToast(mensaje)
    expect(lista().map((t) => t.mensaje)).not.toContain('Uno')

    avanzar(DURACION_TOAST_MS)
    expect(lista()).toHaveLength(0)
  })

  it('cerrarToast quita solo el indicado y ignora un id que no existe', () => {
    const uno = mostrarToast('Uno')
    mostrarToast('Dos')
    cerrarToast(9999)
    expect(lista()).toHaveLength(2)

    cerrarToast(uno)
    expect(lista().map((t) => t.mensaje)).toEqual(['Dos'])
  })

  it('admite otros tipos y otra duración', () => {
    mostrarToast('Algo falló', 'error', 1000)
    expect(lista()[0].tipo).toBe('error')
    avanzar(1000)
    expect(lista()).toHaveLength(0)
  })
})

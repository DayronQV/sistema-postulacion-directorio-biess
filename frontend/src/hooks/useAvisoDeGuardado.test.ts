import { renderHook } from '@testing-library/react'
import { afterEach, beforeEach, vi } from 'vitest'
import { cerrarToast, reiniciarToasts, useToasts } from '@/store/toasts'
import { useAvisoDeGuardado } from './useAvisoDeGuardado'

const avisos = () => renderHook(() => useToasts()).result.current

beforeEach(() => {
  vi.useFakeTimers()
  reiniciarToasts()
})

afterEach(() => {
  reiniciarToasts()
  vi.useRealTimers()
})

function montar(guardado: boolean) {
  return renderHook(({ g }) => useAvisoDeGuardado(g, 'Hecho'), { initialProps: { g: guardado } })
}

describe('useAvisoDeGuardado', () => {
  it('avisa cuando el guardado pasa de falso a verdadero', () => {
    const { rerender } = montar(false)
    expect(avisos()).toHaveLength(0)

    rerender({ g: true })
    expect(avisos().map((t) => t.mensaje)).toEqual(['Hecho'])
  })

  it('no repite el aviso mientras siga verdadero', () => {
    const { rerender } = montar(false)
    rerender({ g: true })
    cerrarToast(avisos()[0].id)

    rerender({ g: true })
    expect(avisos()).toHaveLength(0)
  })

  it('vuelve a avisar en el siguiente guardado', () => {
    const { rerender } = montar(false)
    rerender({ g: true })
    cerrarToast(avisos()[0].id)

    rerender({ g: false })
    rerender({ g: true })
    expect(avisos().map((t) => t.mensaje)).toEqual(['Hecho'])
  })

  it('no avisa si el componente ya nace con el guardado en verdadero', () => {
    montar(true)
    expect(avisos()).toHaveLength(0)
  })

  it('aunque haya nacido guardado, el siguiente guardado sí avisa', () => {
    const { rerender } = montar(true)

    rerender({ g: false })
    rerender({ g: true })
    expect(avisos().map((t) => t.mensaje)).toEqual(['Hecho'])
  })
})

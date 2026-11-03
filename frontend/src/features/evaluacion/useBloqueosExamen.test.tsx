import { renderHook } from '@testing-library/react'
import { vi } from 'vitest'
import { useBloqueosExamen } from './useBloqueosExamen'

function lanzar(nombre: string) {
  const evento = new Event(nombre, { cancelable: true, bubbles: true })
  document.body.dispatchEvent(evento)
  return evento
}

describe('useBloqueosExamen', () => {
  it('con el examen activo bloquea copiar, cortar, pegar y el menú contextual', () => {
    renderHook(() => useBloqueosExamen(true))

    for (const nombre of ['copy', 'cut', 'paste', 'contextmenu']) {
      expect(lanzar(nombre).defaultPrevented).toBe(true)
    }
  })

  it('sin examen no bloquea nada', () => {
    renderHook(() => useBloqueosExamen(false))

    for (const nombre of ['copy', 'cut', 'paste', 'contextmenu']) {
      expect(lanzar(nombre).defaultPrevented).toBe(false)
    }
  })

  it('"Atrás" del navegador vuelve a empujar la página al historial', () => {
    const empujar = vi.spyOn(window.history, 'pushState')
    renderHook(() => useBloqueosExamen(true))
    expect(empujar).toHaveBeenCalledTimes(1)

    window.dispatchEvent(new PopStateEvent('popstate'))
    expect(empujar).toHaveBeenCalledTimes(2)
    empujar.mockRestore()
  })

  it('avisa antes de cerrar o recargar la pestaña', () => {
    renderHook(() => useBloqueosExamen(true))
    const evento = new Event('beforeunload', { cancelable: true })
    window.dispatchEvent(evento)

    expect(evento.defaultPrevented).toBe(true)
  })

  it('al terminar el examen quita todos los bloqueos', () => {
    const { rerender } = renderHook(({ activo }) => useBloqueosExamen(activo), {
      initialProps: { activo: true },
    })
    rerender({ activo: false })

    expect(lanzar('paste').defaultPrevented).toBe(false)
    const salida = new Event('beforeunload', { cancelable: true })
    window.dispatchEvent(salida)
    expect(salida.defaultPrevented).toBe(false)
  })

  it('al desmontar quita los bloqueos', () => {
    const { unmount } = renderHook(() => useBloqueosExamen(true))
    unmount()

    expect(lanzar('copy').defaultPrevented).toBe(false)
  })
})

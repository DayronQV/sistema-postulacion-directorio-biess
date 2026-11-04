import { render, screen } from '@testing-library/react'
import { vi } from 'vitest'
import PanelEditor from './PanelEditor'

const desplazar = vi.fn()

function simularMovimientoReducido(reducido: boolean) {
  vi.stubGlobal('matchMedia', (consulta: string) => ({
    matches: reducido && consulta.includes('reduce'),
    media: consulta,
  }))
}

beforeEach(() => {
  desplazar.mockReset()
  Element.prototype.scrollIntoView = desplazar
  simularMovimientoReducido(false)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

function panel(clave: unknown) {
  return (
    <PanelEditor clave={clave}>
      <label htmlFor="tipo">Tipo</label>
      <select id="tipo" />
      <input aria-label="Cargo" />
    </PanelEditor>
  )
}

describe('PanelEditor', () => {
  it('al abrirse baja la pantalla hasta el panel, con animación', () => {
    render(panel(null))
    expect(desplazar).toHaveBeenCalledTimes(1)
    expect(desplazar).toHaveBeenCalledWith({ behavior: 'smooth', block: 'start' })
  })

  it('pone el cursor en el primer campo del panel', () => {
    render(panel(null))
    expect(screen.getByLabelText('Tipo')).toHaveFocus()
  })

  it('sin animación si el sistema pide reducir el movimiento', () => {
    simularMovimientoReducido(true)
    render(panel(null))
    expect(desplazar).toHaveBeenCalledWith({ behavior: 'auto', block: 'start' })
  })

  it('vuelve a bajar al pasar a otra fila, pero no en cada repintado', () => {
    const { rerender } = render(panel(null))
    rerender(panel(null))
    expect(desplazar).toHaveBeenCalledTimes(1)

    rerender(panel(2))
    expect(desplazar).toHaveBeenCalledTimes(2)
  })

  it('deja libre la barra superior fija', () => {
    const { container } = render(panel(null))
    expect(container.firstElementChild).toHaveClass('scroll-mt-24')
  })
})

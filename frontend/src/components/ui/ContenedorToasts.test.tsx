import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, vi } from 'vitest'
import { DURACION_TOAST_MS, mostrarToast, reiniciarToasts } from '@/store/toasts'
import ContenedorToasts from './ContenedorToasts'

beforeEach(() => {
  vi.useFakeTimers()
  reiniciarToasts()
  render(<ContenedorToasts />)
})

afterEach(() => {
  reiniciarToasts()
  vi.useRealTimers()
})

describe('ContenedorToasts', () => {
  it('no muestra nada mientras no haya avisos', () => {
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('muestra el aviso y lo quita solo a los 4 segundos', () => {
    act(() => void mostrarToast('La información se guardó correctamente.'))
    expect(screen.getByRole('status')).toHaveTextContent('La información se guardó correctamente.')

    act(() => void vi.advanceTimersByTime(DURACION_TOAST_MS - 1))
    expect(screen.getByRole('status')).toBeInTheDocument()
    act(() => void vi.advanceTimersByTime(1))
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('se puede cerrar antes con su botón', () => {
    act(() => void mostrarToast('Guardado'))
    fireEvent.click(screen.getByRole('button', { name: 'Cerrar aviso' }))
    expect(screen.queryByText('Guardado')).not.toBeInTheDocument()
  })

  it('un aviso de error se anuncia como alerta y uno de éxito como estado', () => {
    act(() => {
      mostrarToast('Listo')
      mostrarToast('Falló', 'error')
    })
    expect(screen.getByRole('status')).toHaveTextContent('Listo')
    expect(screen.getByRole('alert')).toHaveTextContent('Falló')
  })

  it('es una región en vivo y queda por encima de la barra y de las ventanas', () => {
    const region = document.querySelector('[aria-live="polite"]')
    expect(region).toBeInTheDocument()
    expect(region?.className).toContain('z-[70]')
  })
})

import { act, render, screen } from '@testing-library/react'
import { afterEach, beforeEach } from 'vitest'
import { mostrarToast, reiniciarToasts } from '@/store/toasts'
import App from './App'

beforeEach(() => reiniciarToasts())
afterEach(() => reiniciarToasts())

describe('App', () => {
  it('monta el contenedor de toasts: un aviso se ve en cualquier pantalla', () => {
    render(<App />)

    act(() => void mostrarToast('La información se guardó correctamente.'))

    expect(screen.getByRole('status')).toHaveTextContent('La información se guardó correctamente.')
  })
})

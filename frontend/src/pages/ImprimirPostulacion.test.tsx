import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { vi } from 'vitest'
import { reiniciarFormulario } from '@/features/postulacion/store'
import { ApiError } from '@/services/errores'
import * as servicio from '@/services/postulacion'
import ImprimirPostulacion from './ImprimirPostulacion'

vi.mock('@/services/postulacion')
const obtener = vi.mocked(servicio.obtenerPostulacion)

function montar() {
  render(
    <MemoryRouter>
      <ImprimirPostulacion />
    </MemoryRouter>,
  )
}

beforeEach(() => {
  reiniciarFormulario()
  obtener.mockReset()
})

describe('ImprimirPostulacion', () => {
  it('carga la postulación y permite imprimir', async () => {
    obtener.mockResolvedValue({ codigo: 'POST-9' } as servicio.Postulacion)
    const imprimir = vi.spyOn(window, 'print').mockImplementation(() => undefined)
    montar()

    expect(await screen.findByText('POST-9')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Imprimir / Guardar PDF' }))
    expect(imprimir).toHaveBeenCalledTimes(1)
  })

  it('los controles se ocultan al imprimir', async () => {
    obtener.mockResolvedValue({ codigo: 'POST-9' } as servicio.Postulacion)
    montar()
    await screen.findByText('POST-9')
    const barra = screen.getByRole('link', { name: /Volver al formulario/ }).parentElement
    expect(barra).toHaveClass('print:hidden')
  })

  it('muestra el error de carga y deja reintentar', async () => {
    obtener.mockRejectedValueOnce(new ApiError(500, 'INTERNAL_ERROR', 'Falla del servidor'))
    obtener.mockResolvedValueOnce({ codigo: 'POST-9' } as servicio.Postulacion)
    montar()

    expect(await screen.findByText('Falla del servidor')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Imprimir / Guardar PDF' })).toBeDisabled()
    await userEvent.click(screen.getByRole('button', { name: 'Reintentar' }))
    expect(await screen.findByText('POST-9')).toBeInTheDocument()
  })
})

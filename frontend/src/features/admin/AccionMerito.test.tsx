import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { marcarMerito } from '@/services/adminPostulaciones'
import { ApiError } from '@/services/errores'
import { cerrarSesion, iniciarSesion } from '@/store/sesion'
import AccionMerito from './AccionMerito'

vi.mock('@/services/adminPostulaciones', () => ({ marcarMerito: vi.fn() }))
const mockMarcar = marcarMerito as Mock

const ID = '11111111-2222-3333-4444-555555555555'
const jwt = (rol: string) => `x.${btoa(JSON.stringify({ rol }))}.y`

function montar(props: Partial<Parameters<typeof AccionMerito>[0]> = {}) {
  const onMarcado = vi.fn()
  render(
    <AccionMerito
      postulacionId={ID}
      codigo="BIESS-2026-000001"
      superoMerito={false}
      enviada
      onMarcado={onMarcado}
      {...props}
    />,
  )
  return { onMarcado }
}

// Con llaves: si el callback devolviera el mock, Vitest lo ejecutaría como limpieza al final.
beforeEach(() => {
  mockMarcar.mockReset()
})
afterEach(() => cerrarSesion())

describe('AccionMerito', () => {
  it('el administrador marca con confirmación y queda en Evaluación Técnica', async () => {
    iniciarSesion(jwt('administrador'), 600)
    mockMarcar.mockResolvedValue({
      id: ID,
      supero_merito: true,
      estado_visible: 'Evaluación Técnica',
    })
    const user = userEvent.setup()
    const { onMarcado } = montar()

    await user.click(screen.getByRole('button', { name: /Marcar que superó/ }))
    expect(screen.getByText(/pasará a "Evaluación Técnica"/)).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Marcar' }))

    expect(mockMarcar).toHaveBeenCalledWith(expect.any(AbortSignal), ID, true)
    expect(await screen.findByText('Superó la Fase de Méritos')).toBeInTheDocument()
    expect(onMarcado).toHaveBeenCalledWith(expect.objectContaining({ supero_merito: true }))
    expect(screen.getByRole('button', { name: /Quitar la marca/ })).toBeInTheDocument()
  })

  it('si se cancela la confirmación no llama al servidor', async () => {
    iniciarSesion(jwt('administrador'), 600)
    const user = userEvent.setup()
    montar()
    await user.click(screen.getByRole('button', { name: /Marcar que superó/ }))
    await user.click(screen.getByRole('button', { name: 'Cancelar' }))
    expect(mockMarcar).not.toHaveBeenCalled()
  })

  it('el administrador puede quitar la marca', async () => {
    iniciarSesion(jwt('administrador'), 600)
    mockMarcar.mockResolvedValue({ id: ID, supero_merito: false, estado_visible: 'Enviada' })
    const user = userEvent.setup()
    montar({ superoMerito: true })

    await user.click(screen.getByRole('button', { name: /Quitar la marca/ }))
    await user.click(
      within(screen.getByRole('dialog')).getByRole('button', { name: 'Quitar la marca' }),
    )
    expect(mockMarcar).toHaveBeenCalledWith(expect.any(AbortSignal), ID, false)
    expect(await screen.findByText('Fase de Méritos pendiente')).toBeInTheDocument()
  })

  it('el analista solo ve el estado, sin acción', () => {
    iniciarSesion(jwt('analista'), 600)
    montar({ superoMerito: true })
    expect(screen.getByText('Superó la Fase de Méritos')).toBeInTheDocument()
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('sin postulación enviada no ofrece la acción', () => {
    iniciarSesion(jwt('administrador'), 600)
    montar({ enviada: false })
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('muestra el error del servidor y no cambia el estado', async () => {
    iniciarSesion(jwt('administrador'), 600)
    mockMarcar.mockRejectedValue(
      new ApiError(
        409,
        'CONFLICT',
        'Solo se puede marcar la Fase de Méritos de una postulación enviada',
      ),
    )
    const user = userEvent.setup()
    montar()
    await user.click(screen.getByRole('button', { name: /Marcar que superó/ }))
    await user.click(screen.getByRole('button', { name: 'Marcar' }))

    expect(
      await screen.findByText('Solo se puede marcar la Fase de Méritos de una postulación enviada'),
    ).toBeInTheDocument()
    expect(screen.getByText('Fase de Méritos pendiente')).toBeInTheDocument()
  })
})

describe('AccionMerito, estado que llega del padre', () => {
  it('si el padre recarga la fila con otro valor, la insignia lo sigue', () => {
    iniciarSesion(jwt('administrador'), 600)
    const props = { postulacionId: ID, codigo: 'BIESS-2026-000001', enviada: true }
    const { rerender } = render(<AccionMerito {...props} superoMerito={false} />)
    expect(screen.getByText('Fase de Méritos pendiente')).toBeInTheDocument()

    rerender(<AccionMerito {...props} superoMerito />)
    expect(screen.getByText('Superó la Fase de Méritos')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Quitar la marca/ })).toBeInTheDocument()
  })

  it('sin código, la confirmación no deja un espacio doble', async () => {
    iniciarSesion(jwt('administrador'), 600)
    const user = userEvent.setup()
    montar({ codigo: undefined })
    await user.click(screen.getByRole('button', { name: /Marcar que superó/ }))
    // getByText colapsa los espacios repetidos: se lee el texto crudo del mensaje.
    const mensaje = screen.getByText(/pasará a "Evaluación Técnica"/).textContent ?? ''
    expect(mensaje.startsWith('La postulación pasará a')).toBe(true)
  })
})

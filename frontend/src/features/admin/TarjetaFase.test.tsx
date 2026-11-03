import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { actualizarFase, type Fase } from '@/services/adminFases'
import { ApiError } from '@/services/errores'
import TarjetaFase from './TarjetaFase'
import ContenedorToasts from '@/components/ui/ContenedorToasts'
import { reiniciarToasts } from '@/store/toasts'

vi.mock('@/services/adminFases', async (original) => ({
  ...(await original<typeof import('@/services/adminFases')>()),
  actualizarFase: vi.fn(),
}))
const mockActualizar = actualizarFase as Mock

const postulacion: Fase = {
  id: 1,
  nombre: 'Postulación',
  orden: 1,
  publicada: true,
  habilita_formulario: true,
  fecha_inicio: null,
  fecha_fin: null,
}
const meritos: Fase = {
  ...postulacion,
  id: 2,
  nombre: 'Méritos',
  orden: 2,
  publicada: false,
  habilita_formulario: false,
}

// Con llaves: si el callback devolviera el mock, Vitest lo ejecutaría como limpieza al final.
beforeEach(() => {
  // El aviso de guardado es un toast: se monta su contenedor y se limpia entre pruebas.
  reiniciarToasts()
  render(<ContenedorToasts />)
  mockActualizar.mockReset()
})

describe('TarjetaFase', () => {
  it('el administrador guarda la fase con sus fechas en RFC 3339', async () => {
    const user = userEvent.setup()
    mockActualizar.mockImplementation((_s, id, c) => Promise.resolve({ ...postulacion, id, ...c }))
    const onGuardada = vi.fn()
    render(<TarjetaFase fase={postulacion} editable onGuardada={onGuardada} />)

    await user.type(screen.getByLabelText('Fecha de inicio'), '2026-11-01T08:00')
    await user.click(screen.getByLabelText('Habilitar el formulario de postulación'))
    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('Fase guardada.')).toBeInTheDocument()
    expect(mockActualizar).toHaveBeenCalledWith(expect.any(AbortSignal), 1, {
      publicada: true,
      habilita_formulario: false,
      fecha_inicio: new Date('2026-11-01T08:00').toISOString(),
      fecha_fin: null,
    })
    expect(onGuardada).toHaveBeenCalled()
    expect(screen.getByText(/^Desde el/)).toBeInTheDocument()
  })

  it('al despublicar deshabilita el formulario', async () => {
    const user = userEvent.setup()
    render(<TarjetaFase fase={postulacion} editable />)
    const habilita = screen.getByLabelText('Habilitar el formulario de postulación')

    await user.click(screen.getByLabelText('Fase publicada'))
    expect(habilita).not.toBeChecked()
    expect(habilita).toBeDisabled()
  })

  it('solo la fase Postulación ofrece habilitar el formulario', () => {
    render(<TarjetaFase fase={meritos} editable />)
    expect(screen.getByText('Sin publicar')).toBeInTheDocument()
    expect(screen.queryByLabelText('Habilitar el formulario de postulación')).toBeNull()
  })

  it('no envía si la fecha de fin es anterior a la de inicio', async () => {
    const user = userEvent.setup()
    render(<TarjetaFase fase={postulacion} editable />)

    await user.type(screen.getByLabelText('Fecha de inicio'), '2026-11-10T08:00')
    await user.type(screen.getByLabelText('Fecha de fin'), '2026-11-01T08:00')
    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(
      screen.getByText('La fecha de fin no puede ser anterior a la de inicio'),
    ).toBeInTheDocument()
    expect(mockActualizar).not.toHaveBeenCalled()
  })

  it('muestra el error del servidor', async () => {
    const user = userEvent.setup()
    mockActualizar.mockRejectedValue(
      new ApiError(403, 'FORBIDDEN', 'No tiene permisos para esta acción'),
    )
    render(<TarjetaFase fase={postulacion} editable />)

    await user.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(await screen.findByText('No tiene permisos para esta acción')).toBeInTheDocument()
  })

  it('el analista la ve sin poder cambiarla', () => {
    render(<TarjetaFase fase={postulacion} editable={false} />)
    expect(screen.getByLabelText('Fase publicada')).toBeDisabled()
    expect(screen.getByLabelText('Fecha de inicio')).toBeDisabled()
    expect(screen.queryByRole('button', { name: 'Guardar' })).toBeNull()
  })
})

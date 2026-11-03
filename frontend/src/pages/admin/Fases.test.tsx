import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { listarFases, type Fase } from '@/services/adminFases'
import { ApiError } from '@/services/errores'
import { cerrarSesion, iniciarSesion } from '@/store/sesion'
import Fases from './Fases'

vi.mock('@/services/adminFases', async (original) => ({
  ...(await original<typeof import('@/services/adminFases')>()),
  listarFases: vi.fn(),
  actualizarFase: vi.fn(),
}))
const mockListar = listarFases as Mock

const jwt = (rol: string) => `x.${btoa(JSON.stringify({ rol }))}.y`
const fases: Fase[] = [
  {
    id: 1,
    nombre: 'Postulación',
    orden: 1,
    publicada: true,
    habilita_formulario: true,
    fecha_inicio: null,
    fecha_fin: null,
  },
  {
    id: 2,
    nombre: 'Méritos',
    orden: 2,
    publicada: false,
    habilita_formulario: false,
    fecha_inicio: null,
    fecha_fin: null,
  },
]

// Con llaves: si el callback devolviera el mock, Vitest lo ejecutaría como limpieza al final.
beforeEach(() => {
  mockListar.mockReset()
})
afterEach(() => cerrarSesion())

describe('pantalla de fases', () => {
  it('el administrador ve una tarjeta editable por fase', async () => {
    iniciarSesion(jwt('administrador'), 600)
    mockListar.mockResolvedValue(fases)
    render(<Fases />)

    expect(screen.getByText('Cargando las fases...')).toBeInTheDocument()
    expect(await screen.findByText('1. Postulación')).toBeInTheDocument()
    expect(screen.getByText('2. Méritos')).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: 'Guardar' })).toHaveLength(2)
    expect(screen.queryByText('Acceso de solo lectura')).toBeNull()
  })

  it('el analista las ve en solo lectura', async () => {
    iniciarSesion(jwt('analista'), 600)
    mockListar.mockResolvedValue(fases)
    render(<Fases />)

    expect(await screen.findByText('1. Postulación')).toBeInTheDocument()
    expect(screen.getByText('Acceso de solo lectura')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Guardar' })).toBeNull()
  })

  it('muestra el error de carga y permite reintentar', async () => {
    iniciarSesion(jwt('administrador'), 600)
    mockListar.mockRejectedValueOnce(new ApiError(500, 'INTERNAL_ERROR', 'Error del servidor'))
    mockListar.mockResolvedValueOnce(fases)
    render(<Fases />)

    expect(await screen.findByText('Error del servidor')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Reintentar' }))
    expect(await screen.findByText('1. Postulación')).toBeInTheDocument()
  })
})

import { act, render, screen } from '@testing-library/react'
import { vi, type Mock } from 'vitest'
import { obtenerDashboard } from '@/services/adminDashboard'
import { ApiError } from '@/services/errores'
import Dashboard from './Dashboard'

vi.mock('@/services/adminDashboard', () => ({ obtenerDashboard: vi.fn() }))
const mockObtener = obtenerDashboard as Mock

const datos = {
  por_cargo: { afiliados: 4, jubilados: 3, empleador: 2 },
  por_estado: { borrador: 2, enviada: 6, cerrada: 1 },
  total: 9,
}

beforeEach(() => {
  mockObtener.mockReset()
})

it('muestra los conteos por estado y cargo', async () => {
  mockObtener.mockResolvedValue(datos)
  render(<Dashboard />)

  expect(await screen.findByText('9')).toBeInTheDocument()
  expect(screen.getByLabelText('Afiliados: 4')).toBeInTheDocument()
  expect(screen.getByLabelText('Jubilados: 3')).toBeInTheDocument()
  expect(screen.getByLabelText('Empleador: 2')).toBeInTheDocument()
  expect(mockObtener).toHaveBeenCalledWith(expect.any(AbortSignal))
})

it('vuelve a consultar cada 30 segundos', async () => {
  vi.useFakeTimers()
  try {
    mockObtener.mockResolvedValue(datos)
    render(<Dashboard />)
    await act(async () => {})
    expect(mockObtener).toHaveBeenCalledTimes(1)

    await act(async () => {
      await vi.advanceTimersByTimeAsync(30_000)
    })
    expect(mockObtener).toHaveBeenCalledTimes(2)
  } finally {
    vi.useRealTimers()
  }
})

it('muestra el error sin dejar la pantalla bloqueada', async () => {
  mockObtener.mockRejectedValue(new ApiError(500, 'INTERNAL_ERROR', 'No disponible'))
  render(<Dashboard />)

  expect(await screen.findByText('No disponible')).toBeInTheDocument()
  expect(screen.queryByText('Cargando indicadores...')).not.toBeInTheDocument()
})

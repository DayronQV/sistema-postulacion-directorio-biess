import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import {
  descargarReporte,
  listarResultados,
  type ResultadoEvaluacion,
} from '@/services/adminEvaluacion'
import { ApiError } from '@/services/errores'
import EvaluacionResultados from './EvaluacionResultados'

vi.mock('@/services/adminEvaluacion', async (original) => ({
  ...(await original<typeof import('@/services/adminEvaluacion')>()),
  listarResultados: vi.fn(),
  descargarReporte: vi.fn(),
}))
const mockListar = listarResultados as Mock
const mockDescargar = descargarReporte as Mock

const base = {
  codigo: 'BIESS-2026-000001',
  cedula: '1712345675',
  nombres: 'Carlos',
  apellidos: 'Mendoza',
  cargo: 'afiliados',
  inicio: '2026-11-03T15:00:00Z',
  cierre: '2026-11-03T15:30:00Z',
  total_preguntas: 10,
}
const finalizado: ResultadoEvaluacion = {
  ...base,
  intento_id: 'i-1',
  estado: 'finalizado',
  finalizado_at: '2026-11-03T15:20:00Z',
  nota: 80,
  aciertos: 8,
}
const enCurso: ResultadoEvaluacion = {
  ...base,
  intento_id: 'i-2',
  nombres: 'Ana',
  apellidos: 'Pérez',
  estado: 'en_curso',
  finalizado_at: null,
  nota: null,
  aciertos: null,
}

beforeEach(() => {
  mockListar.mockReset().mockResolvedValue([finalizado, enCurso])
  mockDescargar.mockReset()
})

describe('EvaluacionResultados', () => {
  it('lista las notas de los exámenes terminados y marca los que siguen en curso', async () => {
    render(<EvaluacionResultados />)

    expect(await screen.findByText('Carlos Mendoza')).toBeInTheDocument()
    expect(screen.getByText('80.00 / 100')).toBeInTheDocument()
    expect(screen.getByText('8 de 10')).toBeInTheDocument()
    expect(screen.getByText('Finalizado')).toBeInTheDocument()
    expect(screen.getByText('En curso')).toBeInTheDocument()
    expect(screen.getByText('Disponible al terminar')).toBeInTheDocument()
  })

  it('solo ofrece el PDF de los exámenes finalizados', async () => {
    render(<EvaluacionResultados />)
    await screen.findByText('Carlos Mendoza')

    expect(
      screen.getByRole('button', { name: 'Descargar el reporte de Carlos Mendoza' }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /reporte de Ana Pérez/ })).not.toBeInTheDocument()
  })

  it('descarga el reporte PDF', async () => {
    mockDescargar.mockResolvedValue({ contenido: new Blob(['%PDF']), nombre: 'reporte.pdf' })
    const crear = vi.fn(() => 'blob:reporte')
    const revocar = vi.fn()
    URL.createObjectURL = crear
    URL.revokeObjectURL = revocar
    render(<EvaluacionResultados />)
    await screen.findByText('Carlos Mendoza')

    await userEvent.click(
      screen.getByRole('button', { name: 'Descargar el reporte de Carlos Mendoza' }),
    )

    await vi.waitFor(() => expect(crear).toHaveBeenCalled())
    expect(mockDescargar).toHaveBeenCalledWith(expect.anything(), 'i-1')
  })

  it('si falla la descarga muestra el mensaje', async () => {
    mockDescargar.mockRejectedValue(new ApiError(409, 'CONFLICT', 'El examen sigue en curso'))
    render(<EvaluacionResultados />)
    await screen.findByText('Carlos Mendoza')

    await userEvent.click(
      screen.getByRole('button', { name: 'Descargar el reporte de Carlos Mendoza' }),
    )

    expect(await screen.findByText('El examen sigue en curso')).toBeInTheDocument()
  })

  it('muestra un mensaje si nadie inició la evaluación', async () => {
    mockListar.mockResolvedValue([])
    render(<EvaluacionResultados />)

    expect(await screen.findByText(/ningún postulante inició la evaluación/)).toBeInTheDocument()
  })

  it('muestra el error y Actualizar vuelve a consultar', async () => {
    mockListar.mockRejectedValueOnce(new ApiError(403, 'FORBIDDEN', 'No tiene permiso'))
    render(<EvaluacionResultados />)
    expect(await screen.findByText('No tiene permiso')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Actualizar' }))

    expect(await screen.findByText('Carlos Mendoza')).toBeInTheDocument()
    expect(mockListar).toHaveBeenCalledTimes(2)
  })
})

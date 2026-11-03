import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { vi, type Mock } from 'vitest'
import { guardarConfig, obtenerConfig } from '@/services/adminEvaluacion'
import { ApiError } from '@/services/errores'
import EvaluacionConfiguracion from './EvaluacionConfiguracion'

vi.mock('@/services/adminEvaluacion', async (original) => ({
  ...(await original<typeof import('@/services/adminEvaluacion')>()),
  obtenerConfig: vi.fn(),
  guardarConfig: vi.fn(),
}))
const mockObtener = obtenerConfig as Mock
const mockGuardar = guardarConfig as Mock

const config = { duracion_minutos: 30, num_preguntas: 10, total_banco: 40, configurada: true }

function montar() {
  render(
    <MemoryRouter>
      <EvaluacionConfiguracion />
    </MemoryRouter>,
  )
}

beforeEach(() => {
  mockObtener.mockReset().mockResolvedValue(config)
  mockGuardar.mockReset()
})

describe('EvaluacionConfiguracion', () => {
  it('muestra la configuración actual y cuántas preguntas tiene el banco', async () => {
    montar()

    expect(await screen.findByLabelText(/Duración/)).toHaveValue('30')
    expect(screen.getByLabelText(/Número de preguntas/)).toHaveValue('10')
    expect(screen.getByText('El banco tiene 40 preguntas')).toBeInTheDocument()
    expect(screen.queryByText('Aún no se ha configurado')).not.toBeInTheDocument()
  })

  it('avisa que son los valores por defecto mientras no se configure', async () => {
    mockObtener.mockResolvedValue({ ...config, configurada: false })
    montar()

    expect(await screen.findByText('Aún no se ha configurado')).toBeInTheDocument()
  })

  it('guarda los nuevos valores y lo confirma', async () => {
    mockGuardar.mockResolvedValue({ ...config, duracion_minutos: 45, num_preguntas: 12 })
    montar()
    const duracion = await screen.findByLabelText(/Duración/)
    await userEvent.clear(duracion)
    await userEvent.type(duracion, '45')
    const preguntas = screen.getByLabelText(/Número de preguntas/)
    await userEvent.clear(preguntas)
    await userEvent.type(preguntas, '12')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar configuración' }))

    expect(await screen.findByText('Configuración guardada.')).toBeInTheDocument()
    expect(mockGuardar).toHaveBeenCalledWith(expect.anything(), {
      duracion_minutos: 45,
      num_preguntas: 12,
    })
  })

  it('valida la duración y el número de preguntas antes de enviar', async () => {
    montar()
    const duracion = await screen.findByLabelText(/Duración/)
    await userEvent.clear(duracion)
    await userEvent.type(duracion, '999')
    const preguntas = screen.getByLabelText(/Número de preguntas/)
    await userEvent.clear(preguntas)
    await userEvent.type(preguntas, '0')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar configuración' }))

    expect(screen.getByText('Debe ser un número entero de 1 a 480')).toBeInTheDocument()
    expect(screen.getByText('Debe ser un número entero mayor que cero')).toBeInTheDocument()
    expect(mockGuardar).not.toHaveBeenCalled()
  })

  it('muestra el error del servidor si pide más preguntas de las que hay', async () => {
    mockGuardar.mockRejectedValue(
      new ApiError(422, 'VALIDATION_ERROR', 'El banco no tiene tantas preguntas', {
        num_preguntas: 'El banco tiene 40 preguntas',
      }),
    )
    montar()
    await screen.findByLabelText(/Duración/)
    await userEvent.click(screen.getByRole('button', { name: 'Guardar configuración' }))

    expect(await screen.findByText('El banco no tiene tantas preguntas')).toBeInTheDocument()
    expect(screen.getAllByText('El banco tiene 40 preguntas').length).toBeGreaterThan(0)
  })

  it('enlaza a Fases para el horario del examen', async () => {
    montar()
    await screen.findByLabelText(/Duración/)

    expect(screen.getByRole('link', { name: 'Fases' })).toHaveAttribute('href', '/admin/fases')
  })

  it('permite reintentar si falla la carga', async () => {
    mockObtener.mockRejectedValueOnce(new ApiError(500, 'INTERNAL_ERROR', 'Error interno'))
    montar()

    await userEvent.click(await screen.findByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByLabelText(/Duración/)).toBeInTheDocument()
  })
})

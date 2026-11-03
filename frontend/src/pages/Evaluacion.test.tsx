import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { vi, type Mock } from 'vitest'
import { ApiError } from '@/services/errores'
import {
  finalizarEvaluacion,
  guardarRespuesta,
  iniciarEvaluacion,
  obtenerPregunta,
  type PreguntaExamen,
} from '@/services/evaluacion'
import Evaluacion from './Evaluacion'

vi.mock('@/services/evaluacion', () => ({
  iniciarEvaluacion: vi.fn(),
  obtenerPregunta: vi.fn(),
  guardarRespuesta: vi.fn(),
  finalizarEvaluacion: vi.fn(),
}))
const mockIniciar = iniciarEvaluacion as Mock
const mockPregunta = obtenerPregunta as Mock
const mockResponder = guardarRespuesta as Mock
const mockFinalizar = finalizarEvaluacion as Mock

const TOTAL = 3
const pregunta = (n: number, extra: Partial<PreguntaExamen> = {}): PreguntaExamen => ({
  numero: n,
  total: TOTAL,
  id: n * 10,
  enunciado: `Enunciado ${n}`,
  opciones: [`Opción A${n}`, `Opción B${n}`],
  cierre: '2026-11-03T16:00:00Z',
  tiempo_restante_segundos: 600,
  opcion_elegida: null,
  ...extra,
})
const calificacion = {
  nota: 66.67,
  aciertos: 2,
  total: TOTAL,
  finalizado_at: '2026-11-03T15:20:00Z',
}

function montar() {
  return render(
    <MemoryRouter>
      <Evaluacion />
    </MemoryRouter>,
  )
}

async function iniciar() {
  await userEvent.click(screen.getByRole('button', { name: 'Iniciar Evaluación' }))
  await screen.findByText('Enunciado 1')
}

beforeEach(() => {
  mockIniciar.mockReset().mockResolvedValue({
    total_preguntas: TOTAL,
    inicio: 'x',
    cierre: 'y',
    tiempo_restante_segundos: 600,
  })
  mockPregunta.mockReset().mockImplementation(async (n: number) => pregunta(n))
  mockResponder.mockReset().mockResolvedValue(undefined)
  mockFinalizar.mockReset().mockResolvedValue(calificacion)
})

describe('Evaluacion: instrucciones', () => {
  it('muestra las instrucciones y el botón Iniciar Evaluación', () => {
    montar()

    expect(screen.getByText('Evaluación técnica')).toBeInTheDocument()
    expect(screen.getByText(/botón «Atrás» del navegador/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Iniciar Evaluación' })).toBeInTheDocument()
  })

  it('al iniciar muestra la primera pregunta y el reloj', async () => {
    montar()
    await iniciar()

    expect(screen.getByText('Pregunta 1 de 3')).toBeInTheDocument()
    expect(screen.getByRole('timer')).toHaveTextContent('10:00')
    expect(mockPregunta).toHaveBeenCalledWith(1)
  })

  it('si no está habilitado muestra el motivo del servidor', async () => {
    mockIniciar.mockRejectedValueOnce(
      new ApiError(403, 'FORBIDDEN', 'La evaluación aún no ha comenzado', {
        inicia_en: '2026-11-10T15:00:00Z',
      }),
    )
    montar()
    await userEvent.click(screen.getByRole('button', { name: 'Iniciar Evaluación' }))

    expect(await screen.findByText('La evaluación aún no ha comenzado')).toBeInTheDocument()
    expect(screen.getByText(/Fecha y hora:/)).toBeInTheDocument()
    expect(screen.queryByRole('timer')).not.toBeInTheDocument()
  })

  it('si ya rindió (409) ofrece ver su resultado', async () => {
    mockIniciar.mockRejectedValueOnce(new ApiError(409, 'CONFLICT', 'Ya rindió la evaluación'))
    montar()
    await userEvent.click(screen.getByRole('button', { name: 'Iniciar Evaluación' }))
    await userEvent.click(await screen.findByRole('button', { name: 'Ver mi resultado' }))

    expect(await screen.findByText('Evaluación finalizada')).toBeInTheDocument()
    expect(screen.getByText('66.67 / 100')).toBeInTheDocument()
  })

  it('si no se puede obtener el resultado lo dice en la pantalla de inicio', async () => {
    mockIniciar.mockRejectedValueOnce(new ApiError(409, 'CONFLICT', 'Ya rindió la evaluación'))
    mockFinalizar.mockRejectedValueOnce(new ApiError(0, 'NETWORK_ERROR', 'Sin conexión'))
    montar()
    await userEvent.click(screen.getByRole('button', { name: 'Iniciar Evaluación' }))
    await userEvent.click(await screen.findByRole('button', { name: 'Ver mi resultado' }))

    expect(await screen.findByText('Sin conexión')).toBeInTheDocument()
  })

  it('al retomar un examen en curso trae lo que ya respondió en las demás preguntas', async () => {
    mockIniciar.mockResolvedValueOnce({
      total_preguntas: TOTAL,
      inicio: 'x',
      cierre: 'y',
      tiempo_restante_segundos: 300,
      reanudado: true,
    })
    mockPregunta.mockImplementation(async (n: number) =>
      pregunta(n, { opcion_elegida: n === 3 ? 0 : null }),
    )
    montar()
    await iniciar()

    expect(await screen.findByText('1 de 3 respondidas')).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Ir a la pregunta 3 (respondida)' }),
    ).toBeInTheDocument()
  })
})

describe('Evaluacion: preguntas', () => {
  it('elegir una opción la guarda en el servidor y la deja marcada', async () => {
    montar()
    await iniciar()

    await userEvent.click(screen.getByRole('radio', { name: 'Opción B1' }))

    expect(mockResponder).toHaveBeenCalledWith(1, 1)
    expect(screen.getByRole('radio', { name: 'Opción B1' })).toBeChecked()
    expect(screen.getByText('1 de 3 respondidas')).toBeInTheDocument()
  })

  it('avanza, retrocede y conserva lo respondido', async () => {
    montar()
    await iniciar()
    await userEvent.click(screen.getByRole('radio', { name: 'Opción A1' }))

    await userEvent.click(screen.getByRole('button', { name: 'Siguiente' }))
    expect(await screen.findByText('Enunciado 2')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Anterior' }))

    expect(await screen.findByText('Enunciado 1')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'Opción A1' })).toBeChecked()
  })

  it('muestra marcada la opción que el servidor ya tenía (al volver a una pregunta)', async () => {
    mockPregunta.mockImplementation(async (n: number) =>
      pregunta(n, { opcion_elegida: n === 1 ? 1 : null }),
    )
    montar()
    await iniciar()

    expect(screen.getByRole('radio', { name: 'Opción B1' })).toBeChecked()
  })

  it('el navegador de preguntas salta a la elegida y marca las respondidas', async () => {
    montar()
    await iniciar()
    await userEvent.click(screen.getByRole('radio', { name: 'Opción A1' }))

    expect(
      screen.getByRole('button', { name: 'Ir a la pregunta 1 (respondida)' }),
    ).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Ir a la pregunta 3' }))
    expect(await screen.findByText('Enunciado 3')).toBeInTheDocument()
  })

  it('no hay Anterior en la primera ni Siguiente en la última', async () => {
    montar()
    await iniciar()
    expect(screen.getByRole('button', { name: 'Anterior' })).toBeDisabled()

    await userEvent.click(screen.getByRole('button', { name: 'Ir a la pregunta 3' }))
    await screen.findByText('Enunciado 3')
    expect(screen.getByRole('button', { name: 'Siguiente' })).toBeDisabled()
  })

  it('si no se pudo guardar la respuesta muestra el error y no la marca', async () => {
    mockResponder.mockRejectedValueOnce(new ApiError(500, 'INTERNAL_ERROR', 'Error interno'))
    montar()
    await iniciar()
    await userEvent.click(screen.getByRole('radio', { name: 'Opción A1' }))

    expect(await screen.findByText('Error interno')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'Opción A1' })).not.toBeChecked()
  })

  it('si el servidor dice que el examen terminó (409) se califica y muestra el resultado', async () => {
    mockResponder.mockRejectedValueOnce(new ApiError(409, 'CONFLICT', 'El examen ya terminó'))
    montar()
    await iniciar()
    await userEvent.click(screen.getByRole('radio', { name: 'Opción A1' }))

    expect(await screen.findByText('Evaluación finalizada')).toBeInTheDocument()
    expect(mockFinalizar).toHaveBeenCalledTimes(1)
  })
})

describe('Evaluacion: finalizar', () => {
  it('pide confirmar, avisa de las preguntas sin responder y muestra el resultado', async () => {
    montar()
    await iniciar()
    await userEvent.click(screen.getByRole('radio', { name: 'Opción A1' }))

    await userEvent.click(screen.getByRole('button', { name: 'Finalizar y Calificar' }))
    expect(screen.getByText(/no podrá cambiar sus respuestas/)).toBeInTheDocument()
    expect(screen.getByText(/faltan 2 preguntas por responder/)).toBeInTheDocument()
    expect(mockFinalizar).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: 'Finalizar' }))

    expect(await screen.findByText('Evaluación finalizada')).toBeInTheDocument()
    expect(screen.getByText('66.67 / 100')).toBeInTheDocument()
    expect(screen.getByText('2 aciertos de 3 preguntas')).toBeInTheDocument()
    expect(screen.getByText(/llegó a su correo/)).toBeInTheDocument()
    expect(screen.queryByRole('timer')).not.toBeInTheDocument()
  })

  it('cancelar la confirmación deja seguir el examen', async () => {
    montar()
    await iniciar()
    await userEvent.click(screen.getByRole('button', { name: 'Finalizar y Calificar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }))

    expect(mockFinalizar).not.toHaveBeenCalled()
    expect(screen.getByRole('timer')).toBeInTheDocument()
  })

  it('si falla la calificación muestra el error y se puede reintentar', async () => {
    mockFinalizar.mockRejectedValueOnce(new ApiError(500, 'INTERNAL_ERROR', 'Error interno'))
    montar()
    await iniciar()
    await userEvent.click(screen.getByRole('button', { name: 'Finalizar y Calificar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Finalizar' }))

    expect(await screen.findByText('Error interno')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Finalizar y Calificar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Finalizar' }))
    expect(await screen.findByText('Evaluación finalizada')).toBeInTheDocument()
  })

  it('espera a que se guarde la última respuesta antes de calificar', async () => {
    let terminarGuardado = () => {}
    mockResponder.mockReturnValueOnce(
      new Promise<void>((resolver) => {
        terminarGuardado = resolver
      }),
    )
    montar()
    await iniciar()
    await userEvent.click(screen.getByRole('radio', { name: 'Opción A1' }))
    await userEvent.click(screen.getByRole('button', { name: 'Finalizar y Calificar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Finalizar' }))

    expect(mockFinalizar).not.toHaveBeenCalled()
    await act(async () => terminarGuardado())

    expect(await screen.findByText('Evaluación finalizada')).toBeInTheDocument()
    expect(mockFinalizar).toHaveBeenCalledTimes(1)
  })
})

describe('Evaluacion: reloj y cierre automático', () => {
  beforeEach(() => vi.useFakeTimers({ shouldAdvanceTime: true }))
  afterEach(() => vi.useRealTimers())

  it('cuenta hacia atrás y al llegar a 00:00 se cierra y califica solo', async () => {
    mockIniciar.mockResolvedValue({
      total_preguntas: TOTAL,
      inicio: 'x',
      cierre: 'y',
      tiempo_restante_segundos: 5,
    })
    mockPregunta.mockImplementation(async (n: number) =>
      pregunta(n, { tiempo_restante_segundos: 5 }),
    )
    const usuario = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
    montar()
    await usuario.click(screen.getByRole('button', { name: 'Iniciar Evaluación' }))
    await screen.findByText('Enunciado 1')
    expect(screen.getByRole('timer')).toHaveTextContent('00:05')

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000)
    })
    expect(screen.getByRole('timer')).toHaveTextContent('00:03')
    expect(mockFinalizar).not.toHaveBeenCalled()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(4000)
    })
    expect(await screen.findByText('Evaluación finalizada')).toBeInTheDocument()
    expect(mockFinalizar).toHaveBeenCalledTimes(1)
  })
})

describe('Evaluacion: bloqueos durante el examen', () => {
  it('bloquea pegar mientras dura el examen y lo libera al terminar', async () => {
    montar()
    await iniciar()
    const durante = new Event('paste', { cancelable: true, bubbles: true })
    document.body.dispatchEvent(durante)
    expect(durante.defaultPrevented).toBe(true)

    await userEvent.click(screen.getByRole('button', { name: 'Finalizar y Calificar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Finalizar' }))
    await screen.findByText('Evaluación finalizada')

    const despues = new Event('paste', { cancelable: true, bubbles: true })
    document.body.dispatchEvent(despues)
    expect(despues.defaultPrevented).toBe(false)
  })
})

import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import {
  actualizarPregunta,
  crearPregunta,
  eliminarPregunta,
  listarPreguntas,
  type PaginaPreguntas,
} from '@/services/adminEvaluacion'
import { ApiError } from '@/services/errores'
import EvaluacionBanco from './EvaluacionBanco'

vi.mock('@/services/adminEvaluacion', async (original) => ({
  ...(await original<typeof import('@/services/adminEvaluacion')>()),
  listarPreguntas: vi.fn(),
  crearPregunta: vi.fn(),
  actualizarPregunta: vi.fn(),
  eliminarPregunta: vi.fn(),
  cargarPreguntasCsv: vi.fn(),
}))
const mockListar = listarPreguntas as Mock
const mockCrear = crearPregunta as Mock
const mockActualizar = actualizarPregunta as Mock
const mockEliminar = eliminarPregunta as Mock

const pagina = (total = 2): PaginaPreguntas => ({
  items: [
    {
      id: 1,
      enunciado: '¿Capital de Ecuador?',
      opciones: ['Guayaquil', 'Quito'],
      respuesta_correcta: 1,
    },
    {
      id: 2,
      enunciado: '¿Moneda de Ecuador?',
      opciones: ['Sucre', 'Dólar'],
      respuesta_correcta: 1,
    },
  ],
  total,
  pagina: 1,
  por_pagina: 20,
})

beforeEach(() => {
  mockListar.mockReset().mockResolvedValue(pagina())
  mockCrear.mockReset()
  mockActualizar.mockReset()
  mockEliminar.mockReset()
})

describe('EvaluacionBanco', () => {
  it('lista las preguntas con su respuesta correcta', async () => {
    render(<EvaluacionBanco />)

    expect(await screen.findByText('¿Capital de Ecuador?')).toBeInTheDocument()
    expect(screen.getByText('2 preguntas')).toBeInTheDocument()
    expect(screen.getByText('Quito')).toBeInTheDocument()
    expect(mockListar).toHaveBeenCalledWith(expect.anything(), {
      buscar: '',
      pagina: 1,
      por_pagina: 20,
    })
  })

  it('busca por el texto del enunciado', async () => {
    render(<EvaluacionBanco />)
    await screen.findByText('¿Capital de Ecuador?')

    await userEvent.type(screen.getByLabelText('Buscar en el enunciado'), ' moneda ')
    await userEvent.click(screen.getByRole('button', { name: 'Buscar' }))

    await screen.findByText('2 preguntas')
    expect(mockListar).toHaveBeenLastCalledWith(expect.anything(), {
      buscar: 'moneda',
      pagina: 1,
      por_pagina: 20,
    })
  })

  it('muestra el banco vacío con una invitación a cargar preguntas', async () => {
    mockListar.mockResolvedValue({ items: [], total: 0, pagina: 1, por_pagina: 20 })
    render(<EvaluacionBanco />)

    expect(await screen.findByText(/El banco está vacío/)).toBeInTheDocument()
  })

  it('muestra el error si no se puede consultar (por ejemplo, 403 al analista)', async () => {
    mockListar.mockRejectedValue(
      new ApiError(403, 'FORBIDDEN', 'No tiene permiso para esta acción'),
    )
    render(<EvaluacionBanco />)

    expect(await screen.findByText('No tiene permiso para esta acción')).toBeInTheDocument()
  })

  it('agrega una pregunta nueva y recarga el listado', async () => {
    mockCrear.mockResolvedValue({ id: 3 })
    render(<EvaluacionBanco />)
    await screen.findByText('¿Capital de Ecuador?')

    await userEvent.click(screen.getByRole('button', { name: 'Nueva pregunta' }))
    await userEvent.type(screen.getByLabelText(/Enunciado/), '¿Nueva?')
    await userEvent.type(screen.getByRole('textbox', { name: 'Opción 1' }), 'Sí')
    await userEvent.type(screen.getByRole('textbox', { name: 'Opción 2' }), 'No')
    await userEvent.click(screen.getByRole('radio', { name: 'Opción 1 es la correcta' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar pregunta' }))

    expect(await screen.findByText('Pregunta agregada al banco.')).toBeInTheDocument()
    expect(mockCrear).toHaveBeenCalledWith(expect.anything(), {
      enunciado: '¿Nueva?',
      opciones: ['Sí', 'No'],
      respuesta_correcta: 0,
    })
    expect(mockListar).toHaveBeenCalledTimes(2)
  })

  it('edita una pregunta existente', async () => {
    mockActualizar.mockResolvedValue({ id: 1 })
    render(<EvaluacionBanco />)
    await screen.findByText('¿Capital de Ecuador?')

    await userEvent.click(screen.getByRole('button', { name: 'Editar la pregunta 1' }))
    const enunciado = screen.getByLabelText(/Enunciado/)
    await userEvent.clear(enunciado)
    await userEvent.type(enunciado, '¿Capital?')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar pregunta' }))

    expect(await screen.findByText('Pregunta actualizada.')).toBeInTheDocument()
    expect(mockActualizar).toHaveBeenCalledWith(expect.anything(), 1, {
      enunciado: '¿Capital?',
      opciones: ['Guayaquil', 'Quito'],
      respuesta_correcta: 1,
    })
  })

  it('si el servidor rechaza el guardado (409) deja abierto el formulario con el mensaje', async () => {
    mockActualizar.mockRejectedValue(new ApiError(409, 'CONFLICT', 'Ya hay evaluaciones iniciadas'))
    render(<EvaluacionBanco />)
    await screen.findByText('¿Capital de Ecuador?')

    await userEvent.click(screen.getByRole('button', { name: 'Editar la pregunta 1' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar pregunta' }))

    expect(await screen.findByText('Ya hay evaluaciones iniciadas')).toBeInTheDocument()
    expect(screen.getByText('Editar pregunta')).toBeInTheDocument()
  })

  it('pide confirmar antes de eliminar y recarga el listado', async () => {
    mockEliminar.mockResolvedValue(undefined)
    render(<EvaluacionBanco />)
    await screen.findByText('¿Capital de Ecuador?')

    await userEvent.click(screen.getByRole('button', { name: 'Eliminar la pregunta 1' }))
    expect(screen.getByText(/No se puede deshacer/)).toBeInTheDocument()
    expect(mockEliminar).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('Pregunta eliminada.')).toBeInTheDocument()
    expect(mockEliminar).toHaveBeenCalledWith(expect.anything(), 1)
    expect(mockListar).toHaveBeenCalledTimes(2)
  })

  it('si no se puede eliminar (409) muestra el motivo dentro de la confirmación', async () => {
    mockEliminar.mockRejectedValue(
      new ApiError(409, 'CONFLICT', 'El banco quedaría con menos preguntas de las configuradas'),
    )
    render(<EvaluacionBanco />)
    await screen.findByText('¿Capital de Ecuador?')

    await userEvent.click(screen.getByRole('button', { name: 'Eliminar la pregunta 1' }))
    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))

    expect(
      await screen.findByText('El banco quedaría con menos preguntas de las configuradas'),
    ).toBeInTheDocument()
  })

  it('abre la carga masiva', async () => {
    render(<EvaluacionBanco />)
    await screen.findByText('¿Capital de Ecuador?')

    await userEvent.click(screen.getByRole('button', { name: 'Carga masiva' }))

    expect(screen.getByText('Carga masiva desde archivo')).toBeInTheDocument()
  })

  it('pagina el listado', async () => {
    mockListar.mockResolvedValue({ ...pagina(45), pagina: 1 })
    render(<EvaluacionBanco />)
    await screen.findByText('45 preguntas')

    await userEvent.click(screen.getByRole('button', { name: 'Siguiente' }))

    await screen.findByText('45 preguntas')
    expect(mockListar).toHaveBeenLastCalledWith(expect.anything(), {
      buscar: '',
      pagina: 2,
      por_pagina: 20,
    })
  })

  it('al eliminar la última pregunta de la última página vuelve a la página anterior', async () => {
    mockEliminar.mockResolvedValue(undefined)
    const segunda = { ...pagina(21), pagina: 2, items: [pagina().items[0]] }
    mockListar
      .mockResolvedValueOnce(pagina(21))
      .mockResolvedValueOnce(segunda)
      .mockResolvedValueOnce({ ...segunda, total: 20, items: [] })
      .mockResolvedValueOnce(pagina(20))
    render(<EvaluacionBanco />)
    await screen.findByText('21 preguntas')
    await userEvent.click(screen.getByRole('button', { name: 'Siguiente' }))
    await screen.findByText('Página 2 de 2')

    await userEvent.click(screen.getByRole('button', { name: 'Eliminar la pregunta 1' }))
    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))

    expect(await screen.findByText('20 preguntas')).toBeInTheDocument()
    expect(screen.getByText('Página 1 de 1')).toBeInTheDocument()
    expect(screen.queryByText(/El banco está vacío/)).not.toBeInTheDocument()
  })
})

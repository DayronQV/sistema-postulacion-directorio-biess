import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { cargarPreguntasCsv } from '@/services/adminEvaluacion'
import { ApiError } from '@/services/errores'
import CargaMasiva from './CargaMasiva'

vi.mock('@/services/adminEvaluacion', async (original) => ({
  ...(await original<typeof import('@/services/adminEvaluacion')>()),
  cargarPreguntasCsv: vi.fn(),
}))
const mockCargar = cargarPreguntasCsv as Mock

const CABECERA = 'enunciado,opciones,respuesta_correcta\n'

function csv(contenido: string, nombre = 'preguntas.csv') {
  return new File([contenido], nombre, { type: 'text/csv' })
}

async function elegir(archivo: File) {
  await userEvent.upload(screen.getByLabelText('Archivo CSV'), archivo)
}

beforeEach(() => mockCargar.mockReset())

describe('CargaMasiva', () => {
  it('muestra la vista previa y deja cargar si todas las filas son válidas', async () => {
    const onCargado = vi.fn()
    mockCargar.mockResolvedValue({ creadas: 2 })
    render(<CargaMasiva onCargado={onCargado} />)
    await elegir(csv(`${CABECERA}"¿Capital?",Guayaquil|Quito,1\n"¿Moneda?",Dólar|Sucre,0\n`))

    expect(
      await screen.findByText(/2 preguntas en el archivo; todas son válidas/),
    ).toBeInTheDocument()
    expect(screen.getAllByText('Válida')).toHaveLength(2)

    await userEvent.click(screen.getByRole('button', { name: /Cargar 2 preguntas/ }))

    expect(mockCargar).toHaveBeenCalledWith(expect.anything(), expect.any(File))
    expect(await screen.findByText(/Se cargaron 2 preguntas al banco/)).toBeInTheDocument()
    expect(onCargado).toHaveBeenCalled()
  })

  it('con filas inválidas muestra el error por línea y no deja cargar', async () => {
    render(<CargaMasiva onCargado={vi.fn()} />)
    await elegir(csv(`${CABECERA}"¿Una?",Sola,0\n"¿Dos?",A|B,1\n`))

    expect(await screen.findByText(/1 con errores/)).toBeInTheDocument()
    expect(screen.getByText('Debe tener de 2 a 8 opciones')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Cargar 2 preguntas/ })).toBeDisabled()
    expect(mockCargar).not.toHaveBeenCalled()
  })

  it('avisa si el archivo no tiene las columnas', async () => {
    render(<CargaMasiva onCargado={vi.fn()} />)
    await elegir(csv('a,b\n1,2\n'))

    expect(await screen.findByText(/Falta la columna enunciado, opciones/)).toBeInTheDocument()
  })

  it('rechaza un archivo de más de 1 MB sin leerlo', async () => {
    render(<CargaMasiva onCargado={vi.fn()} />)
    await elegir(csv('x'.repeat(1024 * 1024 + 1)))

    expect(await screen.findByText('El archivo supera 1 MB')).toBeInTheDocument()
  })

  it('si el servidor rechaza la carga, lista los errores por línea', async () => {
    mockCargar.mockRejectedValueOnce(
      new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos', {
        'filas[2].respuesta_correcta': 'Fuera de las opciones',
      }),
    )
    const onCargado = vi.fn()
    render(<CargaMasiva onCargado={onCargado} />)
    await elegir(csv(`${CABECERA}"¿Capital?",Guayaquil|Quito,1\n`))
    await userEvent.click(await screen.findByRole('button', { name: /Cargar 1 preguntas/ }))

    expect(await screen.findByText('Datos inválidos')).toBeInTheDocument()
    expect(
      screen.getByText('Línea 2 (respuesta_correcta): Fuera de las opciones'),
    ).toBeInTheDocument()
    expect(onCargado).not.toHaveBeenCalled()
  })

  it('Quitar archivo limpia la vista previa', async () => {
    render(<CargaMasiva onCargado={vi.fn()} />)
    await elegir(csv(`${CABECERA}"¿Capital?",Guayaquil|Quito,1\n`))
    await userEvent.click(await screen.findByRole('button', { name: 'Quitar archivo' }))

    await waitFor(() => expect(screen.queryByText(/en el archivo/)).not.toBeInTheDocument())
  })

  it('ofrece descargar una plantilla', () => {
    render(<CargaMasiva onCargado={vi.fn()} />)
    const enlace = screen.getByRole('link', { name: 'Descargar plantilla' })
    expect(enlace).toHaveAttribute('download', 'plantilla-preguntas.csv')
    expect(decodeURIComponent(enlace.getAttribute('href') ?? '')).toContain(
      'enunciado,opciones,respuesta_correcta',
    )
  })
})

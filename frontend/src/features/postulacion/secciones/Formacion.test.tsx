import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { eliminarDocumento, type DocumentoSubido } from '@/services/documentos'
import { ApiError } from '@/services/errores'
import type { FormacionAcademica } from '@/services/postulacion'
import Formacion from './Formacion'

vi.mock('@/services/documentos', () => ({ eliminarDocumento: vi.fn() }))

// Cada clic en "Subir PDF" entrega un documento nuevo (doc-1, doc-2...). También deja ver
// si la sección pide que DocumentoUpload no borre en el servidor.
let siguiente = 0
const propsUpload: { eliminarEnServidor?: boolean }[] = []
vi.mock('@/components/form/DocumentoUpload', () => ({
  default: (props: {
    onChange: (documento: DocumentoSubido | null) => void
    eliminarEnServidor?: boolean
  }) => {
    propsUpload.push(props)
    return (
      <>
        <button
          type="button"
          onClick={() => {
            siguiente += 1
            props.onChange({
              documento_id: `doc-${siguiente}`,
              nombre: 'titulo.pdf',
              tamano_bytes: 1024,
            })
          }}
        >
          Subir PDF
        </button>
        <button type="button" onClick={() => props.onChange(null)}>
          Quitar PDF
        </button>
      </>
    )
  },
}))

const mockEliminar = eliminarDocumento as Mock

const maestria: FormacionAcademica = {
  nivel: 'maestria',
  institucion: 'UPSE',
  tipo: 'nacional',
  titulo: 'Magíster en Finanzas',
  area: 'finanzas',
  senescyt_numero: 'REG-123',
  senescyt_fecha: '2025-10-15',
  documento_id: 'doc-guardado',
}

// Pega el texto de una vez: tipear carácter por carácter vuelve lentas estas pruebas.
async function escribir(etiqueta: RegExp, texto: string) {
  const campo = screen.getByLabelText(etiqueta)
  await userEvent.clear(campo)
  await userEvent.click(campo)
  await userEvent.paste(texto)
}

async function llenarFormacion() {
  await userEvent.selectOptions(screen.getByLabelText(/Nivel de instrucción/), 'maestria')
  await escribir(/Institución educativa/, 'UPSE')
  await userEvent.selectOptions(screen.getByLabelText(/Tipo de institución/), 'nacional')
  await escribir(/Título obtenido/, 'Magíster en Finanzas')
  await userEvent.selectOptions(screen.getByLabelText(/Área de estudios/), 'finanzas')
  await escribir(/Número de registro SENESCYT/, 'REG-123')
  fireEvent.change(screen.getByLabelText(/Fecha de registro SENESCYT/), {
    target: { value: '2025-10-15' },
  })
}

const guardarInformacion = () => screen.queryByRole('button', { name: 'Guardar información' })

beforeEach(() => {
  siguiente = 0
  propsUpload.length = 0
  mockEliminar.mockReset()
  mockEliminar.mockResolvedValue(undefined)
})

describe('Formacion', () => {
  it('agrega una formación con respaldo y la guarda al momento', async () => {
    const guardar = vi.fn()
    render(<Formacion valor={[]} onGuardar={guardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Agregar formación' }))
    await llenarFormacion()
    await userEvent.click(screen.getByRole('button', { name: 'Subir PDF' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(guardar).toHaveBeenCalledWith([{ ...maestria, documento_id: 'doc-1' }])
    expect(await screen.findByText('Magíster en Finanzas')).toBeInTheDocument()
    expect(mockEliminar).not.toHaveBeenCalled()
    // Ya quedó guardada: no hay nada pendiente ni formulario abierto.
    expect(guardarInformacion()).not.toBeInTheDocument()
    expect(screen.queryByLabelText(/Nivel de instrucción/)).not.toBeInTheDocument()
  })

  it('con el formulario abierto solo hay Cancelar y Guardar', async () => {
    render(<Formacion valor={[maestria]} onGuardar={() => undefined} />)

    await userEvent.click(screen.getByRole('button', { name: 'Agregar formación' }))
    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeInTheDocument()
    for (const sobrante of ['Agregar', 'Actualizar', 'Guardar información']) {
      expect(screen.queryByRole('button', { name: sobrante })).not.toBeInTheDocument()
    }
  })

  it('agregar conserva las formaciones que ya estaban guardadas', async () => {
    const guardar = vi.fn()
    render(<Formacion valor={[maestria]} onGuardar={guardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Agregar formación' }))
    await llenarFormacion()
    await escribir(/Título obtenido/, 'Doctor en Economía')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    // La fila nueva no tiene respaldo (documento_id vacío); la que ya estaba conserva el suyo.
    expect(guardar).toHaveBeenCalledWith([
      maestria,
      { ...maestria, titulo: 'Doctor en Economía', documento_id: '' },
    ])
  })

  it('con algo pendiente, abrir el formulario oculta Guardar información', async () => {
    render(<Formacion valor={[maestria]} onGuardar={() => undefined} />)
    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))
    expect(guardarInformacion()).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Agregar formación' }))
    expect(guardarInformacion()).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }))
    expect(guardarInformacion()).toBeInTheDocument()
  })

  it('sin cambios pendientes no ofrece Guardar información', () => {
    render(<Formacion valor={[]} onGuardar={() => undefined} />)
    expect(guardarInformacion()).not.toBeInTheDocument()
  })

  it('el respaldo es opcional', async () => {
    const guardar = vi.fn()
    render(<Formacion valor={[]} onGuardar={guardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Agregar formación' }))
    await llenarFormacion()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(guardar).toHaveBeenCalledWith([{ ...maestria, documento_id: '' }])
  })

  it('edita una formación guardada y la guarda al momento', async () => {
    const guardar = vi.fn()
    render(<Formacion valor={[maestria]} onGuardar={guardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await escribir(/Título obtenido/, 'Maestría en Economía')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(guardar).toHaveBeenCalledWith([{ ...maestria, titulo: 'Maestría en Economía' }])
    expect(await screen.findByText('Maestría en Economía')).toBeInTheDocument()
  })

  it('al eliminar una fila queda pendiente y Guardar información la guarda', async () => {
    const guardar = vi.fn()
    render(<Formacion valor={[maestria]} onGuardar={guardar} />)
    expect(guardarInformacion()).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))
    expect(screen.getByText('Aún no hay registros.')).toBeInTheDocument()
    expect(guardar).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(guardar).toHaveBeenCalledWith([])
    await waitFor(() => expect(guardarInformacion()).not.toBeInTheDocument())
  })

  it('quitar el PDF de una formación guardada no lo borra en el servidor (seguiría en uso)', async () => {
    const guardar = vi.fn()
    render(<Formacion valor={[maestria]} onGuardar={guardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    expect(propsUpload.at(-1)?.eliminarEnServidor).toBe(false)
    await userEvent.click(screen.getByRole('button', { name: 'Quitar PDF' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(guardar).toHaveBeenCalledWith([{ ...maestria, documento_id: '' }])
    expect(mockEliminar).not.toHaveBeenCalled()
  })

  it('al guardar borra los PDF subidos que ninguna fila usa', async () => {
    const guardar = vi.fn()
    render(<Formacion valor={[maestria]} onGuardar={guardar} />)

    // Sube doc-1 y lo reemplaza por doc-2 antes de guardar: doc-1 queda sin usar.
    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Subir PDF' }))
    await userEvent.click(screen.getByRole('button', { name: 'Subir PDF' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(guardar).toHaveBeenCalledWith([{ ...maestria, documento_id: 'doc-2' }])
    await waitFor(() => expect(mockEliminar).toHaveBeenCalledTimes(1))
    expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'doc-1')
  })

  it('al salir sin guardar borra lo subido en la pantalla', async () => {
    const { unmount } = render(<Formacion valor={[]} onGuardar={() => undefined} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar formación' }))
    await userEvent.selectOptions(screen.getByLabelText(/Nivel de instrucción/), 'tercer_nivel')
    await userEvent.click(screen.getByRole('button', { name: 'Subir PDF' }))
    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }))

    unmount()
    expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'doc-1')
  })

  it('exige nivel y campos antes de guardar', async () => {
    const guardar = vi.fn()
    render(<Formacion valor={[]} onGuardar={guardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar formación' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(screen.getByText('Seleccione el nivel de instrucción')).toBeInTheDocument()

    await userEvent.selectOptions(screen.getByLabelText(/Nivel de instrucción/), 'tercer_nivel')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(screen.getByText('Ingrese la fecha de registro')).toBeInTheDocument()
    expect(screen.getByText('Seleccione el área de estudios')).toBeInTheDocument()
    expect(guardar).not.toHaveBeenCalled()
  })

  it('si el servidor rechaza el guardado, muestra el error y deja el formulario abierto', async () => {
    const guardar = vi.fn().mockRejectedValue(
      new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos', {
        'formacion[0].documento_id': 'Referencia no válida',
      }),
    )
    render(<Formacion valor={[maestria]} onGuardar={guardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await escribir(/Título obtenido/, 'Maestría en Economía')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    const alerta = await screen.findByRole('alert')
    expect(alerta).toHaveTextContent('Datos inválidos')
    expect(alerta).toHaveTextContent('formacion[0].documento_id: Referencia no válida')
    // Lo escrito sigue en el formulario y la lista no cambió.
    expect(screen.getByLabelText(/Título obtenido/)).toHaveValue('Maestría en Economía')
    expect(screen.getByText('Magíster en Finanzas')).toBeInTheDocument()
  })
})

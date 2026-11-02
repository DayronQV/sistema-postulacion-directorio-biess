import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { eliminarDocumento, subirDocumento } from '@/services/documentos'
import { ApiError } from '@/services/errores'
import type { ResultadoCargo, TipoCargo } from '@/services/postulacion'
import { ETIQUETA_CARGO, MSG_CARGO_VALIDO, PREGUNTA_RESPALDO } from '../reglasCargo'
import Cargo, { type CargoInicial, type CargoProps } from './Cargo'

vi.mock('@/services/documentos', () => ({
  subirDocumento: vi.fn(),
  descargarDocumento: vi.fn(),
  eliminarDocumento: vi.fn(),
}))

const mockSubir = subirDocumento as Mock
const mockEliminar = eliminarDocumento as Mock

const sinNada: CargoInicial = {
  cargo: '',
  respaldo_organizacion: { nombre: '', documento_id: null },
}

const elegible = (cargo: TipoCargo): ResultadoCargo => ({
  cargo,
  elegible: true,
  mensaje: MSG_CARGO_VALIDO[cargo],
})

function montar(props: Partial<CargoProps> = {}) {
  const onValidar = vi.fn((cargo: TipoCargo) => Promise.resolve(elegible(cargo)))
  const onGuardar = vi.fn(() => Promise.resolve())
  const todas: CargoProps = { valoresIniciales: sinNada, onValidar, onGuardar, ...props }
  render(<Cargo {...todas} />)
  return { onValidar: todas.onValidar as Mock, onGuardar: todas.onGuardar as Mock }
}

const opcion = (cargo: TipoCargo) => screen.getByRole('radio', { name: ETIQUETA_CARGO[cargo] })

function elegirPdf() {
  const entrada = document.querySelector('input[type="file"]') as HTMLInputElement
  return userEvent.upload(entrada, new File(['%PDF'], 'respaldo.pdf', { type: 'application/pdf' }))
}

beforeEach(() => {
  mockSubir.mockReset()
  mockEliminar.mockReset()
  mockEliminar.mockResolvedValue(undefined)
})

describe('<Cargo />', () => {
  it('ofrece una opción única entre los tres cargos, con su texto completo y sin elegir', () => {
    montar()

    for (const cargo of ['afiliados', 'jubilados', 'empleador'] as const) {
      expect(opcion(cargo)).not.toBeChecked()
    }
    expect(screen.getAllByRole('radio', { name: /Miembro Principal y Alterno/ })).toHaveLength(3)
    expect(screen.getByRole('button', { name: 'Guardar Postulación' })).toBeInTheDocument()
    // Hasta validar el cargo no se pregunta por la organización.
    expect(screen.queryByText(/respaldadas por/)).not.toBeInTheDocument()
  })

  it.each([
    ['afiliados', 'Postulante afiliado al IESS', 'organización de trabajadores'],
    ['jubilados', 'Postulante acredita jubilación', 'organización de jubilados'],
    ['empleador', 'Postulante es empleador', 'federación de cámara'],
  ] as const)(
    'al elegir %s valida el cargo, muestra «%s» y pregunta por %s',
    async (cargo, mensaje, organizacion) => {
      const { onValidar } = montar()

      await userEvent.click(opcion(cargo))

      expect(await screen.findByText(mensaje)).toBeInTheDocument()
      expect(onValidar).toHaveBeenCalledWith(cargo)
      expect(opcion(cargo)).toBeChecked()
      expect(screen.getByText(PREGUNTA_RESPALDO[cargo])).toHaveTextContent(organizacion)
    },
  )

  it('mantiene una sola opción elegida al cambiar de cargo', async () => {
    montar()

    await userEvent.click(opcion('afiliados'))
    await screen.findByText('Postulante afiliado al IESS')
    await userEvent.click(opcion('empleador'))

    expect(await screen.findByText('Postulante es empleador')).toBeInTheDocument()
    expect(opcion('empleador')).toBeChecked()
    expect(opcion('afiliados')).not.toBeChecked()
    expect(screen.queryByText('Postulante afiliado al IESS')).not.toBeInTheDocument()
  })

  it('muestra un estado de carga y bloquea el guardado mientras valida', async () => {
    montar({ onValidar: () => new Promise(() => {}) })

    await userEvent.click(opcion('afiliados'))

    expect(await screen.findByText(/Validando con el Historial Laboral/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Guardar Postulación' })).toBeDisabled()
  })

  it('ignora la respuesta de una validación vieja si se eligió otro cargo', async () => {
    let resolverAfiliados: (r: ResultadoCargo) => void = () => {}
    const onValidar = vi.fn((cargo: TipoCargo) =>
      cargo === 'afiliados'
        ? new Promise<ResultadoCargo>((resolver) => (resolverAfiliados = resolver))
        : Promise.resolve(elegible(cargo)),
    )
    montar({ onValidar })

    await userEvent.click(opcion('afiliados'))
    await userEvent.click(opcion('jubilados'))
    await screen.findByText('Postulante acredita jubilación')
    resolverAfiliados(elegible('afiliados'))

    await waitFor(() => expect(onValidar).toHaveBeenCalledTimes(2))
    expect(screen.queryByText('Postulante afiliado al IESS')).not.toBeInTheDocument()
    expect(screen.getByText('Postulante acredita jubilación')).toBeInTheDocument()
  })

  it('si la validación falla muestra el mensaje, no pregunta por la organización y permite reintentar', async () => {
    const onValidar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL_ERROR', 'Error interno del servidor'))
      .mockResolvedValueOnce(elegible('afiliados'))
    montar({ onValidar })

    await userEvent.click(opcion('afiliados'))

    expect(await screen.findByText('Error interno del servidor')).toBeInTheDocument()
    expect(screen.queryByText(/respaldadas por/)).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Reintentar validación' }))

    expect(await screen.findByText('Postulante afiliado al IESS')).toBeInTheDocument()
    expect(screen.queryByText('Error interno del servidor')).not.toBeInTheDocument()
  })

  it('no deja guardar sin cargo validado', async () => {
    const { onGuardar } = montar()

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(screen.getByText(/Seleccione un cargo/)).toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('exige responder la pregunta de respaldo', async () => {
    const { onGuardar } = montar()
    await userEvent.click(opcion('afiliados'))
    await screen.findByText('Postulante afiliado al IESS')

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(screen.getByText('Seleccione Sí o No')).toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('el nombre y el documento aparecen solo si responde Sí', async () => {
    montar()
    await userEvent.click(opcion('afiliados'))
    await screen.findByText('Postulante afiliado al IESS')
    expect(screen.queryByLabelText(/Nombre de Organización/)).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('radio', { name: 'Sí' }))
    expect(screen.getByLabelText(/Nombre de Organización/)).toBeInTheDocument()
    expect(screen.getByText('Documento de respaldo *')).toBeInTheDocument()
    expect(screen.getByText(/máx\. 4 MB/)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('radio', { name: 'No' }))
    expect(screen.queryByLabelText(/Nombre de Organización/)).not.toBeInTheDocument()
  })

  it('con respaldo Sí exige el nombre y el documento antes de guardar', async () => {
    const { onGuardar } = montar()
    await userEvent.click(opcion('jubilados'))
    await screen.findByText('Postulante acredita jubilación')
    await userEvent.click(screen.getByRole('radio', { name: 'Sí' }))

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(screen.getByText('Campo obligatorio')).toBeInTheDocument()
    expect(screen.getByText('Cargue el documento de respaldo')).toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('guarda el cargo con la organización y el documento subido', async () => {
    mockSubir.mockResolvedValue({
      documento_id: 'doc-1',
      nombre: 'respaldo.pdf',
      tamano_bytes: 4,
    })
    const { onGuardar } = montar()
    await userEvent.click(opcion('empleador'))
    await screen.findByText('Postulante es empleador')
    await userEvent.click(screen.getByRole('radio', { name: 'Sí' }))
    await userEvent.type(screen.getByLabelText(/Nombre de Organización/), '  Cámara de Comercio  ')
    await elegirPdf()
    expect(await screen.findByText('respaldo.pdf')).toBeInTheDocument()
    expect(mockSubir).toHaveBeenCalledWith(
      expect.any(AbortSignal),
      'organizacion',
      expect.any(File),
      expect.any(Function),
    )

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(await screen.findByText('La postulación se guardó correctamente.')).toBeInTheDocument()
    expect(onGuardar).toHaveBeenCalledWith({
      cargo: 'empleador',
      respaldo_organizacion: { nombre: 'Cámara de Comercio', documento_id: 'doc-1' },
    })
  })

  it('con respaldo No guarda el cargo sin organización', async () => {
    const { onGuardar } = montar()
    await userEvent.click(opcion('afiliados'))
    await screen.findByText('Postulante afiliado al IESS')
    await userEvent.click(screen.getByRole('radio', { name: 'No' }))

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    await waitFor(() =>
      expect(onGuardar).toHaveBeenCalledWith({
        cargo: 'afiliados',
        respaldo_organizacion: { nombre: '', documento_id: null },
      }),
    )
  })

  it('parte de lo ya guardado: cargo, organización y documento', async () => {
    const { onGuardar, onValidar } = montar({
      valoresIniciales: {
        cargo: 'jubilados',
        respaldo_organizacion: { nombre: 'Asociación Nacional', documento_id: 'doc-9' },
      },
    })

    expect(opcion('jubilados')).toBeChecked()
    expect(screen.getByText('Postulante acredita jubilación')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'Sí' })).toBeChecked()
    expect(screen.getByLabelText(/Nombre de Organización/)).toHaveValue('Asociación Nacional')
    // Abrir la sección no consulta el Historial Laboral: lo hace el guardado.
    expect(onValidar).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    await waitFor(() =>
      expect(onGuardar).toHaveBeenCalledWith({
        cargo: 'jubilados',
        respaldo_organizacion: { nombre: 'Asociación Nacional', documento_id: 'doc-9' },
      }),
    )
  })

  it('un cargo guardado sin organización parte con la respuesta No', () => {
    montar({ valoresIniciales: { ...sinNada, cargo: 'afiliados' } })

    expect(screen.getByRole('radio', { name: 'No' })).toBeChecked()
  })

  it('muestra el error del guardado con los campos que rechazó el servidor', async () => {
    const onGuardar = vi.fn().mockRejectedValue(
      new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos', {
        'respaldo_organizacion.documento_id': 'Referencia no válida',
      }),
    )
    montar({ onGuardar })
    await userEvent.click(opcion('afiliados'))
    await screen.findByText('Postulante afiliado al IESS')
    await userEvent.click(screen.getByRole('radio', { name: 'No' }))

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(await screen.findByText('Datos inválidos')).toBeInTheDocument()
    expect(screen.getByText('Documento de respaldo: Referencia no válida')).toBeInTheDocument()
    expect(screen.queryByText('La postulación se guardó correctamente.')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Guardar Postulación' })).toBeEnabled()
  })

  it('al cambiar a No borra el PDF recién subido para que no quede huérfano', async () => {
    mockSubir.mockResolvedValue({ documento_id: 'doc-1', nombre: 'respaldo.pdf', tamano_bytes: 4 })
    montar()
    await userEvent.click(opcion('afiliados'))
    await screen.findByText('Postulante afiliado al IESS')
    await userEvent.click(screen.getByRole('radio', { name: 'Sí' }))
    await elegirPdf()
    await screen.findByText('respaldo.pdf')

    await userEvent.click(screen.getByRole('radio', { name: 'No' }))

    expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'doc-1')
    await userEvent.click(screen.getByRole('radio', { name: 'Sí' }))
    expect(screen.queryByText('respaldo.pdf')).not.toBeInTheDocument()
  })

  it('si el servidor aún usa el PDF (409) el cambio a No no muestra error', async () => {
    mockEliminar.mockRejectedValue(new ApiError(409, 'CONFLICT', 'En uso'))
    montar({
      valoresIniciales: {
        cargo: 'afiliados',
        respaldo_organizacion: { nombre: 'Federación', documento_id: 'doc-7' },
      },
    })

    await userEvent.click(screen.getByRole('radio', { name: 'No' }))

    await waitFor(() => expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'doc-7'))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('al guardar el No borra el PDF que tenía guardada la postulación', async () => {
    const { onGuardar } = montar({
      valoresIniciales: {
        cargo: 'afiliados',
        respaldo_organizacion: { nombre: 'Federación', documento_id: 'doc-7' },
      },
    })
    await userEvent.click(screen.getByRole('radio', { name: 'No' }))
    mockEliminar.mockClear()

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    await waitFor(() => expect(onGuardar).toHaveBeenCalled())
    await waitFor(() => expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'doc-7'))
  })

  it('muestra un mensaje genérico si el guardado falla sin respuesta del servidor', async () => {
    montar({ onGuardar: vi.fn().mockRejectedValue(new Error('boom')) })
    await userEvent.click(opcion('afiliados'))
    await screen.findByText('Postulante afiliado al IESS')
    await userEvent.click(screen.getByRole('radio', { name: 'No' }))

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(await screen.findByText('No se pudo guardar la información.')).toBeInTheDocument()
  })
})

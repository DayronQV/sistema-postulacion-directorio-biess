import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ApiError } from '@/services/errores'
import {
  guardarPostulacion,
  obtenerPostulacion,
  validarCargo,
  type Postulacion,
} from '@/services/postulacion'
import { ETIQUETA_CARGO } from '../reglasCargo'
import CargoSeccion from './CargoSeccion'
import ContenedorToasts from '@/components/ui/ContenedorToasts'
import { reiniciarToasts } from '@/store/toasts'

vi.mock('@/services/postulacion')

// Mensajes literales del requerimiento (1.4.1.2, 1.4.2.2 y 1.4.3.2).
const NO_AFILIADO =
  'Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS, NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS'
const NO_JUBILADO =
  'Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación'
const NO_EMPLEADOR =
  'Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador'

function postulacion(sobre: Partial<Postulacion> = {}): Postulacion {
  return {
    id: 'p-1',
    codigo: 'BIESS-2026-000001',
    estado: 'borrador',
    informacion_personal: {
      cedula: '1709876542',
      email: 'ana@correo.com',
      nombres: 'Ana',
      apellidos: 'Mora',
      genero: 'Femenino',
      fecha_nacimiento: '1980-01-01',
      nacionalidad: 'Ecuatoriana',
      estado_civil: 'Soltero/a',
    },
    domicilio: {
      pais_id: 1,
      provincia_id: 17,
      canton_id: 170,
      provincia_texto: '',
      ciudad_texto: '',
      telefono_celular: '0999999999',
      email_secundario: '',
      calle_principal: 'Av. Amazonas',
      numero: '',
      calle_secundaria: '',
      sector: '',
      referencia: '',
      telefono_domicilio: '',
    },
    tiene_familiares: true,
    familiares: [
      {
        nombres_apellidos: 'Luis',
        parentesco: 'hijo/a',
        institucion: 'IESS',
        area: 'F',
        cargo: 'A',
      },
    ],
    referencias: [
      { nombres_apellidos: 'Rosa', relacion: 'Colega', telefono: '0988888888', correo: 'r@c.com' },
    ],
    formacion: [],
    experiencia: [],
    duracion_experiencia_total: { anios: 0, meses: 0, dias: 0 },
    publicaciones: [],
    capacitaciones: [],
    cargo: '',
    respaldo_organizacion: { nombre: '', documento_id: null },
    tiene_publicaciones: false,
    consentimiento: true,
    updated_at: '2026-01-01T00:00:00Z',
    ...sobre,
  }
}

const opcion = (cargo: keyof typeof ETIQUETA_CARGO) =>
  screen.getByRole('radio', { name: ETIQUETA_CARGO[cargo] })

afterEach(() => {
  vi.resetAllMocks()
})

beforeEach(() => {
  // El aviso de guardado es un toast: se monta su contenedor y se limpia entre pruebas.
  reiniciarToasts()
  render(<ContenedorToasts />)
})

describe('<CargoSeccion />', () => {
  it('muestra un cargador mientras lee la postulación', () => {
    vi.mocked(obtenerPostulacion).mockReturnValue(new Promise(() => {}))
    render(<CargoSeccion />)

    expect(screen.getByRole('status')).toHaveTextContent('Cargando su postulación')
  })

  it('muestra el error de carga y permite reintentar', async () => {
    vi.mocked(obtenerPostulacion)
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL_ERROR', 'Error interno del servidor'))
      .mockResolvedValueOnce(postulacion())
    render(<CargoSeccion />)

    expect(await screen.findByText('No se pudo cargar la postulación')).toBeInTheDocument()
    expect(screen.getByText('Error interno del servidor')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByText('Selección del cargo')).toBeInTheDocument()
  })

  it('muestra los tres cargos cuando la postulación todavía no tiene uno', async () => {
    vi.mocked(obtenerPostulacion).mockResolvedValue(postulacion())
    render(<CargoSeccion />)

    expect(await screen.findByText('Selección del cargo')).toBeInTheDocument()
    expect(screen.getAllByRole('radio', { name: /Miembro Principal y Alterno/ })).toHaveLength(3)
    expect(validarCargo).not.toHaveBeenCalled()
  })

  it('valida el cargo elegido contra el servidor y muestra el resultado', async () => {
    vi.mocked(obtenerPostulacion).mockResolvedValue(postulacion())
    vi.mocked(validarCargo).mockResolvedValue({
      cargo: 'jubilados',
      elegible: true,
      mensaje: 'Postulante acredita jubilación',
    })
    render(<CargoSeccion />)
    await screen.findByText('Selección del cargo')

    await userEvent.click(opcion('jubilados'))

    expect(await screen.findByText('Postulante acredita jubilación')).toBeInTheDocument()
    expect(validarCargo).toHaveBeenCalledWith('jubilados')
    expect(screen.getByText(/organización de jubilados/)).toBeInTheDocument()
  })

  it.each([
    ['afiliados', NO_AFILIADO],
    ['jubilados', NO_JUBILADO],
    ['empleador', NO_EMPLEADOR],
  ] as const)(
    'si no cumple como %s muestra el mensaje del requerimiento y bloquea el formulario',
    async (cargo, mensaje) => {
      vi.mocked(obtenerPostulacion).mockResolvedValue(postulacion())
      vi.mocked(validarCargo).mockRejectedValue(new ApiError(422, 'VALIDATION_ERROR', mensaje))
      render(<CargoSeccion />)
      await screen.findByText('Selección del cargo')

      await userEvent.click(opcion(cargo))

      expect(
        await screen.findByRole('heading', { name: 'Postulación cerrada' }),
      ).toBeInTheDocument()
      expect(screen.getByRole('alert')).toHaveTextContent(mensaje)
      // Sin acceso al resto del formulario: ni cargos ni botón de guardar.
      expect(screen.queryByRole('radio')).not.toBeInTheDocument()
      expect(screen.queryByRole('button')).not.toBeInTheDocument()
      expect(screen.queryByText('Selección del cargo')).not.toBeInTheDocument()
    },
  )

  it('bloquea de entrada a una persona no ecuatoriana', async () => {
    const base = postulacion()
    vi.mocked(obtenerPostulacion).mockResolvedValue({
      ...base,
      informacion_personal: { ...base.informacion_personal, nacionalidad: 'Colombiana' },
    })
    render(<CargoSeccion />)

    expect(await screen.findByRole('heading', { name: 'Postulación cerrada' })).toBeInTheDocument()
    expect(screen.queryByRole('radio')).not.toBeInTheDocument()
  })

  it('bloquea una postulación ya cerrada', async () => {
    vi.mocked(obtenerPostulacion).mockResolvedValue(postulacion({ estado: 'cerrada' }))
    render(<CargoSeccion />)

    expect(await screen.findByRole('heading', { name: 'Postulación cerrada' })).toBeInTheDocument()
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('guarda la postulación entera cambiando solo el cargo y el respaldo', async () => {
    const actual = postulacion()
    vi.mocked(obtenerPostulacion).mockResolvedValue(actual)
    vi.mocked(guardarPostulacion).mockResolvedValue(actual)
    vi.mocked(validarCargo).mockResolvedValue({
      cargo: 'afiliados',
      elegible: true,
      mensaje: 'Postulante afiliado al IESS',
    })
    render(<CargoSeccion />)
    await screen.findByText('Selección del cargo')
    await userEvent.click(opcion('afiliados'))
    await screen.findByText('Postulante afiliado al IESS')
    await userEvent.click(screen.getByRole('radio', { name: 'No' }))

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(await screen.findByText('La postulación se guardó correctamente.')).toBeInTheDocument()
    expect(guardarPostulacion).toHaveBeenCalledTimes(1)
    const enviado = vi.mocked(guardarPostulacion).mock.calls[0][0]
    expect(enviado).toEqual({
      ...actual,
      cargo: 'afiliados',
      respaldo_organizacion: { nombre: '', documento_id: null },
    })
    expect(enviado.domicilio.calle_principal).toBe('Av. Amazonas')
    expect(enviado.familiares).toHaveLength(1)
    expect(enviado.referencias).toHaveLength(1)
  })

  it('si el servidor cierra la postulación al guardar (422), pasa a la pantalla de bloqueo', async () => {
    vi.mocked(obtenerPostulacion).mockResolvedValue(postulacion({ cargo: 'jubilados' }))
    vi.mocked(guardarPostulacion).mockRejectedValue(
      new ApiError(422, 'VALIDATION_ERROR', NO_JUBILADO),
    )
    render(<CargoSeccion />)
    await screen.findByText('Selección del cargo')

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Postulación cerrada' })).toBeInTheDocument()
    })
    expect(screen.getByRole('alert')).toHaveTextContent(NO_JUBILADO)
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  it('si la postulación ya no es un borrador (409) la vuelve a leer y muestra su estado', async () => {
    vi.mocked(obtenerPostulacion)
      .mockResolvedValueOnce(postulacion({ cargo: 'afiliados' })) // carga inicial
      .mockResolvedValueOnce(postulacion({ cargo: 'afiliados' })) // lectura previa al guardado
      .mockResolvedValue(postulacion({ cargo: 'afiliados', estado: 'cerrada' })) // relectura
    vi.mocked(guardarPostulacion).mockRejectedValue(
      new ApiError(409, 'CONFLICT', 'La postulación ya no se puede modificar'),
    )
    render(<CargoSeccion />)
    await screen.findByText('Selección del cargo')

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(await screen.findByRole('heading', { name: 'Postulación cerrada' })).toBeInTheDocument()
  })

  it('muestra el error del guardado y deja volver a intentarlo', async () => {
    vi.mocked(obtenerPostulacion).mockResolvedValue(postulacion({ cargo: 'afiliados' }))
    vi.mocked(guardarPostulacion).mockRejectedValue(
      new ApiError(500, 'INTERNAL_ERROR', 'Error interno del servidor'),
    )
    render(<CargoSeccion />)
    await screen.findByText('Selección del cargo')

    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(await screen.findByText('Error interno del servidor')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Guardar Postulación' })).toBeEnabled()
    expect(screen.queryByRole('heading', { name: 'Postulación cerrada' })).not.toBeInTheDocument()
  })
})

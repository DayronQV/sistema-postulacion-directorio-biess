import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { listarCantones, listarPaises, listarProvincias } from '@/services/catalogos'
import { ApiError } from '@/services/errores'
import {
  guardarPostulacion,
  obtenerPostulacion,
  type Domicilio as DomicilioApi,
  type Familiar,
  type Postulacion,
} from '@/services/postulacion'
import DomicilioSeccion from './DomicilioSeccion'

vi.mock('@/services/catalogos')
vi.mock('@/services/postulacion')

const SIN_DOMICILIO: DomicilioApi = {
  pais_id: null,
  provincia_id: null,
  canton_id: null,
  provincia_texto: '',
  ciudad_texto: '',
  telefono_celular: '',
  email_secundario: '',
  calle_principal: '',
  numero: '',
  calle_secundaria: '',
  sector: '',
  referencia: '',
  telefono_domicilio: '',
}

function postulacion(sobre: Partial<Postulacion> = {}): Postulacion {
  return {
    id: 'p-1',
    codigo: 'BIESS-2026-000001',
    estado: 'borrador',
    informacion_personal: {
      cedula: '1712345675',
      email: 'ana@correo.com',
      nombres: 'Ana',
      apellidos: 'Pérez',
      genero: 'Mujer',
      fecha_nacimiento: '1990-03-04',
      nacionalidad: 'Ecuatoriana',
      estado_civil: 'soltero',
    },
    domicilio: SIN_DOMICILIO,
    tiene_familiares: false,
    familiares: [],
    referencias: [],
    formacion: [],
    experiencia: [],
    duracion_experiencia_total: { anios: 0, meses: 0, dias: 0 },
    cargo: '',
    respaldo_organizacion: { nombre: '', documento_id: null },
    tiene_publicaciones: false,
    consentimiento: false,
    updated_at: '2026-01-01T00:00:00Z',
    ...sobre,
  }
}

beforeEach(() => {
  vi.mocked(listarPaises).mockResolvedValue([
    { id: 52, nombre: 'Ecuador' },
    { id: 40, nombre: 'Colombia' },
  ])
  vi.mocked(listarProvincias).mockResolvedValue([{ id: 17, nombre: 'Pichincha' }])
  vi.mocked(listarCantones).mockResolvedValue([{ id: 170, nombre: 'Quito' }])
})

afterEach(() => {
  vi.resetAllMocks()
})

describe('DomicilioSeccion (conectado a la API)', () => {
  it('muestra un loader mientras carga la postulación y luego el correo del registro', async () => {
    vi.mocked(obtenerPostulacion).mockResolvedValue(postulacion())
    render(<DomicilioSeccion />)

    expect(screen.getByText('Cargando su postulación...')).toBeInTheDocument()
    expect(await screen.findByLabelText(/^Correo electrónico 1/)).toHaveValue('ana@correo.com')
  })

  it('muestra el error de carga y permite reintentar', async () => {
    vi.mocked(obtenerPostulacion)
      .mockRejectedValueOnce(
        new ApiError(0, 'NETWORK_ERROR', 'No se pudo conectar con el servidor.'),
      )
      .mockResolvedValue(postulacion())
    const user = userEvent.setup()
    render(<DomicilioSeccion />)

    expect(await screen.findByText('No se pudo conectar con el servidor.')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByLabelText(/^Correo electrónico 1/)).toBeInTheDocument()
  })

  it('guarda el formulario entero: conserva las otras secciones, aunque hayan cambiado', async () => {
    const familiar: Familiar = {
      nombres_apellidos: 'Beto',
      parentesco: 'hermano/a',
      institucion: 'IESS',
      area: 'Finanzas',
      cargo: 'Analista',
    }
    // Al abrir la sección no había familiares; mientras el usuario escribía, otra sección
    // guardó uno. El guardado debe partir de lo que hay en el servidor en ese momento.
    vi.mocked(obtenerPostulacion)
      .mockResolvedValueOnce(postulacion())
      .mockResolvedValueOnce(
        postulacion({ tiene_familiares: true, familiares: [familiar], cargo: 'afiliados' }),
      )
    vi.mocked(guardarPostulacion).mockImplementation(async (f) => f)
    const user = userEvent.setup()
    render(<DomicilioSeccion />)

    await user.selectOptions(await screen.findByLabelText(/^País/), 'Colombia')
    await user.type(screen.getByLabelText(/^Provincia \/ Estado/), 'Antioquia')
    await user.type(screen.getByLabelText(/^Ciudad\*$/), 'Medellín')
    await user.type(screen.getByLabelText(/^Teléfono celular/), '0999999999')
    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('Domicilio guardado.')).toBeInTheDocument()
    expect(guardarPostulacion).toHaveBeenCalledTimes(1)
    expect(guardarPostulacion).toHaveBeenCalledWith(
      postulacion({
        tiene_familiares: true,
        familiares: [familiar],
        cargo: 'afiliados',
        domicilio: {
          ...SIN_DOMICILIO,
          pais_id: 40,
          provincia_texto: 'Antioquia',
          ciudad_texto: 'Medellín',
          telefono_celular: '0999999999',
        },
      }),
    )
  })

  it('muestra el rechazo del servidor sin perder lo escrito', async () => {
    vi.mocked(obtenerPostulacion).mockResolvedValue(postulacion())
    vi.mocked(guardarPostulacion).mockRejectedValue(
      new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos', {
        'domicilio.telefono_celular': 'El celular debe tener 10 dígitos y empezar con 09',
      }),
    )
    const user = userEvent.setup()
    render(<DomicilioSeccion />)

    await user.selectOptions(await screen.findByLabelText(/^País/), 'Colombia')
    await user.type(screen.getByLabelText(/^Provincia \/ Estado/), 'Antioquia')
    await user.type(screen.getByLabelText(/^Ciudad\*$/), 'Medellín')
    await user.type(screen.getByLabelText(/^Teléfono celular/), '0999999999')
    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('Datos inválidos')).toBeInTheDocument()
    expect(
      await screen.findByText('El celular debe tener 10 dígitos y empezar con 09'),
    ).toBeInTheDocument()
    expect(screen.getByLabelText(/^Provincia \/ Estado/)).toHaveValue('Antioquia')
    expect(screen.queryByText('Domicilio guardado.')).not.toBeInTheDocument()
  })

  it('abre con el domicilio ya guardado', async () => {
    vi.mocked(obtenerPostulacion).mockResolvedValue(
      postulacion({
        domicilio: {
          ...SIN_DOMICILIO,
          pais_id: 52,
          provincia_id: 17,
          canton_id: 170,
          telefono_celular: '0988888888',
        },
      }),
    )
    render(<DomicilioSeccion />)

    await waitFor(() => expect(screen.getByLabelText(/^Ciudad \/ Cantón/)).toHaveValue('170'))
    expect(screen.getByLabelText(/^Provincia/)).toHaveValue('17')
    expect(screen.getByLabelText(/^Teléfono celular/)).toHaveValue('0988888888')
  })
})

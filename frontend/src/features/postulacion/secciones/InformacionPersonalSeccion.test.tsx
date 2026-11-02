import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ApiError } from '@/services/errores'
import { guardarPostulacion, obtenerPostulacion, type Postulacion } from '@/services/postulacion'
import InformacionPersonalSeccion from './InformacionPersonalSeccion'

vi.mock('@/services/postulacion')

function postulacion(sobre: Partial<Postulacion> = {}): Postulacion {
  return {
    id: 'p-1',
    codigo: 'BIESS-2026-000001',
    estado: 'borrador',
    informacion_personal: {
      cedula: '0912345675',
      email: 'maria@correo.com',
      nombres: 'María Fernanda',
      apellidos: 'Zambrano Cedeño',
      genero: 'Femenino',
      fecha_nacimiento: '1978-11-02',
      nacionalidad: 'Ecuatoriana',
      estado_civil: 'Casado/a',
      pareja: { nombres_apellidos: 'Luis Alberto Párraga Mera', cedula: '0913579249' },
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
        nombres_apellidos: 'Ana Pérez',
        parentesco: 'hermano/a',
        institucion: 'IESS',
        area: 'F',
        cargo: 'A',
      },
    ],
    referencias: [
      {
        nombres_apellidos: 'Rosa Vera',
        relacion: 'Colega',
        telefono: '0988888888',
        correo: 'rosa@correo.com',
      },
    ],
    formacion: [],
    experiencia: [],
    duracion_experiencia_total: { anios: 0, meses: 0, dias: 0 },
    cargo: 'jubilados',
    respaldo_organizacion: { nombre: 'Asociación', documento_id: null },
    tiene_publicaciones: false,
    consentimiento: true,
    updated_at: '2026-01-01T00:00:00Z',
    ...sobre,
  }
}

const extranjera = (estado: Postulacion['estado']) => {
  const base = postulacion({ estado })
  return {
    ...base,
    informacion_personal: {
      ...base.informacion_personal,
      nacionalidad: 'Colombiana',
      estado_civil: 'Soltero/a',
      pareja: undefined,
    },
  }
}

afterEach(() => {
  vi.resetAllMocks()
})

describe('<InformacionPersonalSeccion />', () => {
  it('muestra un cargador mientras lee la postulación', () => {
    vi.mocked(obtenerPostulacion).mockReturnValue(new Promise(() => {}))
    render(<InformacionPersonalSeccion />)

    expect(screen.getByRole('status')).toHaveTextContent('Cargando su postulación')
  })

  it('muestra la información personal y la pareja de una persona casada', async () => {
    vi.mocked(obtenerPostulacion).mockResolvedValue(postulacion())
    render(<InformacionPersonalSeccion />)

    expect(await screen.findByText('María Fernanda')).toBeInTheDocument()
    expect(screen.getByText('02/11/1978')).toBeInTheDocument()
    expect(screen.getByText('Luis Alberto Párraga Mera')).toBeInTheDocument()
    expect(screen.getByText('0913579249')).toBeInTheDocument()
  })

  it('muestra el error de carga y permite reintentar', async () => {
    vi.mocked(obtenerPostulacion)
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL_ERROR', 'Error interno del servidor'))
      .mockResolvedValueOnce(postulacion())
    render(<InformacionPersonalSeccion />)

    expect(await screen.findByText('No se pudo cargar la postulación')).toBeInTheDocument()
    expect(screen.getByText('Error interno del servidor')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByText('María Fernanda')).toBeInTheDocument()
  })

  it.each(['borrador', 'cerrada'] as const)(
    'bloquea a una persona no ecuatoriana con la postulación en %s',
    async (estado) => {
      vi.mocked(obtenerPostulacion).mockResolvedValue(extranjera(estado))
      render(<InformacionPersonalSeccion />)

      expect(
        await screen.findByRole('heading', { name: 'Postulación cerrada' }),
      ).toBeInTheDocument()
      expect(screen.getByRole('alert')).toHaveTextContent(/nacionalidad ecuatoriana/)
      expect(screen.getByRole('alert')).toHaveTextContent('Colombiana')
      // Sin acceso al resto del formulario: ni datos ni botón de guardar.
      expect(screen.queryByRole('button')).not.toBeInTheDocument()
      expect(screen.queryByText('Información personal')).not.toBeInTheDocument()
    },
  )

  it('guarda la postulación entera, sin perder las demás secciones', async () => {
    const actual = postulacion()
    vi.mocked(obtenerPostulacion).mockResolvedValue(actual)
    vi.mocked(guardarPostulacion).mockResolvedValue(actual)
    render(<InformacionPersonalSeccion />)
    await screen.findByText('María Fernanda')

    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('La información personal se guardó.')).toBeInTheDocument()
    expect(guardarPostulacion).toHaveBeenCalledTimes(1)
    const enviado = vi.mocked(guardarPostulacion).mock.calls[0][0]
    expect(enviado).toEqual(actual)
    expect(enviado.domicilio.calle_principal).toBe('Av. Amazonas')
    expect(enviado.familiares).toHaveLength(1)
    expect(enviado.referencias).toHaveLength(1)
    expect(enviado.cargo).toBe('jubilados')
  })

  it('muestra el error del guardado', async () => {
    vi.mocked(obtenerPostulacion).mockResolvedValue(postulacion())
    vi.mocked(guardarPostulacion).mockRejectedValue(
      new ApiError(500, 'INTERNAL_ERROR', 'Error interno del servidor'),
    )
    render(<InformacionPersonalSeccion />)
    await screen.findByText('María Fernanda')

    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('Error interno del servidor')
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeEnabled()
  })

  it('si el servidor cierra la postulación al guardar (422), pasa a la pantalla de bloqueo', async () => {
    vi.mocked(obtenerPostulacion)
      .mockResolvedValueOnce(postulacion()) // carga inicial
      .mockResolvedValueOnce(postulacion()) // lectura previa al guardado
      .mockResolvedValue(postulacion({ estado: 'cerrada' })) // relectura tras el 422
    vi.mocked(guardarPostulacion).mockRejectedValue(
      new ApiError(
        422,
        'VALIDATION_ERROR',
        'Solo pueden postularse personas de nacionalidad ecuatoriana',
      ),
    )
    render(<InformacionPersonalSeccion />)
    await screen.findByText('María Fernanda')

    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Postulación cerrada' })).toBeInTheDocument()
    })
    expect(screen.queryByRole('button', { name: 'Guardar' })).not.toBeInTheDocument()
  })
})

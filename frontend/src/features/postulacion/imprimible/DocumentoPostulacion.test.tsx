import { render, screen, within } from '@testing-library/react'
import type { Postulacion } from '@/services/postulacion'
import DocumentoPostulacion from './DocumentoPostulacion'
import { formatearDuracion, formatearFecha } from './formato'
import { CONSENTIMIENTO } from './textos'

const postulacion = {
  codigo: 'POST-2026-0001',
  estado: 'borrador',
  informacion_personal: {
    nombres: 'Ana María',
    apellidos: 'Pérez Loor',
    cedula: '1712345675',
    genero: 'Femenino',
    fecha_nacimiento: '1990-03-05',
    nacionalidad: 'Ecuatoriana',
    estado_civil: 'Soltero/a',
    email: 'ana@correo.com',
  },
  cargo: 'afiliados',
  experiencia: [
    {
      tipo: 'general',
      cargo: 'Analista',
      institucion: 'Banco Central',
      fecha_inicio: '2018-01-01',
      fecha_fin: '2020-03-15',
      actividades: 'finanzas',
      descripcion: 'Análisis de riesgos',
      documento_id: 'd1',
      duracion: { anios: 2, meses: 2, dias: 14 },
    },
  ],
  inhabilidades: {
    '1': true,
    '7': true,
    removido_detalle: {
      institucion: 'SB',
      cargo: 'Jefe',
      fecha_fin: '2022-06-30',
      motivo: 'Fin de contrato',
    },
  },
} as unknown as Postulacion

describe('formato', () => {
  it('convierte la fecha sin desfase de zona horaria', () => {
    expect(formatearFecha('2020-01-31')).toBe('31/01/2020')
    expect(formatearFecha('')).toBe('—')
  })

  it('expresa la duración en años, meses y días con su plural', () => {
    expect(formatearDuracion({ anios: 2, meses: 1, dias: 14 })).toBe('2 años, 1 mes, 14 días')
    expect(formatearDuracion({ anios: 1, meses: 0, dias: 1 })).toBe('1 año, 0 meses, 1 día')
  })
})

describe('DocumentoPostulacion', () => {
  it('muestra el encabezado con el código y la fecha', () => {
    render(<DocumentoPostulacion postulacion={postulacion} fecha={new Date(2026, 10, 1)} />)
    expect(screen.getByText('POST-2026-0001')).toBeInTheDocument()
    expect(screen.getByText(/01\/11\/2026/)).toBeInTheDocument()
  })

  it('incluye los datos personales y el tiempo de experiencia calculado', () => {
    render(<DocumentoPostulacion postulacion={postulacion} />)
    const personal = screen
      .getByRole('heading', { name: '1. Información personal' })
      .closest('section')!
    expect(within(personal).getByText('Ana María Pérez Loor')).toBeInTheDocument()
    expect(screen.getByText('05/03/1990')).toBeInTheDocument()
    expect(screen.getByText(/Tiempo: 2 años, 2 meses, 14 días/)).toBeInTheDocument()
  })

  it('lista las 13 preguntas de inhabilidades con el detalle de la 7', () => {
    render(<DocumentoPostulacion postulacion={postulacion} />)
    const bloque = screen.getByRole('heading', { name: '9. Inhabilidades' }).closest('section')!
    expect(within(bloque).getAllByRole('listitem')).toHaveLength(13)
    expect(within(bloque).getByText(/Fin de contrato/)).toBeInTheDocument()
  })

  it('incluye los 5 numerales del consentimiento', () => {
    render(<DocumentoPostulacion postulacion={postulacion} />)
    const bloque = screen
      .getByRole('heading', { name: '10. Consentimiento y declaración' })
      .closest('section')!
    expect(within(bloque).getAllByRole('listitem')).toHaveLength(CONSENTIMIENTO.length)
    expect(within(bloque).getByText(/firma manuscrita/)).toBeInTheDocument()
  })

  it('termina con el apartado de firma, nombres y cédula', () => {
    render(<DocumentoPostulacion postulacion={postulacion} />)
    const firma = screen.getByRole('region', { name: 'Firma del postulante' })
    expect(within(firma).getByText('Ana María Pérez Loor')).toBeInTheDocument()
    expect(within(firma).getByText('C.I. 1712345675')).toBeInTheDocument()
    expect(screen.getByRole('article').lastElementChild).toBe(firma)
  })

  it('muestra los datos de la pareja cuando el estado civil los exige', () => {
    const casada = {
      ...postulacion,
      informacion_personal: {
        ...postulacion.informacion_personal,
        estado_civil: 'Casado/a',
        pareja: { nombres_apellidos: 'Luis Párraga Mera', cedula: '0913579249' },
      },
    } as Postulacion
    render(<DocumentoPostulacion postulacion={casada} />)
    expect(screen.getByText('Luis Párraga Mera')).toBeInTheDocument()
    expect(screen.getByText('0913579249')).toBeInTheDocument()
  })

  it('muestra los nombres de país, provincia y ciudad, no los ids del catálogo', () => {
    render(
      <DocumentoPostulacion
        postulacion={postulacion}
        ubicacion={{ pais: 'Ecuador', provincia: 'Pichincha', ciudad: 'Quito' }}
      />,
    )
    expect(screen.getByText('Ecuador')).toBeInTheDocument()
    expect(screen.getByText('Pichincha')).toBeInTheDocument()
    expect(screen.getByText('Quito')).toBeInTheDocument()
  })

  it('usa las etiquetas del formulario y no los valores internos de la API', () => {
    render(<DocumentoPostulacion postulacion={postulacion} />)
    const experiencia = screen.getByRole('heading', { name: '6. Experiencia profesional' })
      .parentElement as HTMLElement
    expect(within(experiencia).getByText(/\(General\)/)).toBeInTheDocument()
    expect(within(experiencia).getByText(/relacionadas a Finanzas/)).toBeInTheDocument()
    expect(within(experiencia).queryByText(/\(general\)/)).toBeNull()
  })

  it('incluye el tiempo total de experiencia', () => {
    render(
      <DocumentoPostulacion
        postulacion={{
          ...postulacion,
          duracion_experiencia_total: { anios: 2, meses: 2, dias: 14 },
        }}
      />,
    )
    expect(
      screen.getByText('Tiempo total de experiencia: 2 años, 2 meses, 14 días.'),
    ).toBeInTheDocument()
  })
})

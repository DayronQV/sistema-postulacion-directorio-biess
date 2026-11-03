import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi, type Mock } from 'vitest'
import { obtenerPostulacionAdmin } from '@/services/adminPostulaciones'
import PostulacionDetalle from './PostulacionDetalle'

vi.mock('@/services/adminPostulaciones', () => ({
  obtenerPostulacionAdmin: vi.fn(),
  descargarDocumentoAdmin: vi.fn(),
}))
vi.mock('@/features/postulacion/imprimible/useUbicacion', () => ({
  useUbicacion: () => ({ pais: 'Ecuador', provincia: 'Guayas', ciudad: 'Guayaquil' }),
}))
const mockObtener = obtenerPostulacionAdmin as Mock

beforeEach(() => {
  mockObtener.mockReset()
})

const postulacion = {
  id: 'p1',
  codigo: 'BIESS-2026-000001',
  estado: 'enviada',
  informacion_personal: {
    cedula: '0912345678',
    email: 'ana@example.com',
    nombres: 'Ana',
    apellidos: 'Pérez',
    genero: 'F',
    fecha_nacimiento: '1990-01-02',
    nacionalidad: 'Ecuatoriana',
    estado_civil: 'Soltera',
  },
  domicilio: {
    pais_id: 1,
    provincia_id: 1,
    canton_id: 1,
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
  },
  tiene_familiares: false,
  familiares: [],
  referencias: [],
  formacion: [
    {
      nivel: 'tercer_nivel',
      institucion: 'UPSE',
      tipo: 'nacional',
      titulo: 'Economista',
      area: 'economia',
      senescyt_numero: '1',
      senescyt_fecha: '2020-01-01',
      documento_id: 'doc-1',
    },
  ],
  experiencia: [
    {
      tipo: 'general',
      cargo: 'Analista',
      institucion: 'BIESS',
      fecha_inicio: '2020-01-01',
      fecha_fin: '2025-03-02',
      actividades: 'economia',
      descripcion: 'Análisis',
      documento_id: '',
      duracion: { anios: 5, meses: 2, dias: 1 },
    },
  ],
  duracion_experiencia_total: { anios: 5, meses: 2, dias: 1 },
  publicaciones: [],
  capacitaciones: [],
  cargo: 'afiliados',
  respaldo_organizacion: { nombre: '', documento_id: null },
  tiene_publicaciones: false,
  consentimiento: true,
  updated_at: '2026-11-02T00:00:00Z',
  inhabilidades: {},
  elegible: true,
}

it('muestra el formulario íntegro y los certificados', async () => {
  mockObtener.mockResolvedValue(postulacion)
  render(
    <MemoryRouter initialEntries={['/admin/postulaciones/p1']}>
      <Routes>
        <Route path="/admin/postulaciones/:id" element={<PostulacionDetalle />} />
      </Routes>
    </MemoryRouter>,
  )

  expect(await screen.findByText('BIESS-2026-000001')).toBeInTheDocument()
  expect(screen.getByText('Formación 1: Economista')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Descargar PDF' })).toBeInTheDocument()
  expect(
    screen.getByText(/Tiempo total de experiencia: 5 años, 2 meses, 1 día/),
  ).toBeInTheDocument()
  expect(mockObtener).toHaveBeenCalledWith(expect.any(AbortSignal), 'p1')
})

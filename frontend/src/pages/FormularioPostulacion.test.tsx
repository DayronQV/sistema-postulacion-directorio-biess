import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { vi } from 'vitest'
import { reiniciarFormulario } from '@/features/postulacion/store'
import { ApiError } from '@/services/errores'
import * as servicio from '@/services/postulacion'
import FormularioPostulacion from './FormularioPostulacion'

vi.mock('@/services/postulacion')
const obtener = vi.mocked(servicio.obtenerPostulacion)
const guardar = vi.mocked(servicio.guardarPostulacion)

const base = {
  id: 'p1',
  codigo: 'POST-2026-0001',
  estado: 'borrador',
  tiene_familiares: false,
  informacion_personal: {
    nombres: 'Ana',
    apellidos: 'Pérez',
    cedula: '1712345675',
    genero: 'F',
    fecha_nacimiento: '1990-01-01',
    nacionalidad: 'Ecuatoriana',
    estado_civil: 'Soltero/a',
    email: 'ana@correo.com',
  },
} as servicio.Postulacion

beforeEach(() => {
  reiniciarFormulario()
  obtener.mockReset()
  guardar.mockReset()
})

describe('FormularioPostulacion', () => {
  it('muestra el cargador y luego el código, el estado y el avance', async () => {
    obtener.mockResolvedValue(base)
    render(<FormularioPostulacion />)
    expect(screen.getByText('Cargando su postulación...')).toBeInTheDocument()

    expect(await screen.findByText('POST-2026-0001')).toBeInTheDocument()
    expect(screen.getByText('Borrador')).toBeInTheDocument()
    expect(screen.getByText(/Secciones completadas:/)).toHaveTextContent('2 de 11')
    expect(screen.getByRole('progressbar', { name: 'Avance del formulario' })).toHaveAttribute(
      'aria-valuenow',
      '18',
    )
  })

  it('muestra el error de carga y permite reintentar', async () => {
    obtener.mockRejectedValueOnce(new ApiError(500, 'INTERNAL_ERROR', 'Error del servidor'))
    obtener.mockResolvedValueOnce(base)
    render(<FormularioPostulacion />)

    expect(await screen.findByText('Error del servidor')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Reintentar' }))
    expect(await screen.findByText('POST-2026-0001')).toBeInTheDocument()
  })

  it('muestra la pantalla de bloqueo si la nacionalidad no es ecuatoriana', async () => {
    obtener.mockResolvedValue({
      ...base,
      informacion_personal: { ...base.informacion_personal, nacionalidad: 'Colombiana' },
    })
    render(<FormularioPostulacion />)

    expect(await screen.findByText('No puede continuar con la postulación')).toBeInTheDocument()
    expect(screen.queryByRole('navigation', { name: 'Secciones del formulario' })).toBeNull()
  })

  it('marca los cambios sin guardar de una sección y avisa antes de salir', async () => {
    obtener.mockResolvedValue({ ...base, familiares: [] })
    const user = userEvent.setup()
    render(<FormularioPostulacion />)
    await screen.findByText('POST-2026-0001')

    await user.click(screen.getByRole('button', { name: /3\. Información adicional/ }))
    // La sección carga por su cuenta; al responder "Sí" queda con cambios sin guardar.
    await user.click(await screen.findByRole('radio', { name: /Sí/ }))
    expect(screen.getByLabelText('Cambios sin guardar')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /5\. Formación académica/ }))
    expect(screen.getByText(/Tiene cambios sin guardar en esta sección/)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Descartar y continuar' }))
    expect(screen.getByRole('button', { name: /5\. Formación académica/ })).toHaveAttribute(
      'aria-current',
      'step',
    )
    expect(screen.queryByLabelText('Cambios sin guardar')).toBeNull()
  })

  it('con la postulación enviada no muestra secciones editables, solo la confirmación', async () => {
    obtener.mockResolvedValue({ ...base, estado: 'enviada', consentimiento: true })
    render(
      <MemoryRouter>
        <FormularioPostulacion />
      </MemoryRouter>,
    )

    expect(await screen.findByText('Postulación enviada')).toBeInTheDocument()
    expect(screen.queryByRole('navigation', { name: 'Secciones del formulario' })).toBeNull()
    expect(screen.queryByRole('button', { name: 'Guardar Postulación' })).toBeNull()
  })

  it('si superó la Fase de Méritos muestra el estado Evaluación Técnica', async () => {
    obtener.mockResolvedValue({
      ...base,
      estado: 'enviada',
      consentimiento: true,
      supero_merito: true,
      estado_visible: 'Evaluación Técnica',
    })
    render(
      <MemoryRouter>
        <FormularioPostulacion />
      </MemoryRouter>,
    )

    expect(await screen.findByText('Evaluación Técnica')).toBeInTheDocument()
    expect(screen.getByText('Superó la Fase de Méritos')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Ir a la evaluación' })).toHaveAttribute(
      'href',
      '/evaluacion',
    )
  })

  it('avisa en el borrador cuando el periodo de postulación está cerrado', async () => {
    obtener.mockResolvedValue({ ...base, periodo_abierto: false })
    render(<FormularioPostulacion />)

    expect(
      await screen.findByText('El periodo de postulación no está habilitado'),
    ).toBeInTheDocument()
    // Las secciones siguen visibles para revisar la postulación.
    expect(screen.getByRole('navigation', { name: 'Secciones del formulario' })).toBeInTheDocument()
  })

  it('con el periodo abierto no muestra el aviso', async () => {
    obtener.mockResolvedValue({ ...base, periodo_abierto: true })
    render(<FormularioPostulacion />)
    await screen.findByText('POST-2026-0001')
    expect(screen.queryByText('El periodo de postulación no está habilitado')).toBeNull()
  })

  it('sin postulación y con el periodo cerrado avisa, sin botón de reintentar', async () => {
    obtener.mockRejectedValue(
      new ApiError(403, 'FORBIDDEN', 'El periodo de postulación no está habilitado'),
    )
    render(<FormularioPostulacion />)

    expect(
      await screen.findByText('El periodo de postulación no está habilitado'),
    ).toBeInTheDocument()
    expect(screen.getByText(/Todavía no puede iniciar su postulación/)).toBeInTheDocument()
    expect(screen.queryByText('No se pudo cargar la postulación')).toBeNull()
    expect(screen.queryByRole('button', { name: 'Reintentar' })).toBeNull()
  })

  it('con otro error de carga sigue ofreciendo reintentar', async () => {
    obtener.mockRejectedValue(new ApiError(500, 'INTERNAL_ERROR', 'Error interno'))
    render(<FormularioPostulacion />)

    expect(await screen.findByText('No se pudo cargar la postulación')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Reintentar' })).toBeInTheDocument()
  })

  it('navega entre secciones desde el menú', async () => {
    obtener.mockResolvedValue(base)
    render(<FormularioPostulacion />)
    await screen.findByText('POST-2026-0001')

    await userEvent.click(screen.getByRole('button', { name: /4\. Selección del cargo/ }))
    expect(screen.getByRole('heading', { level: 2, name: 'Selección del cargo' })).toBeVisible()
    expect(screen.getByRole('button', { name: /4\. Selección del cargo/ })).toHaveAttribute(
      'aria-current',
      'step',
    )
  })
})

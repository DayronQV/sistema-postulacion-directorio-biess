import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
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
    expect(screen.getByText('2 de 11 secciones completas')).toBeInTheDocument()
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

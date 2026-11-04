import { act, render, renderHook, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { vi } from 'vitest'
import * as servicio from '@/services/postulacion'
import * as sesion from '@/store/sesion'
import LayoutPostulante from './LayoutPostulante'
import { nombreCorto } from './informacionPersonal'
import { cargarPostulacion, reiniciarFormulario, useFormulario } from './store'

vi.mock('@/services/postulacion')
const obtener = vi.mocked(servicio.obtenerPostulacion)

function postulacion(): servicio.Postulacion {
  return {
    id: 'p1',
    codigo: 'BIESS-2026-000001',
    estado: 'borrador',
    informacion_personal: {
      cedula: '0942708264',
      nombres: 'Dayron Adrian',
      apellidos: 'Quiñonez Valencia',
    },
  } as servicio.Postulacion
}

function montar() {
  render(
    <MemoryRouter>
      <LayoutPostulante />
    </MemoryRouter>,
  )
  return userEvent.setup()
}

beforeEach(() => {
  // El marco consulta el ancho de la pantalla; jsdom no trae matchMedia.
  vi.stubGlobal('matchMedia', () => ({
    matches: true,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }))
  reiniciarFormulario()
  obtener.mockReset()
})

afterEach(() => {
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
})

describe('nombreCorto', () => {
  it('toma el primer nombre y el primer apellido', () => {
    expect(nombreCorto('Dayron Adrian', 'Quiñonez Valencia')).toBe('Dayron Quiñonez')
    expect(nombreCorto('  Julio   César ', ' Macías Pinargote ')).toBe('Julio Macías')
  })

  it('funciona con un solo nombre o apellido, o sin alguno', () => {
    expect(nombreCorto('Tommy', 'Loor')).toBe('Tommy Loor')
    expect(nombreCorto('Ana', '')).toBe('Ana')
    expect(nombreCorto('', '')).toBe('')
  })
})

describe('LayoutPostulante', () => {
  it('muestra el nombre y la cédula del postulante en lugar de "Postulante"', async () => {
    obtener.mockResolvedValue(postulacion())
    montar()

    expect(await screen.findByText('Dayron Quiñonez')).toBeInTheDocument()
    expect(screen.getByText('C.I. 0942708264')).toBeInTheDocument()
    expect(screen.queryByText('Postulante')).not.toBeInTheDocument()
  })

  it('mientras carga muestra el texto genérico', () => {
    obtener.mockReturnValue(new Promise(() => {}))
    montar()

    expect(screen.getByText('Postulante')).toBeInTheDocument()
  })

  it('mientras se vuelve a cargar no muestra el nombre anterior', async () => {
    obtener.mockResolvedValueOnce(postulacion())
    await cargarPostulacion()
    obtener.mockReturnValue(new Promise(() => {}))
    montar()
    expect(screen.getByText('Dayron Quiñonez')).toBeInTheDocument()

    // Otra carga en curso, por ejemplo al abrir el formulario con otra cuenta en la pestaña.
    act(() => {
      void cargarPostulacion()
    })

    expect(screen.queryByText('Dayron Quiñonez')).not.toBeInTheDocument()
    expect(screen.getByText('Postulante')).toBeInTheDocument()
  })

  it('no vuelve a pedir la postulación si ya está cargada', async () => {
    obtener.mockResolvedValue(postulacion())
    await cargarPostulacion()
    montar()

    expect(screen.getByText('Dayron Quiñonez')).toBeInTheDocument()
    expect(obtener).toHaveBeenCalledTimes(1)
  })

  it('al cerrar sesión limpia la postulación cargada', async () => {
    const cerrar = vi.spyOn(sesion, 'cerrarSesion').mockImplementation(() => {})
    obtener.mockResolvedValue(postulacion())
    const user = montar()
    await screen.findByText('Dayron Quiñonez')

    await user.click(screen.getByRole('button', { name: /Cerrar sesión/ }))

    expect(cerrar).toHaveBeenCalledTimes(1)
    expect(renderHook(() => useFormulario()).result.current.postulacion).toBeNull()
  })
})

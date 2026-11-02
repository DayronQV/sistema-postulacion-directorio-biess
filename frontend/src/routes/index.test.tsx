import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { vi } from 'vitest'
import AppRoutes from '@/routes'
import * as servicio from '@/services/postulacion'
import { cerrarSesion, expirarSesion, iniciarSesion } from '@/store/sesion'

const tituloFormulario = () =>
  screen.getByRole('heading', { level: 1, name: /^Formulario de postulación$/ })
const tituloLogin = () => screen.getByRole('heading', { level: 1, name: /^Iniciar sesión$/ })

// Token con el claim rol del backend: las rutas del postulante exigen ese rol.
const tokenPostulante = `x.${btoa(JSON.stringify({ rol: 'postulante' }))}.y`

function renderEn(ruta: string) {
  return render(
    <MemoryRouter initialEntries={[ruta]}>
      <AppRoutes />
    </MemoryRouter>,
  )
}

afterEach(() => cerrarSesion())

describe('AppRoutes', () => {
  it.each([
    ['/', /Concurso de Méritos y Oposición/],
    ['/registro', /^Registro$/],
    ['/login', /^Iniciar sesión$/],
    ['/recuperar', /^Recuperar contraseña$/],
    ['/restablecer?token=abc', /^Restablecer contraseña$/],
  ])('muestra la página pública de %s', (ruta, titulo) => {
    renderEn(ruta)
    expect(screen.getByRole('heading', { level: 1, name: titulo })).toBeInTheDocument()
  })

  it('redirige /postulacion al login si no hay sesión', () => {
    renderEn('/postulacion')
    expect(screen.getByRole('heading', { level: 1, name: /^Iniciar sesión$/ })).toBeInTheDocument()
  })

  it('muestra /postulacion con sesión abierta', async () => {
    vi.spyOn(servicio, 'obtenerPostulacion').mockResolvedValue({
      codigo: 'POST-1',
      estado: 'borrador',
      informacion_personal: {
        nombres: 'Ana',
        apellidos: 'Pérez',
        cedula: '1712345675',
        genero: 'Femenino',
        fecha_nacimiento: '1990-01-01',
        nacionalidad: 'Ecuatoriana',
        estado_civil: 'Soltero/a',
        email: 'ana@correo.com',
      },
    } as servicio.Postulacion)
    iniciarSesion(tokenPostulante, 600)
    renderEn('/postulacion')
    expect(
      await screen.findByRole('heading', { level: 1, name: /^Formulario de postulación$/ }),
    ).toBeInTheDocument()
  })

  it.each(['/login', '/registro'])('con sesión abierta, %s lleva al formulario', (ruta) => {
    iniciarSesion(tokenPostulante, 600)
    renderEn(ruta)
    expect(tituloFormulario()).toBeInTheDocument()
  })

  it('una URL desconocida lleva al inicio en lugar de una página en blanco', () => {
    renderEn('/no-existe')
    expect(
      screen.getByRole('heading', { level: 1, name: /Concurso de Méritos y Oposición/ }),
    ).toBeInTheDocument()
  })
})

describe('AppRoutes: fin de la sesión', () => {
  afterEach(() => vi.useRealTimers())

  it('al vencer el token avisa "Su sesión expiró" en el login', () => {
    vi.useFakeTimers()
    iniciarSesion(tokenPostulante, 60)
    renderEn('/postulacion')
    expect(tituloFormulario()).toBeInTheDocument()

    act(() => {
      vi.advanceTimersByTime(61_000)
    })
    expect(tituloLogin()).toBeInTheDocument()
    expect(screen.getByText('Su sesión expiró')).toBeInTheDocument()
  })

  it('un 401 del servidor (sesión expirada) también muestra el aviso', () => {
    iniciarSesion(tokenPostulante, 600)
    renderEn('/postulacion')
    act(() => expirarSesion())
    expect(tituloLogin()).toBeInTheDocument()
    expect(screen.getByText('Su sesión expiró')).toBeInTheDocument()
  })

  it('Cerrar sesión lleva al login sin el aviso de expiración', async () => {
    iniciarSesion(tokenPostulante, 600)
    renderEn('/postulacion')
    await userEvent.click(screen.getByRole('button', { name: /Cerrar sesión/ }))
    expect(tituloLogin()).toBeInTheDocument()
    expect(screen.queryByText('Su sesión expiró')).not.toBeInTheDocument()
  })
})

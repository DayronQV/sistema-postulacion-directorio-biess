import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AppRoutes from '@/routes'
import { cerrarSesion, iniciarSesion } from '@/store/sesion'

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
  ])('muestra la página pública de %s', (ruta, titulo) => {
    renderEn(ruta)
    expect(screen.getByRole('heading', { level: 1, name: titulo })).toBeInTheDocument()
  })

  it('redirige /postulacion al login si no hay sesión', () => {
    renderEn('/postulacion')
    expect(screen.getByRole('heading', { level: 1, name: /^Iniciar sesión$/ })).toBeInTheDocument()
  })

  it('muestra /postulacion con sesión abierta', () => {
    iniciarSesion('token', 600)
    renderEn('/postulacion')
    expect(
      screen.getByRole('heading', { level: 1, name: /^Formulario de postulación$/ }),
    ).toBeInTheDocument()
  })
})

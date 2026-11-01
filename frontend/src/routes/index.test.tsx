import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AppRoutes from '@/routes'

function renderEn(ruta: string) {
  return render(
    <MemoryRouter initialEntries={[ruta]}>
      <AppRoutes />
    </MemoryRouter>,
  )
}

describe('AppRoutes', () => {
  it.each([
    ['/', /Concurso de Méritos y Oposición/],
    ['/registro', /^Registro$/],
    ['/login', /^Iniciar sesión$/],
    ['/recuperar', /^Recuperar contraseña$/],
    ['/postulacion', /^Formulario de postulación$/],
  ])('muestra la página de %s', (ruta, titulo) => {
    renderEn(ruta)
    expect(screen.getByRole('heading', { level: 1, name: titulo })).toBeInTheDocument()
  })
})

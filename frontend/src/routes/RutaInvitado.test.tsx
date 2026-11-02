import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { cerrarSesion, iniciarSesion } from '@/store/sesion'
import RutaInvitado from './RutaInvitado'

function renderEn(ruta: string) {
  return render(
    <MemoryRouter initialEntries={[ruta]}>
      <Routes>
        <Route path="/" element={<p>inicio</p>} />
        <Route path="/postulacion" element={<p>formulario</p>} />
        <Route element={<RutaInvitado />}>
          <Route path="/login" element={<p>pantalla de login</p>} />
          <Route path="/registro" element={<p>pantalla de registro</p>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  )
}

const jwt = (payload: object) => `x.${btoa(JSON.stringify(payload))}.y`

afterEach(() => cerrarSesion())

describe('RutaInvitado', () => {
  it('sin sesión muestra el login y el registro', () => {
    renderEn('/login')
    expect(screen.getByText('pantalla de login')).toBeInTheDocument()
  })

  it.each(['/login', '/registro'])(
    'un postulante con sesión que abre %s va al formulario',
    (ruta) => {
      iniciarSesion(jwt({ rol: 'postulante' }), 600)
      renderEn(ruta)
      expect(screen.getByText('formulario')).toBeInTheDocument()
    },
  )

  it('un analista con sesión va al inicio mientras no existan las pantallas de EF-13', () => {
    iniciarSesion(jwt({ rol: 'analista' }), 600)
    renderEn('/login')
    expect(screen.getByText('inicio')).toBeInTheDocument()
  })
})

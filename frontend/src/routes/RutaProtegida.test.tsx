import { act, render, screen } from '@testing-library/react'
import { vi } from 'vitest'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom'
import { cerrarSesion, expirarSesion, iniciarSesion, marcarPostulacionCreada } from '@/store/sesion'
import RutaProtegida from './RutaProtegida'

function Login() {
  const { state } = useLocation()
  return (
    <p>
      login desde {state?.desde}
      {state?.expirada ? ' (expirada)' : ''}
    </p>
  )
}

function renderConRoles(roles?: Parameters<typeof RutaProtegida>[0]['roles'], ruta = '/admin') {
  return render(
    <MemoryRouter initialEntries={[ruta]}>
      <Routes>
        <Route path="/" element={<p>inicio</p>} />
        <Route path="/login" element={<Login />} />
        <Route element={<RutaProtegida roles={roles} />}>
          <Route path="/admin" element={<p>panel</p>} />
          <Route path="/postulacion" element={<p>formulario</p>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  )
}

// Deja el store sin sesión y sin marca de expiración de una prueba anterior.
function sinSesionPrevia() {
  iniciarSesion('token', 600)
  cerrarSesion()
}

afterEach(() => cerrarSesion())

describe('RutaProtegida', () => {
  it('sin sesión redirige al login y recuerda la ruta de origen', () => {
    sinSesionPrevia()
    renderConRoles(undefined, '/admin?pestana=2')
    // Texto exacto: sin "(expirada)", porque nadie cerró la sesión por vencimiento.
    expect(screen.getByText('login desde /admin?pestana=2')).toBeInTheDocument()
  })

  it('si la sesión expiró lo indica al login', () => {
    iniciarSesion('token', 600)
    expirarSesion()
    renderConRoles()
    expect(screen.getByText('login desde /admin (expirada)')).toBeInTheDocument()
  })

  it('con sesión deja pasar', () => {
    iniciarSesion('token', 600)
    renderConRoles()
    expect(screen.getByText('panel')).toBeInTheDocument()
  })

  it('con sesión pero sin el rol pedido vuelve al inicio', () => {
    iniciarSesion('token', 600)
    renderConRoles(['administrador'])
    expect(screen.getByText('inicio')).toBeInTheDocument()
  })

  it('con el rol pedido deja pasar', () => {
    iniciarSesion(`x.${btoa(JSON.stringify({ rol: 'administrador' }))}.y`, 600)
    renderConRoles(['administrador'])
    expect(screen.getByText('panel')).toBeInTheDocument()
  })
})

describe('RutaProtegida en el primer ingreso', () => {
  it('solo deja abrir el formulario: otra ruta privada lleva al formulario', () => {
    iniciarSesion('token', 600, { primerIngreso: true })
    renderConRoles(undefined, '/admin')
    expect(screen.getByText('formulario')).toBeInTheDocument()
  })

  it('el formulario sí se abre', () => {
    iniciarSesion('token', 600, { primerIngreso: true })
    renderConRoles(undefined, '/postulacion')
    expect(screen.getByText('formulario')).toBeInTheDocument()
  })

  it('con la postulación creada ya deja abrir las demás rutas privadas', () => {
    iniciarSesion('token', 600, { primerIngreso: true })
    marcarPostulacionCreada()
    renderConRoles(undefined, '/admin')
    expect(screen.getByText('panel')).toBeInTheDocument()
  })
})

describe('RutaProtegida al vencer el token', () => {
  afterEach(() => vi.useRealTimers())

  it('redirige al login sola cuando vence la sesión, sin otra interacción', () => {
    vi.useFakeTimers()
    iniciarSesion('token', 60)
    renderConRoles()
    expect(screen.getByText('panel')).toBeInTheDocument()
    act(() => {
      vi.advanceTimersByTime(61_000)
    })
    expect(screen.getByText('login desde /admin (expirada)')).toBeInTheDocument()
  })
})

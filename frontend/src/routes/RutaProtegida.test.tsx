import { act, render, screen } from '@testing-library/react'
import { vi } from 'vitest'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom'
import { cerrarSesion, iniciarSesion } from '@/store/sesion'
import RutaProtegida from './RutaProtegida'

function Login() {
  const { state } = useLocation()
  return <p>login desde {state?.desde}</p>
}

function renderConRoles(roles?: Parameters<typeof RutaProtegida>[0]['roles']) {
  return render(
    <MemoryRouter initialEntries={['/admin']}>
      <Routes>
        <Route path="/" element={<p>inicio</p>} />
        <Route path="/login" element={<Login />} />
        <Route element={<RutaProtegida roles={roles} />}>
          <Route path="/admin" element={<p>panel</p>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  )
}

afterEach(() => cerrarSesion())

describe('RutaProtegida', () => {
  it('sin sesión redirige al login y recuerda la ruta de origen', () => {
    renderConRoles()
    expect(screen.getByText('login desde /admin')).toBeInTheDocument()
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
    expect(screen.getByText('login desde /admin')).toBeInTheDocument()
  })
})

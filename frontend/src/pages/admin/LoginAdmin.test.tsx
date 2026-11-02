import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { vi, type Mock } from 'vitest'
import AppRoutes from '@/routes'
import { iniciarSesionAdmin } from '@/services/adminAuth'
import { ApiError } from '@/services/errores'
import { cerrarSesion, iniciarSesion, obtenerSesion } from '@/store/sesion'

vi.mock('@/services/adminAuth', () => ({ iniciarSesionAdmin: vi.fn() }))
const mockLogin = iniciarSesionAdmin as Mock

const jwt = (rol: string) => `x.${btoa(JSON.stringify({ rol }))}.y`

function renderEn(ruta: string) {
  return render(
    <MemoryRouter initialEntries={[ruta]}>
      <AppRoutes />
    </MemoryRouter>,
  )
}

const titulo = (nombre: RegExp) => screen.findByRole('heading', { level: 1, name: nombre })

async function ingresar(user: ReturnType<typeof userEvent.setup>, password = 'Admin2026') {
  await user.type(screen.getByLabelText(/^Usuario/), 'administrador')
  await user.type(screen.getByLabelText(/^Contraseña/), password)
  await user.click(screen.getByRole('button', { name: /Ingresar al panel/ }))
}

// Con llaves: si el callback devolviera el mock, Vitest lo ejecutaría como limpieza al final.
beforeEach(() => {
  mockLogin.mockReset()
})
afterEach(() => cerrarSesion())

describe('Login del panel de administración', () => {
  it('guarda la sesión con el rol y entra al panel', async () => {
    mockLogin.mockResolvedValue({
      token: jwt('administrador'),
      expires_in: 3600,
      usuario: 'administrador',
      rol: 'administrador',
    })
    const user = userEvent.setup()
    renderEn('/admin/login')
    await ingresar(user)

    expect(mockLogin).toHaveBeenCalledWith(expect.any(AbortSignal), {
      usuario: 'administrador',
      password: 'Admin2026',
    })
    expect(await titulo(/^Panel de administración$/)).toBeInTheDocument()
    expect(obtenerSesion()?.rol).toBe('administrador')
    expect(screen.getByText('Acceso de administrador')).toBeInTheDocument()
  })

  it('muestra el mensaje de credenciales incorrectas', async () => {
    mockLogin.mockRejectedValue(
      new ApiError(401, 'UNAUTHORIZED', 'sus credenciales de ingreso son incorrectas'),
    )
    const user = userEvent.setup()
    renderEn('/admin/login')
    await ingresar(user, 'mala1234')

    expect(
      await screen.findByText('sus credenciales de ingreso son incorrectas'),
    ).toBeInTheDocument()
    expect(screen.getByLabelText(/^Contraseña/)).toHaveValue('')
    expect(obtenerSesion()).toBeNull()
  })

  it('con el bloqueo muestra la cuenta regresiva y deshabilita el formulario', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      mockLogin.mockRejectedValue(
        new ApiError(
          429,
          'LOCKED',
          'Cuenta bloqueada por intentos fallidos, inténtelo nuevamente en 10 minutos',
          undefined,
          2,
        ),
      )
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
      renderEn('/admin/login')
      await ingresar(user, 'mala1234')

      expect(await screen.findByRole('timer')).toHaveTextContent('00:02')
      expect(screen.getByRole('button', { name: /Ingresar al panel/ })).toBeDisabled()
      expect(screen.getByLabelText(/^Usuario/)).toBeDisabled()

      await act(async () => {
        vi.advanceTimersByTime(2000)
      })
      expect(await screen.findByText('Ya puede intentarlo nuevamente.')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /Ingresar al panel/ })).toBeEnabled()
    } finally {
      vi.useRealTimers()
    }
  })
})

describe('Rutas del panel por rol', () => {
  it('sin sesión, /admin lleva al login del panel', async () => {
    renderEn('/admin')
    expect(await titulo(/^Panel de administración$/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Ingresar al panel/ })).toBeInTheDocument()
  })

  it('un postulante no entra al panel', async () => {
    iniciarSesion(jwt('postulante'), 600)
    renderEn('/admin')
    expect(screen.queryByRole('navigation', { name: 'Menú del panel' })).toBeNull()
    expect(await titulo(/Concurso de Méritos y Oposición/)).toBeInTheDocument()
  })

  it('un analista entra con acceso de solo lectura', () => {
    iniciarSesion(jwt('analista'), 600)
    renderEn('/admin')
    expect(screen.getByRole('navigation', { name: 'Menú del panel' })).toBeInTheDocument()
    expect(screen.getByText('Acceso de solo lectura')).toBeInTheDocument()
  })

  it('un administrador no abre las pantallas del postulante', async () => {
    iniciarSesion(jwt('administrador'), 600)
    renderEn('/postulacion')
    expect(await titulo(/Concurso de Méritos y Oposición/)).toBeInTheDocument()
  })

  it('con sesión de panel abierta, /admin/login lleva al panel', () => {
    iniciarSesion(jwt('analista'), 600)
    renderEn('/admin/login')
    expect(screen.getByRole('navigation', { name: 'Menú del panel' })).toBeInTheDocument()
  })

  it('Cerrar sesión vuelve al login del panel', async () => {
    iniciarSesion(jwt('administrador'), 600)
    const user = userEvent.setup()
    renderEn('/admin')
    await user.click(screen.getByRole('button', { name: /Cerrar sesión/ }))
    expect(obtenerSesion()).toBeNull()
    expect(screen.getByRole('button', { name: /Ingresar al panel/ })).toBeInTheDocument()
  })
})

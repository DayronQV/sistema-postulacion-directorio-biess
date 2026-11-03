import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi, type Mock } from 'vitest'
import { iniciarSesionPostulante } from '@/services/auth'
import { ApiError } from '@/services/errores'
import { cerrarSesion, obtenerSesion } from '@/store/sesion'
import Login from './Login'

vi.mock('@/services/auth', () => ({ iniciarSesionPostulante: vi.fn() }))

const mockLogin = iniciarSesionPostulante as Mock

const credencialesIncorrectas = () =>
  new ApiError(401, 'UNAUTHORIZED', 'sus credenciales de ingreso son incorrectas')
const bloqueo = (segundos: number) =>
  new ApiError(
    429,
    'LOCKED',
    'Cuenta bloqueada por intentos fallidos, inténtelo nuevamente en 10 minutos',
    undefined,
    segundos,
  )

function renderLogin(state?: object) {
  return render(
    <MemoryRouter initialEntries={[{ pathname: '/login', state }]}>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/postulacion" element={<p>formulario de postulación</p>} />
        <Route path="/otra-privada" element={<p>otra ruta privada</p>} />
        <Route path="/recuperar" element={<p>recuperar contraseña</p>} />
      </Routes>
    </MemoryRouter>,
  )
}

const boton = () => screen.getByRole('button', { name: /Iniciar sesión/ })
const campoCedula = () => screen.getByLabelText(/Cédula de identidad/)
const campoPassword = () => screen.getByLabelText(/^Contraseña/)

async function ingresar(user: ReturnType<typeof userEvent.setup>, password = 'Clave12345') {
  await user.clear(campoCedula())
  await user.type(campoCedula(), '1712345675')
  await user.type(campoPassword(), password)
  await user.click(boton())
}

// Con llaves: si el callback devolviera el mock, Vitest lo ejecutaría como limpieza al final.
beforeEach(() => {
  mockLogin.mockReset()
})
afterEach(() => cerrarSesion())

describe('Login', () => {
  it('en el primer ingreso guarda la sesión y lleva obligatoriamente al formulario', async () => {
    mockLogin.mockResolvedValue({ token: 'jwt', expires_in: 3600, primer_ingreso: true })
    const user = userEvent.setup()
    renderLogin({ desde: '/otra-privada' })
    await ingresar(user)

    expect(mockLogin).toHaveBeenCalledWith(expect.any(AbortSignal), {
      cedula: '1712345675',
      password: 'Clave12345',
    })
    expect(await screen.findByText('formulario de postulación')).toBeInTheDocument()
    expect(obtenerSesion()?.token).toBe('jwt')
  })

  it('avisa que la sesión expiró cuando llega desde una ruta protegida', () => {
    renderLogin({ desde: '/otra-privada', expirada: true })
    expect(screen.getByText('Su sesión expiró')).toBeInTheDocument()
  })

  it('sin expiración no muestra el aviso', () => {
    renderLogin({ desde: '/otra-privada', expirada: false })
    expect(screen.queryByText('Su sesión expiró')).not.toBeInTheDocument()
  })

  it('si no es el primer ingreso vuelve a la ruta privada que pidió el login', async () => {
    mockLogin.mockResolvedValue({ token: 'jwt', expires_in: 3600, primer_ingreso: false })
    const user = userEvent.setup()
    renderLogin({ desde: '/otra-privada' })
    await ingresar(user)
    expect(await screen.findByText('otra ruta privada')).toBeInTheDocument()
  })

  it('muestra el mensaje del requerimiento y limpia la contraseña si son incorrectas', async () => {
    mockLogin.mockRejectedValue(credencialesIncorrectas())
    const user = userEvent.setup()
    renderLogin()
    await ingresar(user, 'mala1234')

    expect(
      await screen.findByText('sus credenciales de ingreso son incorrectas'),
    ).toBeInTheDocument()
    expect(campoPassword()).toHaveValue('')
    expect(obtenerSesion()).toBeNull()
  })

  it('tres intentos fallidos muestran el bloqueo y al terminar la cuenta regresiva se puede reintentar', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      mockLogin
        .mockRejectedValueOnce(credencialesIncorrectas())
        .mockRejectedValueOnce(credencialesIncorrectas())
        .mockRejectedValueOnce(bloqueo(3))
        .mockResolvedValueOnce({ token: 'jwt', expires_in: 3600, primer_ingreso: false })
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
      renderLogin()

      await ingresar(user, 'mala1234')
      await screen.findByText('sus credenciales de ingreso son incorrectas')
      await ingresar(user, 'mala1234')
      await ingresar(user, 'mala1234')

      expect(await screen.findByText(/Cuenta bloqueada por intentos fallidos/)).toBeInTheDocument()
      expect(screen.getByRole('timer')).toHaveTextContent('00:03')
      expect(boton()).toBeDisabled()
      expect(campoCedula()).toBeDisabled()
      expect(campoPassword()).toBeDisabled()

      await act(async () => {
        vi.advanceTimersByTime(3000)
      })
      expect(await screen.findByText('Ya puede intentarlo nuevamente.')).toBeInTheDocument()
      expect(boton()).toBeEnabled()

      await ingresar(user)
      expect(await screen.findByText('formulario de postulación')).toBeInTheDocument()
    } finally {
      vi.useRealTimers()
    }
  })

  it('no llama al servidor con una cédula inválida o sin contraseña', async () => {
    const user = userEvent.setup()
    renderLogin()
    await user.type(campoCedula(), '1712345676')
    await user.click(boton())

    expect(await screen.findByText('Cédula no válida')).toBeInTheDocument()
    expect(screen.getByText('Ingrese su contraseña')).toBeInTheDocument()
    expect(mockLogin).not.toHaveBeenCalled()
  })

  it('muestra un error de conexión sin bloquear el formulario', async () => {
    mockLogin.mockRejectedValue(
      new ApiError(0, 'NETWORK_ERROR', 'No se pudo conectar con el servidor.'),
    )
    const user = userEvent.setup()
    renderLogin()
    await ingresar(user)
    expect(await screen.findByText('No se pudo conectar con el servidor.')).toBeInTheDocument()
    expect(boton()).toBeEnabled()
  })

  it('avisa que el registro terminó al llegar desde el registro', () => {
    renderLogin({ registrado: true })
    expect(screen.getByText('Registro exitoso')).toBeInTheDocument()
  })

  it('enlaza a la recuperación de contraseña', async () => {
    const user = userEvent.setup()
    renderLogin()
    await user.click(screen.getByRole('link', { name: '¿Olvidaste tu contraseña?' }))
    expect(screen.getByText('recuperar contraseña')).toBeInTheDocument()
  })
})

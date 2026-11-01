import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom'
import { vi, type Mock } from 'vitest'
import { registrarCuenta, validarIdentidad } from '@/services/auth'
import { ApiError } from '@/services/errores'
import Inicio from './Inicio'
import Registro, { ESPERA_REDIRECCION_MS } from './Registro'

vi.mock('@/services/auth', () => ({
  validarIdentidad: vi.fn(),
  registrarCuenta: vi.fn(),
}))

const mockIdentidad = validarIdentidad as Mock
const mockRegistro = registrarCuenta as Mock

function LoginFalso() {
  const { state } = useLocation()
  return <p>pantalla de login {state?.registrado ? '(registrado)' : ''}</p>
}

function renderApp(ruta = '/registro') {
  return render(
    <MemoryRouter initialEntries={[ruta]}>
      <Routes>
        <Route path="/" element={<Inicio />} />
        <Route path="/registro" element={<Registro />} />
        <Route path="/login" element={<LoginFalso />} />
      </Routes>
    </MemoryRouter>,
  )
}

const boton = (nombre: RegExp) => screen.getByRole('button', { name: nombre })

async function completarIdentidad(user: ReturnType<typeof userEvent.setup>, cedula = '1712345675') {
  await user.type(screen.getByLabelText(/Cédula de identidad/), cedula)
  fireEvent.change(screen.getByLabelText(/Fecha de expedición de la cédula/), {
    target: { value: '2019-05-20' },
  })
  await user.click(boton(/Validar identidad/))
}

async function completarCuenta(user: ReturnType<typeof userEvent.setup>, email = 'ana@correo.com') {
  await user.type(screen.getByLabelText(/Correo electrónico/), email)
  await user.type(screen.getByLabelText(/^Contraseña/), 'Clave12345')
  await user.type(screen.getByLabelText(/Confirmar contraseña/), 'Clave12345')
  await user.click(boton(/Crear cuenta/))
}

beforeEach(() => {
  mockIdentidad.mockReset()
  mockRegistro.mockReset()
})

describe('Inicio', () => {
  it('ofrece iniciar sesión y registrarse', async () => {
    const user = userEvent.setup()
    renderApp('/')
    expect(boton(/Iniciar sesión/)).toBeInTheDocument()
    await user.click(boton(/Registrarse/))
    expect(screen.getByRole('heading', { level: 1, name: 'Registro' })).toBeInTheDocument()
  })
})

describe('Registro en dos pasos', () => {
  it('valida la identidad, crea la cuenta y lleva al login', async () => {
    mockIdentidad.mockResolvedValue('ticket-1')
    mockRegistro.mockResolvedValue('id-1')
    const user = userEvent.setup()
    renderApp()

    expect(screen.getByText(/Paso 1 de 2/)).toBeInTheDocument()
    await completarIdentidad(user)
    expect(mockIdentidad).toHaveBeenCalledWith(expect.any(AbortSignal), {
      cedula: '1712345675',
      fecha_expedicion: '2019-05-20',
    })

    expect(await screen.findByText(/Paso 2 de 2/)).toBeInTheDocument()
    await completarCuenta(user)
    expect(mockRegistro).toHaveBeenCalledWith(expect.any(AbortSignal), {
      ticket: 'ticket-1',
      email: 'ana@correo.com',
      password: 'Clave12345',
    })

    expect(await screen.findByText('Registro exitoso')).toBeInTheDocument()
    await user.click(boton(/Ir a iniciar sesión/))
    expect(screen.getByText(/pantalla de login \(registrado\)/)).toBeInTheDocument()
  })

  it('envía el código dactilar en mayúsculas si se elige esa opción', async () => {
    mockIdentidad.mockResolvedValue('ticket-1')
    const user = userEvent.setup()
    renderApp()

    await user.type(screen.getByLabelText(/Cédula de identidad/), '1712345675')
    await user.click(screen.getByRole('radio', { name: 'Código dactilar' }))
    await user.type(screen.getByLabelText(/Código dactilar/), 'v1234v5678')
    await user.click(boton(/Validar identidad/))

    await waitFor(() =>
      expect(mockIdentidad).toHaveBeenCalledWith(expect.any(AbortSignal), {
        cedula: '1712345675',
        codigo_dactilar: 'V1234V5678',
      }),
    )
  })

  it('no llama al servidor con una cédula inválida', async () => {
    const user = userEvent.setup()
    renderApp()
    await completarIdentidad(user, '1712345676')
    expect(await screen.findByText('Cédula no válida')).toBeInTheDocument()
    expect(mockIdentidad).not.toHaveBeenCalled()
  })

  it('muestra el mensaje del servidor si la información es incorrecta', async () => {
    mockIdentidad.mockRejectedValue(
      new ApiError(401, 'UNAUTHORIZED', 'la información ingresada es incorrecta'),
    )
    const user = userEvent.setup()
    renderApp()
    await completarIdentidad(user)
    expect(await screen.findByText('la información ingresada es incorrecta')).toBeInTheDocument()
    expect(boton(/Validar identidad/)).toBeEnabled()
  })

  it('con el bloqueo muestra la cuenta regresiva, deshabilita el formulario y luego deja reintentar', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      mockIdentidad.mockRejectedValue(
        new ApiError(
          429,
          'LOCKED',
          'registro fallido, inténtelo nuevamente en 10 minutos',
          undefined,
          3,
        ),
      )
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
      renderApp()
      await completarIdentidad(user)

      expect(
        await screen.findByText('registro fallido, inténtelo nuevamente en 10 minutos'),
      ).toBeInTheDocument()
      expect(screen.getByRole('timer')).toHaveTextContent('00:03')
      expect(boton(/Validar identidad/)).toBeDisabled()
      expect(screen.getByLabelText(/Cédula de identidad/)).toBeDisabled()

      await act(async () => {
        vi.advanceTimersByTime(3000)
      })
      expect(await screen.findByText('Ya puede intentarlo nuevamente.')).toBeInTheDocument()
      expect(boton(/Validar identidad/)).toBeEnabled()
      expect(screen.getByLabelText(/Cédula de identidad/)).toBeEnabled()
    } finally {
      vi.useRealTimers()
    }
  })

  it('si la cédula ya tiene cuenta, invita a iniciar sesión', async () => {
    mockIdentidad.mockRejectedValue(
      new ApiError(409, 'CONFLICT', 'La cédula ya se encuentra registrada', {
        cedula: 'La cédula ya se encuentra registrada',
      }),
    )
    const user = userEvent.setup()
    renderApp()
    await completarIdentidad(user)
    expect(await screen.findByRole('link', { name: 'inicie sesión' })).toBeInTheDocument()
  })

  it('valida la contraseña y su confirmación antes de enviar', async () => {
    mockIdentidad.mockResolvedValue('ticket-1')
    const user = userEvent.setup()
    renderApp()
    await completarIdentidad(user)
    await screen.findByText(/Paso 2 de 2/)

    await user.type(screen.getByLabelText(/Correo electrónico/), 'ana@correo.com')
    await user.type(screen.getByLabelText(/^Contraseña/), 'corta1')
    expect(screen.getByText(/Seguridad: Débil/)).toBeInTheDocument()
    await user.type(screen.getByLabelText(/Confirmar contraseña/), 'otra')
    await user.click(boton(/Crear cuenta/))

    expect(
      await screen.findByText('La contraseña debe tener al menos 8 caracteres'),
    ).toBeInTheDocument()
    expect(screen.getByText('Las contraseñas no coinciden')).toBeInTheDocument()
    expect(mockRegistro).not.toHaveBeenCalled()
  })

  it('muestra junto al campo el correo ya registrado', async () => {
    mockIdentidad.mockResolvedValue('ticket-1')
    mockRegistro.mockRejectedValue(
      new ApiError(409, 'CONFLICT', 'El correo electrónico ya se encuentra registrado', {
        email: 'El correo electrónico ya se encuentra registrado',
      }),
    )
    const user = userEvent.setup()
    renderApp()
    await completarIdentidad(user)
    await screen.findByText(/Paso 2 de 2/)
    await completarCuenta(user)

    expect(
      await screen.findByText('El correo electrónico ya se encuentra registrado'),
    ).toBeInTheDocument()
    expect(screen.getByText(/Paso 2 de 2/)).toBeInTheDocument()
  })

  it('si el ticket venció vuelve al paso 1 con el aviso', async () => {
    mockIdentidad.mockResolvedValue('ticket-1')
    mockRegistro.mockRejectedValue(
      new ApiError(
        401,
        'UNAUTHORIZED',
        'La validación de identidad expiró. Vuelva a validar su identidad',
      ),
    )
    const user = userEvent.setup()
    renderApp()
    await completarIdentidad(user)
    await screen.findByText(/Paso 2 de 2/)
    await completarCuenta(user)

    expect(
      await screen.findByText('La validación de identidad expiró. Vuelva a validar su identidad'),
    ).toBeInTheDocument()
    expect(screen.getByText(/Paso 1 de 2/)).toBeInTheDocument()
  })

  it('tras el registro redirige solo al login', async () => {
    mockIdentidad.mockResolvedValue('ticket-1')
    mockRegistro.mockResolvedValue('id-1')
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
      renderApp()
      await completarIdentidad(user)
      await screen.findByText(/Paso 2 de 2/)
      await completarCuenta(user)
      await screen.findByText('Registro exitoso')

      await act(async () => {
        vi.advanceTimersByTime(ESPERA_REDIRECCION_MS)
      })
      expect(screen.getByText(/pantalla de login/)).toBeInTheDocument()
    } finally {
      vi.useRealTimers()
    }
  })
})

import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi, type Mock } from 'vitest'
import { restablecerPassword } from '@/services/auth'
import { ApiError } from '@/services/errores'
import Restablecer, { ESPERA_REDIRECCION_MS } from './Restablecer'

vi.mock('@/services/auth', () => ({ restablecerPassword: vi.fn() }))

const mockRestablecer = restablecerPassword as Mock

function renderRestablecer(ruta = '/restablecer?token=tok-123') {
  return render(
    <MemoryRouter initialEntries={[ruta]}>
      <Routes>
        <Route path="/restablecer" element={<Restablecer />} />
        <Route path="/login" element={<p>pantalla de login</p>} />
        <Route path="/recuperar" element={<p>pantalla de recuperar</p>} />
      </Routes>
    </MemoryRouter>,
  )
}

async function guardar(
  user: ReturnType<typeof userEvent.setup>,
  password = 'Nueva12345',
  confirmacion = password,
) {
  await user.type(screen.getByLabelText(/Nueva contraseña/), password)
  await user.type(screen.getByLabelText(/Confirmar contraseña/), confirmacion)
  await user.click(screen.getByRole('button', { name: /Guardar contraseña/ }))
}

beforeEach(() => {
  mockRestablecer.mockReset()
})

describe('Restablecer contraseña', () => {
  it('cambia la contraseña con el token del enlace y lleva al login', async () => {
    mockRestablecer.mockResolvedValue(undefined) // 204 sin cuerpo
    const user = userEvent.setup()
    renderRestablecer()
    await guardar(user)

    expect(mockRestablecer).toHaveBeenCalledWith(expect.any(AbortSignal), {
      token: 'tok-123',
      password: 'Nueva12345',
    })
    expect(await screen.findByText('Contraseña actualizada')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: /Ir a iniciar sesión/ }))
    expect(screen.getByText('pantalla de login')).toBeInTheDocument()
  })

  it('tras el cambio redirige solo al login', async () => {
    mockRestablecer.mockResolvedValue(undefined)
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
      renderRestablecer()
      await guardar(user)
      await screen.findByText('Contraseña actualizada')
      await act(async () => {
        vi.advanceTimersByTime(ESPERA_REDIRECCION_MS)
      })
      expect(screen.getByText('pantalla de login')).toBeInTheDocument()
    } finally {
      vi.useRealTimers()
    }
  })

  it('con el enlace vencido o usado muestra el aviso y permite pedir otro', async () => {
    const mensaje =
      'El enlace para restablecer la contraseña no es válido o expiró. Solicite uno nuevo'
    mockRestablecer.mockRejectedValue(new ApiError(401, 'UNAUTHORIZED', mensaje))
    const user = userEvent.setup()
    renderRestablecer()
    await guardar(user)

    expect(await screen.findByText(mensaje)).toBeInTheDocument()
    expect(screen.queryByLabelText(/Nueva contraseña/)).not.toBeInTheDocument()
    await user.click(screen.getByRole('link', { name: 'Solicitar un nuevo enlace' }))
    expect(screen.getByText('pantalla de recuperar')).toBeInTheDocument()
  })

  it('sin token en la URL no muestra el formulario', () => {
    renderRestablecer('/restablecer')
    expect(screen.getByText('Enlace no válido o vencido')).toBeInTheDocument()
    expect(screen.queryByLabelText(/Nueva contraseña/)).not.toBeInTheDocument()
  })

  it('valida la contraseña y su confirmación antes de enviar', async () => {
    const user = userEvent.setup()
    renderRestablecer()
    await guardar(user, 'áááá1', 'otra')

    expect(
      await screen.findByText('La contraseña debe tener al menos 8 caracteres'),
    ).toBeInTheDocument()
    expect(screen.getByText('Las contraseñas no coinciden')).toBeInTheDocument()
    expect(mockRestablecer).not.toHaveBeenCalled()
  })

  it('muestra junto al campo el error de contraseña del servidor', async () => {
    mockRestablecer.mockRejectedValue(
      new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos', {
        password: 'La contraseña debe combinar letras y números',
      }),
    )
    const user = userEvent.setup()
    renderRestablecer()
    await guardar(user)

    expect(
      await screen.findByText('La contraseña debe combinar letras y números'),
    ).toBeInTheDocument()
    expect(screen.getByLabelText(/Nueva contraseña/)).toBeInTheDocument()
  })
})

import { act, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { vi, type Mock } from 'vitest'
import { solicitarRecuperacion } from '@/services/auth'
import { ApiError } from '@/services/errores'
import Recuperar from './Recuperar'

vi.mock('@/services/auth', () => ({ solicitarRecuperacion: vi.fn() }))

const mockSolicitar = solicitarRecuperacion as Mock

function renderRecuperar() {
  return render(
    <MemoryRouter initialEntries={['/recuperar']}>
      <Routes>
        <Route path="/recuperar" element={<Recuperar />} />
        <Route path="/login" element={<p>pantalla de login</p>} />
      </Routes>
    </MemoryRouter>,
  )
}

const boton = () => screen.getByRole('button', { name: /Enviar enlace/ })

async function solicitar(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText(/Cédula de identidad/), '1802468130')
  fireEvent.change(screen.getByLabelText(/Fecha de expedición de la cédula/), {
    target: { value: '2016-03-22' },
  })
  await user.click(boton())
}

beforeEach(() => {
  mockSolicitar.mockReset()
})

describe('Recuperar contraseña', () => {
  it('envía la identidad y muestra el mensaje neutro de confirmación', async () => {
    mockSolicitar.mockResolvedValue(undefined) // 202 sin cuerpo
    const user = userEvent.setup()
    renderRecuperar()
    await solicitar(user)

    expect(mockSolicitar).toHaveBeenCalledWith(expect.any(AbortSignal), {
      cedula: '1802468130',
      fecha_expedicion: '2016-03-22',
    })
    expect(await screen.findByText('Solicitud recibida')).toBeInTheDocument()
    expect(
      screen.getByText(/Si la cédula tiene una cuenta, enviamos al correo/),
    ).toBeInTheDocument()
    expect(screen.queryByLabelText(/Cédula de identidad/)).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /Enviar otra solicitud/ }))
    expect(screen.getByLabelText(/Cédula de identidad/)).toBeInTheDocument()
  })

  it('muestra el error si la identidad no coincide', async () => {
    mockSolicitar.mockRejectedValue(
      new ApiError(401, 'UNAUTHORIZED', 'la información ingresada es incorrecta'),
    )
    const user = userEvent.setup()
    renderRecuperar()
    await solicitar(user)
    expect(await screen.findByText('la información ingresada es incorrecta')).toBeInTheDocument()
    expect(screen.queryByText('Solicitud recibida')).not.toBeInTheDocument()
  })

  it('con el bloqueo muestra la cuenta regresiva y luego deja reintentar', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      mockSolicitar.mockRejectedValue(
        new ApiError(
          429,
          'LOCKED',
          'recuperación fallida, inténtelo nuevamente en 10 minutos',
          undefined,
          2,
        ),
      )
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
      renderRecuperar()
      await solicitar(user)

      expect(
        await screen.findByText('recuperación fallida, inténtelo nuevamente en 10 minutos'),
      ).toBeInTheDocument()
      expect(screen.getByRole('timer')).toHaveTextContent('00:02')
      expect(boton()).toBeDisabled()
      expect(screen.getByLabelText(/Cédula de identidad/)).toBeDisabled()

      await act(async () => {
        vi.advanceTimersByTime(2000)
      })
      expect(await screen.findByText('Ya puede intentarlo nuevamente.')).toBeInTheDocument()
      expect(boton()).toBeEnabled()
    } finally {
      vi.useRealTimers()
    }
  })

  it('no llama al servidor con una cédula inválida', async () => {
    const user = userEvent.setup()
    renderRecuperar()
    await user.type(screen.getByLabelText(/Cédula de identidad/), '1802468131')
    await user.click(boton())
    expect(await screen.findByText('Cédula no válida')).toBeInTheDocument()
    expect(mockSolicitar).not.toHaveBeenCalled()
  })
})

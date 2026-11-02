import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { vi } from 'vitest'
import { ApiError } from '@/services/errores'
import type { Postulacion } from '@/services/postulacion'
import * as store from '../store'
import Cierre from './Cierre'

vi.mock('../store', async (original) => ({
  ...(await original<typeof import('../store')>()),
  enviarFormulario: vi.fn(),
}))
const enviarFormulario = vi.mocked(store.enviarFormulario)

function montar(
  datos: Partial<Postulacion> = {},
  extra: { incompletas?: { id: string; titulo: string }[] } = {},
) {
  const onGuardar = vi.fn().mockResolvedValue(true)
  const onCambios = vi.fn()
  const onIrA = vi.fn()
  render(
    <MemoryRouter>
      <Cierre
        titulo="Cierre"
        datos={{ codigo: 'BIESS-2026-000001', estado: 'borrador', ...datos } as Postulacion}
        guardando={false}
        onGuardar={onGuardar}
        onCambios={onCambios}
        onIrA={onIrA}
        {...extra}
      />
    </MemoryRouter>,
  )
  return { onGuardar, onCambios, onIrA }
}

// Con llaves: si el callback devolviera el mock, Vitest lo ejecutaría como limpieza al final.
beforeEach(() => {
  enviarFormulario.mockReset()
})

describe('Cierre', () => {
  it('muestra los cinco numerales del consentimiento', () => {
    montar()
    expect(screen.getAllByRole('listitem')).toHaveLength(5)
    expect(screen.getByText(/Autorizo de manera libre y voluntaria/)).toBeInTheDocument()
    expect(screen.getByText(/debe ser impreso y con firma manuscrita/)).toBeInTheDocument()
  })

  it('no guarda sin aceptar y avisa', async () => {
    const { onGuardar } = montar()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(screen.getByText('Debe aceptar para guardar la postulación')).toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('al aceptar y guardar manda consentimiento: true', async () => {
    const { onGuardar, onCambios } = montar()
    await userEvent.click(screen.getByRole('checkbox'))
    expect(onCambios).toHaveBeenLastCalledWith(true)
    await userEvent.click(screen.getByRole('button', { name: 'Guardar Postulación' }))

    expect(onGuardar).toHaveBeenCalledWith({ consentimiento: true })
  })

  it('no deja enviar mientras no se guarde la aceptación', () => {
    montar()
    expect(screen.getByRole('button', { name: 'Enviar Postulación' })).toBeDisabled()
    expect(screen.getByText(/Guarde la postulación con su aceptación/)).toBeInTheDocument()
  })

  it('con secciones incompletas las lista con un enlace a cada una y no deja enviar', async () => {
    const { onIrA } = montar(
      { consentimiento: true },
      { incompletas: [{ id: 'cargo', titulo: 'Selección del cargo' }] },
    )
    expect(screen.getByRole('button', { name: 'Enviar Postulación' })).toBeDisabled()

    await userEvent.click(screen.getByRole('button', { name: 'Selección del cargo' }))
    expect(onIrA).toHaveBeenCalledWith('cargo')
  })

  it('pide confirmar antes de enviar y no envía si se cancela', async () => {
    montar({ consentimiento: true })
    await userEvent.click(screen.getByRole('button', { name: 'Enviar Postulación' }))

    expect(screen.getByText(/no podrá editar su postulación/)).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }))
    expect(enviarFormulario).not.toHaveBeenCalled()
  })

  it('al confirmar envía la postulación', async () => {
    enviarFormulario.mockResolvedValue(null)
    montar({ consentimiento: true })
    await userEvent.click(screen.getByRole('button', { name: 'Enviar Postulación' }))
    await userEvent.click(screen.getByRole('button', { name: 'Enviar postulación' }))

    expect(enviarFormulario).toHaveBeenCalledTimes(1)
  })

  it('si el servidor dice que faltan secciones, las muestra con su enlace', async () => {
    enviarFormulario.mockResolvedValue(
      new ApiError(422, 'VALIDATION_ERROR', 'La postulación está incompleta', {
        capacitaciones: 'Agregue al menos una capacitación',
      }),
    )
    const { onIrA } = montar({ consentimiento: true })
    await userEvent.click(screen.getByRole('button', { name: 'Enviar Postulación' }))
    await userEvent.click(screen.getByRole('button', { name: 'Enviar postulación' }))

    expect(await screen.findByText('La postulación está incompleta')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Capacitación' }))
    expect(onIrA).toHaveBeenCalledWith('capacitacion')
  })

  it('enlaza también Información adicional y Publicaciones cuando faltan sus filas', async () => {
    enviarFormulario.mockResolvedValue(
      new ApiError(422, 'VALIDATION_ERROR', 'La postulación está incompleta', {
        familiares: 'Agregue al menos un familiar o responda que no tiene',
        publicaciones: 'Agregue al menos una publicación o responda que no tiene',
      }),
    )
    const { onIrA } = montar({ consentimiento: true })
    await userEvent.click(screen.getByRole('button', { name: 'Enviar Postulación' }))
    await userEvent.click(screen.getByRole('button', { name: 'Enviar postulación' }))

    await userEvent.click(await screen.findByRole('button', { name: 'Información adicional' }))
    expect(onIrA).toHaveBeenCalledWith('adicional')
    await userEvent.click(screen.getByRole('button', { name: 'Publicaciones' }))
    expect(onIrA).toHaveBeenCalledWith('publicaciones')
  })

  it('si ya fue enviada muestra el éxito con el código y el enlace a imprimir', () => {
    montar({ estado: 'enviada', consentimiento: true })

    expect(screen.getByText('Postulación enviada')).toBeInTheDocument()
    expect(screen.getByText('BIESS-2026-000001')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Ver e imprimir/ })).toHaveAttribute(
      'href',
      '/postulacion/imprimir',
    )
    expect(screen.queryByRole('button', { name: 'Enviar Postulación' })).not.toBeInTheDocument()
  })
})

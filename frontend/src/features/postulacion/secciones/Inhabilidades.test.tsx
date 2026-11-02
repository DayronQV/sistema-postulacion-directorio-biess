import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { PREGUNTAS_INHABILIDADES } from '../inhabilidades'
import type { Postulacion } from '@/services/postulacion'
import Inhabilidades from './Inhabilidades'

function montar(onGuardar = vi.fn().mockResolvedValue(true), datos = {} as Postulacion) {
  const onCambios = vi.fn()
  render(
    <Inhabilidades
      titulo="Inhabilidades"
      datos={datos}
      guardando={false}
      onGuardar={onGuardar}
      onCambios={onCambios}
    />,
  )
  return { onGuardar, onCambios }
}

async function responder(n: number, valor: 'Sí' | 'No') {
  const grupo = screen.getByRole('radiogroup', { name: new RegExp(`^${n}\\. `) })
  await userEvent.click(within(grupo).getByRole('radio', { name: valor }))
}

async function responderTodasBien() {
  for (let n = 1; n <= 13; n++) await responder(n, n === 1 ? 'Sí' : 'No')
}

describe('sección Inhabilidades', () => {
  it('muestra las 13 preguntas con su texto', () => {
    montar()
    expect(screen.getAllByRole('radiogroup')).toHaveLength(13)
    expect(
      screen.getByRole('radiogroup', { name: `1. ${PREGUNTAS_INHABILIDADES[0]}` }),
    ).toBeVisible()
  })

  it('no guarda si falta alguna respuesta y señala las faltantes', async () => {
    const { onGuardar } = montar()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar Progreso' }))
    expect(onGuardar).not.toHaveBeenCalled()
    expect(screen.getAllByText('Responda esta pregunta')).toHaveLength(13)
  })

  it('guarda las 13 respuestas con las claves del contrato', async () => {
    const { onGuardar } = montar()
    await responderTodasBien()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar Progreso' }))

    expect(onGuardar).toHaveBeenCalledTimes(1)
    const { inhabilidades } = onGuardar.mock.calls[0][0]
    expect(inhabilidades['1']).toBe(true)
    expect(inhabilidades['2']).toBe(false)
    expect(inhabilidades['13']).toBe(false)
    expect(inhabilidades.removido_detalle).toEqual({
      institucion: '',
      cargo: '',
      fecha_fin: '',
      motivo: '',
    })
  })

  it('avisa cuando una respuesta inhabilita pero no bloquea el guardado', async () => {
    const { onGuardar } = montar()
    await responderTodasBien()
    await responder(10, 'Sí')

    expect(screen.getByText(/Revise las preguntas 10\./)).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar Progreso' }))
    expect(onGuardar).toHaveBeenCalledTimes(1)
  })

  it('con la pregunta 7 en Sí exige institución y motivo', async () => {
    const { onGuardar } = montar()
    await responderTodasBien()
    await responder(7, 'Sí')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar Progreso' }))

    expect(onGuardar).not.toHaveBeenCalled()
    expect(screen.getByText('Indique la institución')).toBeInTheDocument()
    expect(screen.getByText('Indique el motivo')).toBeInTheDocument()

    await userEvent.type(screen.getByLabelText(/^Institución/), 'Superintendencia de Bancos')
    await userEvent.type(screen.getByLabelText(/^Motivo de finalización/), 'Fin de contrato')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar Progreso' }))

    expect(onGuardar).toHaveBeenCalledTimes(1)
    expect(onGuardar.mock.calls[0][0].inhabilidades.removido_detalle).toMatchObject({
      institucion: 'Superintendencia de Bancos',
      motivo: 'Fin de contrato',
    })
  })

  it('precarga lo guardado y avisa de cambios sin guardar', async () => {
    const guardadas: Record<string, boolean> = {}
    for (let n = 1; n <= 13; n++) guardadas[String(n)] = n === 1
    const { onCambios } = montar(undefined, { inhabilidades: guardadas } as Postulacion)

    const grupo = screen.getByRole('radiogroup', { name: /^1\. / })
    expect(within(grupo).getByRole('radio', { name: 'Sí' })).toHaveAttribute('aria-checked', 'true')
    expect(onCambios).toHaveBeenLastCalledWith(false)

    await responder(2, 'Sí')
    expect(onCambios).toHaveBeenLastCalledWith(true)
  })
})

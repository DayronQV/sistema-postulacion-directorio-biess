import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { ApiError } from '@/services/errores'
import Referencias from './Referencias'

async function llenarReferencia() {
  await userEvent.type(screen.getByLabelText(/Nombres y apellidos/), 'Ana Pérez')
  await userEvent.type(screen.getByLabelText(/Relación/), 'Amiga')
  await userEvent.type(screen.getByLabelText(/Teléfono/), '0999999999')
  await userEvent.type(screen.getByLabelText(/Correo/), 'ana@example.com')
}

describe('Referencias', () => {
  it('agrega, edita, elimina y guarda referencias', async () => {
    const guardar = vi.fn()
    render(<Referencias valor={[]} onGuardar={guardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Agregar referencia' }))
    await llenarReferencia()
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('Ana Pérez')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    const relacion = screen.getByLabelText(/Relación/)
    await userEvent.clear(relacion)
    await userEvent.type(relacion, 'Compañera')
    await userEvent.click(screen.getByRole('button', { name: 'Actualizar' }))
    expect(screen.getByText('Compañera')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(guardar).toHaveBeenCalledWith([
      expect.objectContaining({ nombres_apellidos: 'Ana Pérez', relacion: 'Compañera' }),
    ])

    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))
    expect(screen.getByText('Aún no hay registros.')).toBeInTheDocument()
  })

  it('con tres referencias oculta Agregar pero deja editar y eliminar', async () => {
    const referencia = {
      nombres_apellidos: 'Ana Pérez',
      relacion: 'Amiga',
      telefono: '0999999999',
      correo: 'ana@example.com',
    }
    render(
      <Referencias
        valor={[
          referencia,
          { ...referencia, correo: 'b@example.com' },
          { ...referencia, correo: 'c@example.com' },
        ]}
        onGuardar={() => undefined}
      />,
    )
    expect(screen.queryByRole('button', { name: 'Agregar referencia' })).not.toBeInTheDocument()
    expect(screen.getByText(/máximo de 3 referencias/)).toBeInTheDocument()
    for (const boton of screen.getAllByRole('button', { name: 'Editar' }))
      expect(boton).toBeEnabled()

    await userEvent.click(screen.getAllByRole('button', { name: 'Eliminar' })[0])
    expect(screen.getAllByRole('button', { name: 'Eliminar' })).toHaveLength(2)
    expect(screen.getByRole('button', { name: 'Agregar referencia' })).toBeEnabled()
  })

  it('valida antes de agregar y quita espacios sobrantes', async () => {
    render(<Referencias valor={[]} onGuardar={() => undefined} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar referencia' }))
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getAllByText('Campo obligatorio')).toHaveLength(4)

    await userEvent.type(screen.getByLabelText(/Nombres y apellidos/), '  Ana Pérez  ')
    await userEvent.type(screen.getByLabelText(/Relación/), 'Amiga')
    await userEvent.type(screen.getByLabelText(/Teléfono/), '12')
    await userEvent.type(screen.getByLabelText(/Correo/), 'sin-arroba')
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('Teléfono no válido')).toBeInTheDocument()
    expect(screen.getByText('Correo no válido')).toBeInTheDocument()

    await userEvent.clear(screen.getByLabelText(/Teléfono/))
    await userEvent.type(screen.getByLabelText(/Teléfono/), '0999999999')
    await userEvent.clear(screen.getByLabelText(/Correo/))
    await userEvent.type(screen.getByLabelText(/Correo/), 'ana@example.com')
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('Ana Pérez')).toBeInTheDocument()
  })

  it('muestra el mensaje y los campos que devuelve el servidor', async () => {
    const guardar = vi.fn().mockRejectedValue(
      new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos', {
        'referencias[0].correo': 'Correo no válido',
      }),
    )
    render(
      <Referencias
        valor={[
          {
            nombres_apellidos: 'Ana',
            relacion: 'Amiga',
            telefono: '0999999999',
            correo: 'a@b.co',
          },
        ]}
        onGuardar={guardar}
      />,
    )
    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    const alerta = await screen.findByRole('alert')
    expect(alerta).toHaveTextContent('Datos inválidos')
    expect(alerta).toHaveTextContent('referencias[0].correo: Correo no válido')
  })
})

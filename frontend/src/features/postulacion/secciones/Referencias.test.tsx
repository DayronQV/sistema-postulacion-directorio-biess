import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
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

  it('muestra validaciones y limita la lista a tres referencias', async () => {
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
    expect(screen.getByRole('button', { name: 'Agregar referencia' })).toBeDisabled()
  })
})

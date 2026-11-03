import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { ApiError } from '@/services/errores'
import Referencias from './Referencias'

// Reemplaza el texto del campo pegándolo de una vez: tipear carácter por carácter hace
// estas pruebas lentas y, con todos los archivos en paralelo, superaban los 5 s del CI.
async function escribir(etiqueta: RegExp, texto: string) {
  const campo = screen.getByLabelText(etiqueta)
  await userEvent.clear(campo)
  await userEvent.click(campo)
  await userEvent.paste(texto)
}

async function llenarReferencia() {
  await escribir(/Nombres y apellidos/, 'Ana Pérez')
  await escribir(/Relación/, 'Amiga')
  await escribir(/Teléfono/, '0999999999')
  await escribir(/Correo/, 'ana@example.com')
}

const ana = {
  nombres_apellidos: 'Ana Pérez',
  relacion: 'Amiga',
  telefono: '0999999999',
  correo: 'ana@example.com',
}

describe('Referencias', () => {
  it('agrega una referencia y la guarda al momento', async () => {
    const guardar = vi.fn()
    render(<Referencias valor={[]} onGuardar={guardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Agregar referencia' }))
    await llenarReferencia()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(guardar).toHaveBeenCalledWith([ana])
    expect(await screen.findByText('Ana Pérez')).toBeInTheDocument()
  })

  it('edita una referencia existente y la guarda al momento', async () => {
    const guardar = vi.fn()
    render(<Referencias valor={[ana]} onGuardar={guardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await escribir(/Relación/, 'Compañera')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(guardar).toHaveBeenCalledWith([{ ...ana, relacion: 'Compañera' }])
    expect(await screen.findByText('Compañera')).toBeInTheDocument()
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

  it('valida antes de guardar y quita espacios sobrantes', async () => {
    const guardar = vi.fn()
    render(<Referencias valor={[]} onGuardar={guardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar referencia' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(screen.getAllByText('Campo obligatorio')).toHaveLength(4)
    expect(guardar).not.toHaveBeenCalled()

    await escribir(/Nombres y apellidos/, '  Ana Pérez  ')
    await escribir(/Relación/, 'Amiga')
    await escribir(/Teléfono/, '12')
    await escribir(/Correo/, 'sin-arroba')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(screen.getByText('Teléfono no válido')).toBeInTheDocument()
    expect(screen.getByText('Correo no válido')).toBeInTheDocument()

    await escribir(/Teléfono/, '0999999999')
    await escribir(/Correo/, 'ana@example.com')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(guardar).toHaveBeenCalledWith([ana])
    expect(await screen.findByText('Ana Pérez')).toBeInTheDocument()
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
    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    const alerta = await screen.findByRole('alert')
    expect(alerta).toHaveTextContent('Datos inválidos')
    expect(alerta).toHaveTextContent('referencias[0].correo: Correo no válido')
    // El formulario sigue abierto con los datos de la fila.
    expect(screen.getByLabelText(/Nombres y apellidos/)).toHaveValue('Ana')
  })

  it('con el formulario abierto solo hay Cancelar y Guardar', async () => {
    render(<Referencias valor={[ana]} onGuardar={() => undefined} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar referencia' }))
    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeInTheDocument()
    for (const sobrante of ['Agregar', 'Actualizar', 'Guardar información']) {
      expect(screen.queryByRole('button', { name: sobrante })).not.toBeInTheDocument()
    }
  })

  it('Guardar información solo aparece con algo pendiente (tras eliminar)', async () => {
    const guardar = vi.fn()
    render(<Referencias valor={[ana]} onGuardar={guardar} />)
    const boton = () => screen.queryByRole('button', { name: 'Guardar información' })
    expect(boton()).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))
    expect(screen.getByText('Aún no hay registros.')).toBeInTheDocument()
    expect(guardar).not.toHaveBeenCalled()
    // Con el formulario abierto el botón se oculta y al cancelar vuelve.
    await userEvent.click(screen.getByRole('button', { name: 'Agregar referencia' }))
    expect(boton()).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }))

    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(guardar).toHaveBeenCalledWith([])
    await waitFor(() => expect(boton()).not.toBeInTheDocument())
  })

  it('agregar conserva las referencias que ya estaban guardadas', async () => {
    const guardar = vi.fn()
    render(<Referencias valor={[{ ...ana, nombres_apellidos: 'Luis Mora' }]} onGuardar={guardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar referencia' }))
    await llenarReferencia()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(guardar).toHaveBeenCalledWith([{ ...ana, nombres_apellidos: 'Luis Mora' }, ana])
  })
})

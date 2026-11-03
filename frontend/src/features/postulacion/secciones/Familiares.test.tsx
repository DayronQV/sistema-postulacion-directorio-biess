import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { ApiError } from '@/services/errores'
import type { Familiar } from '@/services/postulacion'
import Familiares, { type ValoresFamiliares } from './Familiares'
import ContenedorToasts from '@/components/ui/ContenedorToasts'
import { reiniciarToasts } from '@/store/toasts'

const ana: Familiar = {
  nombres_apellidos: 'Ana Pérez',
  parentesco: 'hermano/a',
  institucion: 'IESS',
  area: 'Finanzas',
  cargo: 'Analista',
}
const beto: Familiar = {
  ...ana,
  nombres_apellidos: 'Beto Mora',
  parentesco: 'padre',
  institucion: 'BIESS',
}

function montar(valores: ValoresFamiliares, onGuardar = vi.fn().mockResolvedValue(undefined)) {
  const user = userEvent.setup()
  render(<Familiares valoresIniciales={valores} onGuardar={onGuardar} />)
  return { user, onGuardar }
}

const boton = (nombre: string) => screen.getByRole('button', { name: nombre })
const guardarInformacion = () => screen.queryByRole('button', { name: 'Guardar información' })

// Cambia la respuesta a Sí y vuelve a No: deja la sección con un cambio pendiente de guardar.
async function dejarPendiente(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('radio', { name: 'Sí' }))
  await user.click(screen.getByRole('radio', { name: 'No' }))
}

async function completarFila(user: ReturnType<typeof userEvent.setup>, f: Partial<Familiar> = {}) {
  const fila = { ...ana, ...f }
  const modal = within(screen.getByRole('dialog'))
  await user.clear(modal.getByLabelText(/Nombres y apellidos/))
  await user.type(modal.getByLabelText(/Nombres y apellidos/), fila.nombres_apellidos)
  await user.selectOptions(modal.getByLabelText(/Parentesco/), fila.parentesco)
  await user.selectOptions(modal.getByLabelText(/Institución/), fila.institucion)
  await user.clear(modal.getByLabelText(/Área/))
  await user.type(modal.getByLabelText(/Área/), fila.area)
  await user.clear(modal.getByLabelText(/Cargo/))
  await user.type(modal.getByLabelText(/Cargo/), fila.cargo)
}

const nombresEnTabla = () =>
  screen
    .queryAllByRole('row')
    .slice(1)
    .map((r) => within(r).queryAllByRole('cell')[0]?.textContent)

beforeEach(() => {
  // El aviso de guardado es un toast: se monta su contenedor y se limpia entre pruebas.
  reiniciarToasts()
  render(<ContenedorToasts />)
})

describe('Familiares', () => {
  it('con "No" no muestra la tabla y no ofrece Guardar información hasta cambiar algo', async () => {
    const { user, onGuardar } = montar({ tiene_familiares: false, familiares: [] })
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'No' })).toBeChecked()
    expect(guardarInformacion()).not.toBeInTheDocument()

    await dejarPendiente(user)
    await user.click(boton('Guardar información'))

    await waitFor(() =>
      expect(onGuardar).toHaveBeenCalledWith({ tiene_familiares: false, familiares: [] }),
    )
    expect(await screen.findByText('La información de familiares se guardó.')).toBeInTheDocument()
    await waitFor(() => expect(guardarInformacion()).not.toBeInTheDocument())
  })

  it('con "Sí" muestra la tabla con las filas recibidas, en su orden', () => {
    montar({ tiene_familiares: true, familiares: [beto, ana] })
    expect(nombresEnTabla()).toEqual(['Beto Mora', 'Ana Pérez'])
    expect(screen.getByText('Padre')).toBeInTheDocument()
    expect(guardarInformacion()).not.toBeInTheDocument()
  })

  it('al responder Sí aparece la tabla vacía y al volver a No se oculta', async () => {
    const { user } = montar({ tiene_familiares: false, familiares: [] })
    await user.click(screen.getByRole('radio', { name: 'Sí' }))
    expect(screen.getByText('Aún no agregó familiares.')).toBeInTheDocument()
    // Con "Sí" y sin filas no hay nada que guardar hasta agregar una.
    expect(guardarInformacion()).not.toBeInTheDocument()
    await user.click(screen.getByRole('radio', { name: 'No' }))
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
    expect(guardarInformacion()).toBeInTheDocument()
  })

  it('agrega una fila y la guarda al momento', async () => {
    const { user, onGuardar } = montar({ tiene_familiares: true, familiares: [] })
    await user.click(boton('Agregar familiar'))
    await completarFila(user, { nombres_apellidos: '  Carla Ruiz ', parentesco: 'cónyuge' })
    await user.click(boton('Guardar'))

    await waitFor(() =>
      expect(onGuardar).toHaveBeenCalledWith({
        tiene_familiares: true,
        familiares: [{ ...ana, nombres_apellidos: 'Carla Ruiz', parentesco: 'cónyuge' }],
      }),
    )
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(nombresEnTabla()).toEqual(['Carla Ruiz'])
    expect(guardarInformacion()).not.toBeInTheDocument()
  })

  it('con la ventana abierta solo hay Cancelar y Guardar', async () => {
    const { user } = montar({ tiene_familiares: true, familiares: [ana] })
    await user.click(boton('Agregar familiar'))

    const modal = within(screen.getByRole('dialog'))
    expect(modal.getByRole('button', { name: 'Cancelar' })).toBeInTheDocument()
    expect(modal.getByRole('button', { name: 'Guardar' })).toBeInTheDocument()
    for (const sobrante of ['Aceptar', 'Agregar', 'Guardar información']) {
      expect(screen.queryByRole('button', { name: sobrante, hidden: true })).not.toBeInTheDocument()
    }
  })

  it('Enter en un campo guarda la fila sin enviar ningún formulario externo', async () => {
    const onGuardar = vi.fn().mockResolvedValue(undefined)
    const alEnviar = vi.fn((e: React.FormEvent) => e.preventDefault())
    const user = userEvent.setup()
    render(
      <form onSubmit={alEnviar}>
        <Familiares
          valoresIniciales={{ tiene_familiares: true, familiares: [] }}
          onGuardar={onGuardar}
        />
      </form>,
    )
    await user.click(boton('Agregar familiar'))
    await completarFila(user)
    await user.type(within(screen.getByRole('dialog')).getByLabelText(/Cargo/), '{Enter}')

    await waitFor(() => expect(nombresEnTabla()).toEqual(['Ana Pérez']))
    expect(onGuardar).toHaveBeenCalledTimes(1)
    expect(alEnviar).not.toHaveBeenCalled()
  })

  it('edita una fila conservando su posición y la guarda al momento', async () => {
    const { user, onGuardar } = montar({ tiene_familiares: true, familiares: [ana, beto] })
    await user.click(within(screen.getAllByRole('row')[1]).getByRole('button', { name: 'Editar' }))
    const modal = within(screen.getByRole('dialog'))
    expect(modal.getByLabelText(/Nombres y apellidos/)).toHaveValue('Ana Pérez')
    expect(modal.getByLabelText(/Institución/)).toHaveValue('IESS')

    await completarFila(user, { nombres_apellidos: 'Ana Pérez Mora', cargo: 'Jefa' })
    await user.click(boton('Guardar'))

    await waitFor(() => expect(onGuardar).toHaveBeenCalled())
    expect(onGuardar.mock.calls[0][0].familiares[0]).toMatchObject({
      nombres_apellidos: 'Ana Pérez Mora',
      cargo: 'Jefa',
    })
    await waitFor(() => expect(nombresEnTabla()).toEqual(['Ana Pérez Mora', 'Beto Mora']))
  })

  it('agregar conserva los familiares que ya estaban guardados', async () => {
    const { user, onGuardar } = montar({ tiene_familiares: true, familiares: [ana] })
    await user.click(boton('Agregar familiar'))
    await completarFila(user, beto)
    await user.click(boton('Guardar'))

    await waitFor(() =>
      expect(onGuardar).toHaveBeenCalledWith({
        tiene_familiares: true,
        familiares: [ana, beto],
      }),
    )
  })

  it('cancelar la edición no cambia la fila', async () => {
    const { user } = montar({ tiene_familiares: true, familiares: [ana] })
    await user.click(boton('Editar'))
    await user.type(within(screen.getByRole('dialog')).getByLabelText(/Cargo/), ' extra')
    await user.click(boton('Cancelar'))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(screen.getByText('Analista')).toBeInTheDocument()
  })

  it('elimina una fila solo tras confirmar y deja Guardar información pendiente', async () => {
    const { user, onGuardar } = montar({ tiene_familiares: true, familiares: [ana, beto] })
    await user.click(
      within(screen.getAllByRole('row')[1]).getByRole('button', { name: 'Eliminar' }),
    )
    expect(screen.getByText('¿Eliminar a Ana Pérez de la lista?')).toBeInTheDocument()

    await user.click(boton('Cancelar'))
    expect(nombresEnTabla()).toEqual(['Ana Pérez', 'Beto Mora'])
    expect(guardarInformacion()).not.toBeInTheDocument()

    await user.click(
      within(screen.getAllByRole('row')[1]).getByRole('button', { name: 'Eliminar' }),
    )
    await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Eliminar' }))
    expect(nombresEnTabla()).toEqual(['Beto Mora'])
    expect(onGuardar).not.toHaveBeenCalled()

    // Con la ventana de agregar abierta se oculta, y al cancelar vuelve.
    await user.click(boton('Agregar familiar'))
    expect(guardarInformacion()).not.toBeInTheDocument()
    await user.click(boton('Cancelar'))
    await user.click(boton('Guardar información'))
    await waitFor(() =>
      expect(onGuardar).toHaveBeenCalledWith({ tiene_familiares: true, familiares: [beto] }),
    )
  })

  it('muestra los errores de la fila y no la guarda si falta algún campo', async () => {
    const { user, onGuardar } = montar({ tiene_familiares: true, familiares: [] })
    await user.click(boton('Agregar familiar'))
    await user.click(boton('Guardar'))

    const modal = within(screen.getByRole('dialog'))
    expect(await modal.findByText('Ingrese los nombres y apellidos')).toBeInTheDocument()
    expect(modal.getByText('Seleccione el parentesco')).toBeInTheDocument()
    expect(modal.getByText('Seleccione la institución')).toBeInTheDocument()
    expect(modal.getByText('Ingrese el área')).toBeInTheDocument()
    expect(modal.getByText('Ingrese el cargo')).toBeInTheDocument()
    expect(screen.getByText('Aún no agregó familiares.')).toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('si el servidor rechaza la fila, la ventana sigue abierta y muestra el error', async () => {
    const onGuardar = vi
      .fn()
      .mockRejectedValue(new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos'))
    const { user } = montar({ tiene_familiares: true, familiares: [] }, onGuardar)
    await user.click(boton('Agregar familiar'))
    await completarFila(user)
    await user.click(boton('Guardar'))

    const modal = within(await screen.findByRole('dialog'))
    expect(await modal.findByText('Datos inválidos')).toBeInTheDocument()
    expect(modal.getByLabelText(/Nombres y apellidos/)).toHaveValue('Ana Pérez')
    expect(screen.getByText('Aún no agregó familiares.')).toBeInTheDocument()
  })

  it('con "Sí" y sin filas no ofrece Guardar información ni llama a onGuardar', async () => {
    const { user, onGuardar } = montar({ tiene_familiares: false, familiares: [] })
    await user.click(screen.getByRole('radio', { name: 'Sí' }))
    expect(guardarInformacion()).not.toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('al eliminar la última fila con "Sí" se oculta Guardar información', async () => {
    const { user } = montar({ tiene_familiares: true, familiares: [ana] })
    await user.click(boton('Eliminar'))
    await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Eliminar' }))
    expect(screen.getByText('Aún no agregó familiares.')).toBeInTheDocument()
    expect(guardarInformacion()).not.toBeInTheDocument()
  })

  it('al cambiar a "No" tras tener filas guarda la lista vacía', async () => {
    const { user, onGuardar } = montar({ tiene_familiares: true, familiares: [ana] })
    await user.click(screen.getByRole('radio', { name: 'No' }))
    await user.click(boton('Guardar información'))
    await waitFor(() =>
      expect(onGuardar).toHaveBeenCalledWith({ tiene_familiares: false, familiares: [] }),
    )
  })

  it('muestra el estado de carga y bloquea el botón mientras guarda', async () => {
    let terminar = () => {}
    const onGuardar = vi.fn(() => new Promise<void>((resolver) => (terminar = resolver)))
    const { user } = montar({ tiene_familiares: false, familiares: [] }, onGuardar)

    await dejarPendiente(user)
    await user.click(boton('Guardar información').closest('button')!)
    const guardando = await screen.findByRole('button', { name: /Guardando/ })
    expect(guardando).toBeDisabled()

    terminar()
    expect(await screen.findByText('La información de familiares se guardó.')).toBeInTheDocument()
  })

  it('muestra los errores del servidor en la fila indicada', async () => {
    const onGuardar = vi.fn().mockRejectedValue(
      new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos', {
        'familiares[1].parentesco': 'Valor no permitido',
        'familiares[1].area': 'Campo obligatorio',
      }),
    )
    const { user } = montar({ tiene_familiares: true, familiares: [ana, beto] }, onGuardar)
    await user.click(screen.getByRole('radio', { name: 'No' }))
    await user.click(screen.getByRole('radio', { name: 'Sí' }))
    await user.click(boton('Guardar información'))

    expect(await screen.findByText('Datos inválidos')).toBeInTheDocument()
    const filaBeto = screen.getAllByRole('row')[2]
    expect(within(filaBeto).getByText('Parentesco: Valor no permitido')).toBeInTheDocument()
    expect(within(filaBeto).getByText('Área: Campo obligatorio')).toBeInTheDocument()
    expect(within(screen.getAllByRole('row')[1]).queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.queryByText('La información de familiares se guardó.')).not.toBeInTheDocument()

    // Editar algo deja obsoletos los errores del servidor.
    await user.click(screen.getByRole('radio', { name: 'No' }))
    expect(screen.queryByText('Datos inválidos')).not.toBeInTheDocument()
  })

  it('muestra el error de la sección y los de otros campos', async () => {
    const onGuardar = vi.fn().mockRejectedValue(
      new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos', {
        familiares: 'Agregue al menos un familiar',
        cargo: 'Cargo no válido',
      }),
    )
    const { user } = montar({ tiene_familiares: true, familiares: [ana] }, onGuardar)
    await user.click(screen.getByRole('radio', { name: 'No' }))
    await user.click(screen.getByRole('radio', { name: 'Sí' }))
    await user.click(boton('Guardar información'))

    expect(await screen.findByText('Agregue al menos un familiar')).toBeInTheDocument()
    expect(screen.getByText('cargo: Cargo no válido')).toBeInTheDocument()
  })

  it('muestra un mensaje si el guardado falla sin ApiError y deja volver a intentar', async () => {
    const onGuardar = vi.fn().mockRejectedValueOnce(new Error('boom')).mockResolvedValue(undefined)
    const { user } = montar({ tiene_familiares: false, familiares: [] }, onGuardar)
    await dejarPendiente(user)
    await user.click(boton('Guardar información'))
    expect(await screen.findByText('No se pudo guardar la información.')).toBeInTheDocument()

    await user.click(boton('Guardar información'))
    expect(await screen.findByText('La información de familiares se guardó.')).toBeInTheDocument()
    expect(screen.queryByText('No se pudo guardar la información.')).not.toBeInTheDocument()
  })

  it('el grupo Sí/No toma el nombre de la pregunta', () => {
    montar({ tiene_familiares: false, familiares: [] })
    expect(
      screen.getByRole('radiogroup', { name: /¿Tiene familiares hasta el 2\.º grado/ }),
    ).toBeInTheDocument()
  })
})

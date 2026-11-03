import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import FormularioPregunta from './FormularioPregunta'

function montar(props: Partial<React.ComponentProps<typeof FormularioPregunta>> = {}) {
  const onGuardar = vi.fn()
  const onCerrar = vi.fn()
  render(
    <FormularioPregunta guardando={false} onGuardar={onGuardar} onCerrar={onCerrar} {...props} />,
  )
  return { onGuardar, onCerrar }
}

describe('FormularioPregunta', () => {
  it('parte con dos opciones vacías y no guarda una pregunta incompleta', async () => {
    const { onGuardar } = montar()
    expect(screen.getAllByRole('textbox', { name: /^Opción \d$/ })).toHaveLength(2)

    await userEvent.click(screen.getByRole('button', { name: 'Guardar pregunta' }))

    expect(onGuardar).not.toHaveBeenCalled()
    expect(screen.getByText('Escriba el enunciado')).toBeInTheDocument()
    expect(screen.getByText('Marque la opción correcta')).toBeInTheDocument()
  })

  it('guarda enunciado, opciones y la correcta', async () => {
    const { onGuardar } = montar()
    await userEvent.type(screen.getByLabelText(/Enunciado/), '  ¿Capital de Ecuador? ')
    await userEvent.type(screen.getByRole('textbox', { name: 'Opción 1' }), 'Guayaquil')
    await userEvent.type(screen.getByRole('textbox', { name: 'Opción 2' }), 'Quito')
    await userEvent.click(screen.getByRole('radio', { name: 'Opción 2 es la correcta' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar pregunta' }))

    expect(onGuardar).toHaveBeenCalledWith({
      enunciado: '¿Capital de Ecuador?',
      opciones: ['Guayaquil', 'Quito'],
      respuesta_correcta: 1,
    })
  })

  it('agrega y quita opciones, con un mínimo de dos y un máximo de ocho', async () => {
    montar()
    expect(screen.getByRole('button', { name: 'Quitar la opción 1' })).toBeDisabled()

    for (let i = 0; i < 6; i++) {
      await userEvent.click(screen.getByRole('button', { name: 'Agregar opción' }))
    }
    expect(screen.getAllByRole('textbox', { name: /^Opción \d$/ })).toHaveLength(8)
    expect(screen.getByRole('button', { name: 'Agregar opción' })).toBeDisabled()

    await userEvent.click(screen.getByRole('button', { name: 'Quitar la opción 8' }))
    expect(screen.getAllByRole('textbox', { name: /^Opción \d$/ })).toHaveLength(7)
  })

  it('la marca de la correcta sigue a su opción al quitar una anterior', async () => {
    const { onGuardar } = montar({
      pregunta: { id: 1, enunciado: '¿P?', opciones: ['A', 'B', 'C'], respuesta_correcta: 2 },
    })
    await userEvent.click(screen.getByRole('button', { name: 'Quitar la opción 1' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar pregunta' }))

    expect(onGuardar).toHaveBeenCalledWith({
      enunciado: '¿P?',
      opciones: ['B', 'C'],
      respuesta_correcta: 1,
    })
  })

  it('al quitar la opción correcta hay que marcar otra', async () => {
    const { onGuardar } = montar({
      pregunta: { id: 1, enunciado: '¿P?', opciones: ['A', 'B', 'C'], respuesta_correcta: 0 },
    })
    await userEvent.click(screen.getByRole('button', { name: 'Quitar la opción 1' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar pregunta' }))

    expect(onGuardar).not.toHaveBeenCalled()
    expect(screen.getByText('Marque la opción correcta')).toBeInTheDocument()
  })

  it('rechaza opciones repetidas', async () => {
    const { onGuardar } = montar()
    await userEvent.type(screen.getByLabelText(/Enunciado/), '¿P?')
    await userEvent.type(screen.getByRole('textbox', { name: 'Opción 1' }), 'Igual')
    await userEvent.type(screen.getByRole('textbox', { name: 'Opción 2' }), 'Igual')
    await userEvent.click(screen.getByRole('radio', { name: 'Opción 1 es la correcta' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar pregunta' }))

    expect(onGuardar).not.toHaveBeenCalled()
    expect(screen.getByText('La opción está repetida')).toBeInTheDocument()
  })

  it('muestra los errores que devuelve el servidor', () => {
    montar({
      erroresServidor: { 'opciones[1]': 'Opción duplicada en el servidor' },
      errorGeneral: 'Ya hay evaluaciones iniciadas',
    })
    expect(screen.getByText('Opción duplicada en el servidor')).toBeInTheDocument()
    expect(screen.getByText('Ya hay evaluaciones iniciadas')).toBeInTheDocument()
  })

  it('edita una pregunta existente con sus datos cargados', () => {
    montar({
      pregunta: { id: 5, enunciado: '¿Existente?', opciones: ['Sí', 'No'], respuesta_correcta: 1 },
    })
    expect(screen.getByText('Editar pregunta')).toBeInTheDocument()
    expect(screen.getByLabelText(/Enunciado/)).toHaveValue('¿Existente?')
    expect(screen.getByRole('radio', { name: 'Opción 2 es la correcta' })).toBeChecked()
  })

  it('Cancelar cierra sin guardar', async () => {
    const { onCerrar, onGuardar } = montar()
    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }))

    expect(onCerrar).toHaveBeenCalled()
    expect(onGuardar).not.toHaveBeenCalled()
  })
})

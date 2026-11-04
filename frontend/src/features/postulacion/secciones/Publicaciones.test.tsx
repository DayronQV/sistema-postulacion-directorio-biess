import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { eliminarDocumento } from '@/services/documentos'
import { ApiError } from '@/services/errores'
import type { PublicacionAcademica } from '@/services/postulacion'
import Publicaciones from './Publicaciones'

vi.mock('@/services/documentos', () => ({ eliminarDocumento: vi.fn() }))

// Cada clic entrega un documento nuevo (doc-1, doc-2...).
let siguiente = 0
vi.mock('@/components/form/DocumentoUpload', () => ({
  default: ({ onChange }: { onChange: (valor: unknown) => void }) => (
    <button
      type="button"
      onClick={() => {
        siguiente += 1
        onChange({ documento_id: `doc-${siguiente}`, nombre: 'respaldo.pdf', tamano_bytes: 100 })
      }}
    >
      Adjuntar PDF de prueba
    </button>
  ),
}))

const mockEliminar = eliminarDocumento as Mock

const articulo: PublicacionAcademica = {
  titulo: 'Mercado de valores',
  tipo: 'publicacion',
  medio: 'Revista',
  fecha: '2024-01-15',
  relacionada: true,
  documento_id: 'doc-guardado',
}

// Pega el texto de una vez: tipear carácter por carácter vuelve lentas estas pruebas.
async function escribir(etiqueta: RegExp, texto: string) {
  const campo = screen.getByLabelText(etiqueta)
  await userEvent.click(campo)
  await userEvent.paste(texto)
}

const preguntaPrincipal = () => screen.getByRole('radiogroup', { name: /Cuenta con publicaciones/ })
const preguntaRelacion = () => screen.getByRole('radiogroup', { name: /Guarda relación/ })

beforeEach(() => {
  siguiente = 0
  mockEliminar.mockReset()
  mockEliminar.mockResolvedValue(undefined)
})

describe('Publicaciones', () => {
  it('Agregar lleva la pantalla al formulario y pone el cursor en su primer campo', async () => {
    const desplazar = vi.fn()
    Element.prototype.scrollIntoView = desplazar
    render(<Publicaciones tienePublicaciones={false} valor={[]} onGuardar={vi.fn()} />)
    await userEvent.click(within(preguntaPrincipal()).getByRole('radio', { name: 'Sí' }))
    await userEvent.click(screen.getByRole('button', { name: 'Agregar publicación' }))
    expect(desplazar).toHaveBeenCalledTimes(1)
    expect(screen.getByLabelText(/^Título/)).toHaveFocus()
  })

  it('con Sí y sin filas no ofrece Guardar información: hay que agregar una', async () => {
    const onGuardar = vi.fn()
    render(<Publicaciones tienePublicaciones={false} valor={[]} onGuardar={onGuardar} />)
    await userEvent.click(within(preguntaPrincipal()).getByRole('radio', { name: 'Sí' }))
    expect(screen.queryByRole('button', { name: 'Guardar información' })).not.toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('agrega una publicación con respaldo y la guarda al momento', async () => {
    const onGuardar = vi.fn()
    render(<Publicaciones tienePublicaciones valor={[]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar publicación' }))
    await escribir(/^Título/, 'Mercado de valores')
    await userEvent.selectOptions(screen.getByLabelText(/^Tipo/), 'publicacion')
    await escribir(/^Medio de publicación/, 'Revista')
    fireEvent.change(screen.getByLabelText(/^Fecha/), { target: { value: '2024-01-15' } })
    await userEvent.click(within(preguntaRelacion()).getByRole('radio', { name: 'Sí' }))
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(onGuardar).toHaveBeenCalledWith(true, [{ ...articulo, documento_id: 'doc-1' }])
    // Al guardar, la fila aparece en la tabla y el formulario se cierra.
    expect(await screen.findByText('Revista')).toBeInTheDocument()
    expect(screen.queryByLabelText(/^Medio de publicación/)).not.toBeInTheDocument()
  })

  it('"¿Guarda relación…?" empieza sin responder y es obligatoria', async () => {
    render(<Publicaciones tienePublicaciones valor={[]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar publicación' }))

    for (const opcion of within(preguntaRelacion()).getAllByRole('radio')) {
      expect(opcion).toHaveAttribute('aria-checked', 'false')
    }
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(screen.getByText('Seleccione Sí o No')).toBeInTheDocument()
    expect(screen.getByText('Adjunte el documento de respaldo')).toHaveAttribute('role', 'alert')
  })

  it('al responder No guarda una lista vacía', async () => {
    const onGuardar = vi.fn()
    render(<Publicaciones tienePublicaciones valor={[articulo]} onGuardar={onGuardar} />)
    await userEvent.click(within(preguntaPrincipal()).getByRole('radio', { name: 'No' }))
    expect(screen.queryByRole('button', { name: 'Agregar publicación' })).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(onGuardar).toHaveBeenCalledWith(false, [])
    // Ya quedó guardado: no queda nada pendiente.
    await waitFor(() =>
      expect(screen.queryByRole('button', { name: 'Guardar información' })).not.toBeInTheDocument(),
    )
  })

  it('marcar No por error y volver a Sí no pierde las filas sin guardar', async () => {
    render(<Publicaciones tienePublicaciones valor={[articulo]} onGuardar={vi.fn()} />)
    await userEvent.click(within(preguntaPrincipal()).getByRole('radio', { name: 'No' }))
    await userEvent.click(within(preguntaPrincipal()).getByRole('radio', { name: 'Sí' }))
    expect(screen.getByText('Mercado de valores')).toBeInTheDocument()
  })

  it('al guardar borra los PDF subidos que ninguna fila usa', async () => {
    const onGuardar = vi.fn()
    render(<Publicaciones tienePublicaciones valor={[articulo]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(onGuardar).toHaveBeenCalledWith(true, [{ ...articulo, documento_id: 'doc-2' }])
    await waitFor(() => expect(mockEliminar).toHaveBeenCalledTimes(1))
    expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'doc-1')
  })

  it('con el formulario abierto solo hay Cancelar y Guardar', async () => {
    render(<Publicaciones tienePublicaciones valor={[articulo]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar publicación' }))
    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeInTheDocument()
    for (const sobrante of ['Agregar', 'Actualizar', 'Guardar información']) {
      expect(screen.queryByRole('button', { name: sobrante })).not.toBeInTheDocument()
    }
  })

  it('Guardar información solo aparece con algo pendiente (cambiar la respuesta o eliminar)', async () => {
    const onGuardar = vi.fn()
    const otro = { ...articulo, titulo: 'Otro artículo' }
    render(<Publicaciones tienePublicaciones valor={[articulo, otro]} onGuardar={onGuardar} />)
    const boton = () => screen.queryByRole('button', { name: 'Guardar información' })
    expect(boton()).not.toBeInTheDocument()

    await userEvent.click(screen.getAllByRole('button', { name: 'Eliminar' })[0])
    expect(onGuardar).not.toHaveBeenCalled()
    // Con el formulario abierto se oculta y al cancelar vuelve.
    await userEvent.click(screen.getByRole('button', { name: 'Agregar publicación' }))
    expect(boton()).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(onGuardar).toHaveBeenCalledWith(true, [otro])
  })

  it('al eliminar la última fila con Sí se oculta Guardar información', async () => {
    render(<Publicaciones tienePublicaciones valor={[articulo]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))
    expect(screen.queryByRole('button', { name: 'Guardar información' })).not.toBeInTheDocument()
  })

  it('agregar conserva las publicaciones que ya estaban guardadas', async () => {
    const onGuardar = vi.fn()
    render(
      <Publicaciones
        tienePublicaciones
        valor={[{ ...articulo, titulo: 'Otro artículo' }]}
        onGuardar={onGuardar}
      />,
    )
    await userEvent.click(screen.getByRole('button', { name: 'Agregar publicación' }))
    await escribir(/^Título/, 'Mercado de valores')
    await userEvent.selectOptions(screen.getByLabelText(/^Tipo/), 'publicacion')
    await escribir(/^Medio de publicación/, 'Revista')
    fireEvent.change(screen.getByLabelText(/^Fecha/), { target: { value: '2024-01-15' } })
    await userEvent.click(within(preguntaRelacion()).getByRole('radio', { name: 'Sí' }))
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(onGuardar).toHaveBeenCalledWith(true, [
      { ...articulo, titulo: 'Otro artículo' },
      { ...articulo, documento_id: 'doc-1' },
    ])
  })

  it('si el servidor rechaza el guardado, el formulario sigue abierto y la lista no cambia', async () => {
    const onGuardar = vi.fn().mockRejectedValue(new ApiError(500, 'INTERNAL', 'Error interno'))
    render(<Publicaciones tienePublicaciones valor={[articulo]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await userEvent.clear(screen.getByLabelText(/^Medio de publicación/))
    await escribir(/^Medio de publicación/, 'Gaceta')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('Error interno')).toBeInTheDocument()
    expect(screen.getByLabelText(/^Medio de publicación/)).toHaveValue('Gaceta')
    expect(screen.getByText('Revista')).toBeInTheDocument()
  })
})

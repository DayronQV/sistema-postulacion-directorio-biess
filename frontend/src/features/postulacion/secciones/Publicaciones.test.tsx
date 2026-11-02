import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { eliminarDocumento } from '@/services/documentos'
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
  it('con Sí exige al menos una fila', async () => {
    render(<Publicaciones tienePublicaciones valor={[]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(
      screen.getByText('Agregue al menos una publicación o investigación.'),
    ).toBeInTheDocument()
  })

  it('agrega y guarda una publicación con respaldo', async () => {
    const onGuardar = vi.fn()
    render(<Publicaciones tienePublicaciones valor={[]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar publicación' }))
    await escribir(/^Título/, 'Mercado de valores')
    await userEvent.selectOptions(screen.getByLabelText(/^Tipo/), 'publicacion')
    await escribir(/^Medio de publicación/, 'Revista')
    fireEvent.change(screen.getByLabelText(/^Fecha/), { target: { value: '2024-01-15' } })
    await userEvent.click(within(preguntaRelacion()).getByRole('radio', { name: 'Sí' }))
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))

    expect(onGuardar).toHaveBeenCalledWith(true, [{ ...articulo, documento_id: 'doc-1' }])
  })

  it('"¿Guarda relación…?" empieza sin responder y es obligatoria', async () => {
    render(<Publicaciones tienePublicaciones valor={[]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar publicación' }))

    for (const opcion of within(preguntaRelacion()).getAllByRole('radio')) {
      expect(opcion).toHaveAttribute('aria-checked', 'false')
    }
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
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
    await userEvent.click(screen.getByRole('button', { name: 'Actualizar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))

    expect(onGuardar).toHaveBeenCalledWith(true, [{ ...articulo, documento_id: 'doc-2' }])
    expect(mockEliminar).toHaveBeenCalledTimes(1)
    expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'doc-1')
  })
})

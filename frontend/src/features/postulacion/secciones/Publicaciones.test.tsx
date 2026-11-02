import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import Publicaciones from './Publicaciones'

vi.mock('@/components/form/DocumentoUpload', () => ({
  default: ({ onChange }: { onChange: (valor: unknown) => void }) => <button type="button" onClick={() => onChange({ documento_id: 'doc-1', nombre: 'respaldo.pdf', tamano_bytes: 100 })}>Adjuntar PDF de prueba</button>,
}))

describe('Publicaciones', () => {
  it('con Sí exige al menos una fila', async () => {
    const usuario = userEvent.setup()
    render(<Publicaciones tienePublicaciones valor={[]} onGuardar={vi.fn()} />)
    await usuario.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(screen.getByText('Agregue al menos una publicación o investigación.')).toBeInTheDocument()
  })

  it('agrega y guarda una publicación con respaldo', async () => {
    const usuario = userEvent.setup()
    const onGuardar = vi.fn()
    render(<Publicaciones tienePublicaciones valor={[]} onGuardar={onGuardar} />)
    await usuario.click(screen.getByRole('button', { name: 'Agregar publicación' }))
    await usuario.type(screen.getByLabelText(/^Título/), 'Mercado de valores')
    await usuario.selectOptions(screen.getByLabelText(/^Tipo/), 'publicacion')
    await usuario.type(screen.getByLabelText(/^Medio de publicación/), 'Revista')
    await usuario.type(screen.getByLabelText(/^Fecha/), '2024-01-15')
    await usuario.selectOptions(screen.getByLabelText(/^¿Guarda relación/), 'true')
    await usuario.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await usuario.click(screen.getByRole('button', { name: 'Agregar' }))
    await usuario.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(onGuardar).toHaveBeenCalledWith(true, [expect.objectContaining({ titulo: 'Mercado de valores', documento_id: 'doc-1', relacionada: true })])
  })

  it('al responder No guarda una lista vacía', async () => {
    const usuario = userEvent.setup()
    const onGuardar = vi.fn()
    render(<Publicaciones tienePublicaciones valor={[]} onGuardar={onGuardar} />)
    await usuario.click(screen.getByLabelText('No'))
    await usuario.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(onGuardar).toHaveBeenCalledWith(false, [])
    expect(screen.queryByRole('button', { name: 'Agregar publicación' })).not.toBeInTheDocument()
  })
})

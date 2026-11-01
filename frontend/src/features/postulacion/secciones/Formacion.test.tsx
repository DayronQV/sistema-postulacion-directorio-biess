import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import type { DocumentoSubido } from '@/services/documentos'
import Formacion from './Formacion'

vi.mock('@/components/form/DocumentoUpload', () => ({
  default: ({ onChange }: { onChange: (documento: DocumentoSubido) => void }) => (
    <button
      type="button"
      onClick={() => onChange({ documento_id: 'doc-1', nombre: 'titulo.pdf', tamano_bytes: 1024 })}
    >
      Subir PDF
    </button>
  ),
}))

async function escribir(etiqueta: RegExp, texto: string) {
  const campo = screen.getByLabelText(etiqueta)
  await userEvent.click(campo)
  await userEvent.paste(texto)
}

async function llenarFormacion() {
  await userEvent.selectOptions(screen.getByLabelText(/Nivel de instrucción/), 'maestria')
  await escribir(/Institución educativa/, 'UPSE')
  await userEvent.selectOptions(screen.getByLabelText(/Tipo de institución/), 'nacional')
  await escribir(/Título obtenido/, 'Magíster en Finanzas')
  await userEvent.selectOptions(screen.getByLabelText(/Área de estudios/), 'finanzas')
  await escribir(/Número de registro SENESCYT/, 'REG-123')
  fireEvent.change(screen.getByLabelText(/Fecha de registro SENESCYT/), {
    target: { value: '2025-10-15' },
  })
  await userEvent.click(screen.getByRole('button', { name: 'Subir PDF' }))
}

describe('Formacion', () => {
  it('agrega, edita, elimina y guarda varias formaciones', async () => {
    const guardar = vi.fn()
    render(<Formacion valor={[]} onGuardar={guardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Agregar formación' }))
    await llenarFormacion()
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('Magíster en Finanzas')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    const titulo = screen.getByLabelText(/Título obtenido/)
    await userEvent.clear(titulo)
    await userEvent.type(titulo, 'Maestría en Economía')
    await userEvent.click(screen.getByRole('button', { name: 'Actualizar' }))
    expect(screen.getByText('Maestría en Economía')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(guardar).toHaveBeenCalledWith([
      expect.objectContaining({
        nivel: 'maestria',
        titulo: 'Maestría en Economía',
        documento_id: 'doc-1',
      }),
    ])

    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))
    expect(screen.getByText('Aún no hay registros.')).toBeInTheDocument()
  })

  it('exige listas, fecha y documento antes de agregar', async () => {
    render(<Formacion valor={[]} onGuardar={() => undefined} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar formación' }))
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('Seleccione el nivel de instrucción')).toBeInTheDocument()

    await userEvent.selectOptions(screen.getByLabelText(/Nivel de instrucción/), 'tercer_nivel')
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('El documento de respaldo es obligatorio')).toBeInTheDocument()
  })
})

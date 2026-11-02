import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import Experiencia from './Experiencia'

vi.mock('@/components/form/DocumentoUpload', () => ({
  default: ({ onChange }: { onChange: (valor: unknown) => void }) => (
    <button
      type="button"
      onClick={() => onChange({ documento_id: 'doc-1', nombre: 'respaldo.pdf', tamano_bytes: 100 })}
    >
      Adjuntar PDF de prueba
    </button>
  ),
}))

describe('Experiencia', () => {
  it('agrega una experiencia, calcula su tiempo y la guarda', async () => {
    const usuario = userEvent.setup()
    const onGuardar = vi.fn()
    render(<Experiencia valor={[]} onGuardar={onGuardar} />)

    await usuario.click(screen.getByRole('button', { name: 'Agregar experiencia' }))
    await usuario.selectOptions(screen.getByLabelText(/^Tipo/), 'general')
    await usuario.type(screen.getByLabelText(/^Cargo o puesto/), 'Analista')
    await usuario.type(screen.getByLabelText(/^Institución o empresa/), 'Empresa')
    await usuario.selectOptions(screen.getByLabelText(/^Actividades relacionadas a/), 'finanzas')
    await usuario.type(screen.getByLabelText(/^Fecha de inicio/), '2018-01-10')
    await usuario.type(screen.getByLabelText(/^Fecha de finalización/), '2020-03-15')
    await usuario.type(screen.getByLabelText(/^Descripción de funciones/), 'Análisis financiero')
    await usuario.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await usuario.click(screen.getByRole('button', { name: 'Agregar' }))

    expect(screen.getByText('2 años, 2 meses y 5 días')).toBeInTheDocument()
    await usuario.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(onGuardar).toHaveBeenCalledWith([
      expect.objectContaining({
        cargo: 'Analista',
        documento_id: 'doc-1',
        duracion: { anios: 2, meses: 2, dias: 5 },
      }),
    ])
  })

  it('exige el respaldo y muestra el contador de descripción', async () => {
    const usuario = userEvent.setup()
    render(<Experiencia valor={[]} onGuardar={vi.fn()} />)
    await usuario.click(screen.getByRole('button', { name: 'Agregar experiencia' }))
    expect(screen.getByText('0/1000')).toBeInTheDocument()
    await usuario.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('Adjunte el documento de respaldo')).toBeInTheDocument()
  })
})

import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import type { ExperienciaProfesional } from '@/services/postulacion'
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

// Pega el texto de una vez: tipear carácter por carácter vuelve lentas estas pruebas y,
// con todos los archivos en paralelo, pueden superar los 5 s.
async function escribir(etiqueta: RegExp, texto: string) {
  const campo = screen.getByLabelText(etiqueta)
  await userEvent.click(campo)
  await userEvent.paste(texto)
}

const analista: ExperienciaProfesional = {
  tipo: 'general',
  cargo: 'Analista',
  institucion: 'Empresa',
  fecha_inicio: '2018-01-10',
  fecha_fin: '2020-03-15',
  actividades: 'finanzas',
  descripcion: 'Análisis financiero',
  documento_id: 'doc-1',
  duracion: { anios: 2, meses: 2, dias: 5 },
}

describe('Experiencia', () => {
  it('agrega una experiencia, calcula su tiempo y la guarda', async () => {
    const onGuardar = vi.fn()
    render(<Experiencia valor={[]} onGuardar={onGuardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Agregar experiencia' }))
    await userEvent.selectOptions(screen.getByLabelText(/^Tipo/), 'general')
    await escribir(/^Cargo o puesto/, 'Analista')
    await escribir(/^Institución o empresa/, 'Empresa')
    await userEvent.selectOptions(screen.getByLabelText(/^Actividades relacionadas a/), 'finanzas')
    fireEvent.change(screen.getByLabelText(/^Fecha de inicio/), { target: { value: '2018-01-10' } })
    fireEvent.change(screen.getByLabelText(/^Fecha de finalización/), {
      target: { value: '2020-03-15' },
    })
    await escribir(/^Descripción de funciones/, 'Análisis financiero')
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))

    expect(screen.getByText('2 años, 2 meses y 5 días')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(onGuardar).toHaveBeenCalledWith([analista])
  })

  it('muestra el tiempo de cada fila y el total', () => {
    render(
      <Experiencia
        valor={[
          analista,
          {
            ...analista,
            documento_id: 'doc-2',
            fecha_inicio: '2021-01-01',
            fecha_fin: '2022-02-02',
            duracion: { anios: 1, meses: 1, dias: 1 },
          },
        ]}
        onGuardar={vi.fn()}
      />,
    )
    expect(screen.getByText('1 año, 1 mes y 1 día')).toBeInTheDocument()
    expect(screen.getByText('Tiempo total: 3 años, 3 meses y 6 días')).toBeInTheDocument()
  })

  it('exige el respaldo y muestra el contador de descripción', async () => {
    render(<Experiencia valor={[]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar experiencia' }))
    expect(screen.getByText('0/1000')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('Adjunte el documento de respaldo')).toHaveAttribute('role', 'alert')
  })
})

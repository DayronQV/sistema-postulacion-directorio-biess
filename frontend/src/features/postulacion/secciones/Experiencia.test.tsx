import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { ApiError } from '@/services/errores'
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
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(onGuardar).toHaveBeenCalledWith([analista])
    expect(await screen.findByText('2 años, 2 meses y 5 días')).toBeInTheDocument()
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
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(screen.getByText('Adjunte el documento de respaldo')).toHaveAttribute('role', 'alert')
  })

  it('con el formulario abierto solo hay Cancelar y Guardar', async () => {
    render(<Experiencia valor={[analista]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar experiencia' }))
    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeInTheDocument()
    for (const sobrante of ['Agregar', 'Actualizar', 'Guardar información']) {
      expect(screen.queryByRole('button', { name: sobrante })).not.toBeInTheDocument()
    }
  })

  it('Guardar información solo aparece con algo pendiente (tras eliminar)', async () => {
    const onGuardar = vi.fn()
    render(<Experiencia valor={[analista]} onGuardar={onGuardar} />)
    const boton = () => screen.queryByRole('button', { name: 'Guardar información' })
    expect(boton()).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))
    expect(onGuardar).not.toHaveBeenCalled()
    // Con el formulario abierto se oculta y al cancelar vuelve.
    await userEvent.click(screen.getByRole('button', { name: 'Agregar experiencia' }))
    expect(boton()).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(onGuardar).toHaveBeenCalledWith([])
    await waitFor(() => expect(boton()).not.toBeInTheDocument())
  })

  it('edita una experiencia y la guarda al momento; si falla, el formulario sigue abierto', async () => {
    const onGuardar = vi.fn().mockRejectedValue(new ApiError(500, 'INTERNAL', 'Error interno'))
    render(<Experiencia valor={[analista]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await userEvent.clear(screen.getByLabelText(/^Cargo o puesto/))
    await escribir(/^Cargo o puesto/, 'Gerente')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(onGuardar).toHaveBeenCalledWith([{ ...analista, cargo: 'Gerente' }])
    expect(await screen.findByText('Error interno')).toBeInTheDocument()
    expect(screen.getByLabelText(/^Cargo o puesto/)).toHaveValue('Gerente')
    expect(screen.getByText('Analista')).toBeInTheDocument()
  })

  it('agregar conserva las experiencias que ya estaban guardadas', async () => {
    const onGuardar = vi.fn()
    render(<Experiencia valor={[analista]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar experiencia' }))
    await userEvent.selectOptions(screen.getByLabelText(/^Tipo/), 'direccion')
    await escribir(/^Cargo o puesto/, 'Director')
    await escribir(/^Institución o empresa/, 'Banco')
    await userEvent.selectOptions(screen.getByLabelText(/^Actividades relacionadas a/), 'derecho')
    fireEvent.change(screen.getByLabelText(/^Fecha de inicio/), { target: { value: '2021-01-01' } })
    fireEvent.change(screen.getByLabelText(/^Fecha de finalización/), {
      target: { value: '2021-12-31' },
    })
    await escribir(/^Descripción de funciones/, 'Dirección')
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(onGuardar).toHaveBeenCalledTimes(1)
    const enviadas = onGuardar.mock.calls[0][0]
    expect(enviadas).toHaveLength(2)
    expect(enviadas[0]).toEqual(analista)
    expect(enviadas[1]).toMatchObject({ tipo: 'direccion', cargo: 'Director' })
  })
})

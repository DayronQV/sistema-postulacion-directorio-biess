import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import Alert from './Alert'
import Button from './Button'
import FileUpload from './FileUpload'
import Input from './Input'
import { ModalConfirmacion } from './Modal'
import RadioSiNo from './RadioSiNo'
import TablaFilas from './TablaFilas'

describe('Button', () => {
  it('muestra el texto y responde al clic', async () => {
    const alClic = vi.fn()
    render(<Button onClick={alClic}>Guardar</Button>)
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(alClic).toHaveBeenCalledTimes(1)
  })

  it('en estado cargando se deshabilita y muestra el texto de carga', () => {
    render(
      <Button cargando textoCargando="Ingresando...">
        Ingresar
      </Button>,
    )
    const boton = screen.getByRole('button', { name: 'Ingresando...' })
    expect(boton).toBeDisabled()
    expect(boton).toHaveAttribute('aria-busy', 'true')
  })
})

describe('Input', () => {
  it('permite escribir y marca el estado inválido', async () => {
    render(<Input aria-label="Cédula" invalido />)
    const campo = screen.getByLabelText('Cédula')
    await userEvent.type(campo, '1712345675')
    expect(campo).toHaveValue('1712345675')
    expect(campo).toHaveAttribute('aria-invalid', 'true')
  })
})

describe('Alert', () => {
  it('muestra título y mensaje de error con rol alert', () => {
    render(
      <Alert tipo="error" titulo="Sus credenciales de ingreso son incorrectas">
        Verifique los datos.
      </Alert>,
    )
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Sus credenciales de ingreso son incorrectas',
    )
    expect(screen.getByText('Verifique los datos.')).toBeInTheDocument()
  })

  it('se puede cerrar', async () => {
    const alCerrar = vi.fn()
    render(
      <Alert tipo="info" onCerrar={alCerrar}>
        Aviso
      </Alert>,
    )
    await userEvent.click(screen.getByRole('button', { name: 'Cerrar aviso' }))
    expect(alCerrar).toHaveBeenCalled()
  })
})

describe('RadioSiNo', () => {
  it('devuelve el booleano elegido', async () => {
    const alCambiar = vi.fn()
    render(<RadioSiNo aria-label="Pregunta" valor={null} onChange={alCambiar} />)
    await userEvent.click(screen.getByRole('radio', { name: 'Sí' }))
    await userEvent.click(screen.getByRole('radio', { name: 'No' }))
    expect(alCambiar).toHaveBeenNthCalledWith(1, true)
    expect(alCambiar).toHaveBeenNthCalledWith(2, false)
  })

  it('marca la opción activa', () => {
    render(<RadioSiNo aria-label="Pregunta" valor={false} onChange={() => undefined} />)
    expect(screen.getByRole('radio', { name: 'No' })).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByRole('radio', { name: 'Sí' })).toHaveAttribute('aria-checked', 'false')
  })
})

describe('TablaFilas', () => {
  const columnas = [{ clave: 'nombre', titulo: 'Nombre' }]
  const filas = [
    { id: 1, nombre: 'Ana' },
    { id: 2, nombre: 'Luis' },
  ]

  it('lista filas y dispara agregar, editar y eliminar', async () => {
    const agregar = vi.fn()
    const editar = vi.fn()
    const eliminar = vi.fn()
    render(
      <TablaFilas
        columnas={columnas}
        filas={filas}
        obtenerClave={(f) => f.id}
        onAgregar={agregar}
        onEditar={editar}
        onEliminar={eliminar}
      />,
    )
    expect(screen.getByText('Ana')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    await userEvent.click(screen.getAllByRole('button', { name: 'Editar' })[1])
    await userEvent.click(screen.getAllByRole('button', { name: 'Eliminar' })[0])
    expect(agregar).toHaveBeenCalled()
    expect(editar).toHaveBeenCalledWith(filas[1])
    expect(eliminar).toHaveBeenCalledWith(filas[0])
  })

  it('muestra el texto de vacío', () => {
    render(<TablaFilas columnas={columnas} filas={[]} obtenerClave={() => 0} />)
    expect(screen.getByText('Aún no hay registros.')).toBeInTheDocument()
  })
})

describe('ModalConfirmacion', () => {
  it('no se muestra cerrado y confirma o cancela abierto', async () => {
    const confirmar = vi.fn()
    const cancelar = vi.fn()
    const { rerender } = render(
      <ModalConfirmacion
        abierto={false}
        titulo="Eliminar"
        mensaje="¿Seguro?"
        onConfirmar={confirmar}
        onCancelar={cancelar}
      />,
    )
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()

    rerender(
      <ModalConfirmacion
        abierto
        titulo="Eliminar"
        mensaje="¿Seguro?"
        onConfirmar={confirmar}
        onCancelar={cancelar}
      />,
    )
    expect(screen.getByRole('dialog', { name: 'Eliminar' })).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Confirmar' }))
    await userEvent.keyboard('{Escape}')
    expect(confirmar).toHaveBeenCalled()
    expect(cancelar).toHaveBeenCalled()
  })
})

describe('FileUpload', () => {
  it('rechaza un archivo que no es PDF', async () => {
    const alSeleccionar = vi.fn()
    render(<FileUpload archivo={null} maxMb={2} onSeleccionar={alSeleccionar} />)
    const entrada = document.querySelector('input[type="file"]') as HTMLInputElement
    const noPdf = new File(['x'], 'foto.png', { type: 'image/png' })
    await userEvent.upload(entrada, noPdf, { applyAccept: false })
    expect(screen.getByRole('alert')).toHaveTextContent('Solo se permiten archivos PDF.')
    expect(alSeleccionar).not.toHaveBeenCalled()
  })

  it('acepta un PDF dentro del límite', async () => {
    const alSeleccionar = vi.fn()
    render(<FileUpload archivo={null} maxMb={2} onSeleccionar={alSeleccionar} />)
    const entrada = document.querySelector('input[type="file"]') as HTMLInputElement
    const pdf = new File(['%PDF'], 'titulo.pdf', { type: 'application/pdf' })
    await userEvent.upload(entrada, pdf)
    expect(alSeleccionar).toHaveBeenCalledWith(pdf)
  })
})

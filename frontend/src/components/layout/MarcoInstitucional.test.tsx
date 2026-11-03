import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, vi } from 'vitest'
import MarcoInstitucional, { type ItemMenu } from './MarcoInstitucional'

const menu: ItemMenu[] = [
  { ruta: '/inicio', titulo: 'Inicio', icono: 'home' },
  { ruta: '/fases', titulo: 'Fases', icono: 'event' },
]

// jsdom no trae matchMedia: se simula el ancho de escritorio (>= 1024 px) o el de móvil.
function simularEscritorio(escritorio: boolean) {
  vi.stubGlobal('matchMedia', (consulta: string) => ({
    matches: escritorio,
    media: consulta,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }))
}

function montar(onCerrarSesion = vi.fn()) {
  render(
    <MemoryRouter initialEntries={['/inicio']}>
      <MarcoInstitucional
        menu={menu}
        inicio="/inicio"
        etiquetaMenu="Menú de prueba"
        nombreUsuario="Carlos Andrés"
        detalleUsuario="C.I. 1712345675"
        onCerrarSesion={onCerrarSesion}
      >
        <p>Contenido</p>
      </MarcoInstitucional>
    </MemoryRouter>,
  )
  return { user: userEvent.setup(), onCerrarSesion }
}

const botonMenu = () => screen.getByRole('button', { name: 'Mostrar u ocultar el menú' })
const menuDeslizable = () => screen.queryByRole('button', { name: 'Cerrar el menú' })

afterEach(() => vi.unstubAllGlobals())

describe('MarcoInstitucional', () => {
  it('muestra el menú, el usuario y el contenido', () => {
    montar()
    expect(screen.getByRole('navigation', { name: 'Menú de prueba' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Fases/ })).toHaveAttribute('href', '/fases')
    expect(screen.getByText('Carlos Andrés')).toBeInTheDocument()
    expect(screen.getByText('C.I. 1712345675')).toBeInTheDocument()
    expect(screen.getByText('Contenido')).toBeInTheDocument()
  })

  it('Cerrar sesión avisa al padre', async () => {
    const { user, onCerrarSesion } = montar()
    await user.click(screen.getByRole('button', { name: /Cerrar sesión/ }))
    expect(onCerrarSesion).toHaveBeenCalledTimes(1)
  })

  it('en móvil abre el menú encima y lo cierra con Escape, con el fondo o al navegar', async () => {
    simularEscritorio(false)
    const { user } = montar()

    await user.click(botonMenu())
    expect(menuDeslizable()).toBeInTheDocument()
    await user.keyboard('{Escape}')
    expect(menuDeslizable()).not.toBeInTheDocument()

    await user.click(botonMenu())
    await user.click(menuDeslizable()!)
    expect(menuDeslizable()).not.toBeInTheDocument()

    await user.click(botonMenu())
    // Con el menú abierto hay dos menús (el fijo, oculto por CSS, y el deslizable): el
    // último es el deslizable.
    const enlaces = screen.getAllByRole('link', { name: /Fases/ })
    await user.click(enlaces[enlaces.length - 1])
    expect(menuDeslizable()).not.toBeInTheDocument()
  })

  it('en escritorio el botón oculta la barra fija en vez de abrir el menú encima', async () => {
    simularEscritorio(true)
    const { user } = montar()
    expect(screen.getAllByRole('navigation', { name: 'Menú de prueba' })).toHaveLength(1)

    await user.click(botonMenu())
    expect(menuDeslizable()).not.toBeInTheDocument()
    // La barra fija sigue en el DOM, oculta con la clase hidden (sin lg:block).
    const barra = screen.getByRole('navigation', { name: 'Menú de prueba' }).closest('aside')
    expect(barra).not.toHaveClass('lg:block')

    await user.click(botonMenu())
    expect(barra).toHaveClass('lg:block')
  })
})

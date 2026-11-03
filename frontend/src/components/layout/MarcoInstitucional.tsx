import { useState, type ReactNode } from 'react'
import { Link, NavLink } from 'react-router-dom'
import Icon from '@/components/ui/Icon'
import { cx } from '@/lib/cx'

export interface ItemMenu {
  ruta: string
  titulo: string
  icono: string
}

interface MarcoInstitucionalProps {
  menu: ItemMenu[]
  // Destino del logo.
  inicio: string
  // Nombre accesible del menú lateral.
  etiquetaMenu: string
  nombreUsuario: string
  detalleUsuario?: ReactNode
  onCerrarSesion: () => void
  children: ReactNode
}

// Barra lateral institucional: logo, menú, frase, foto del edificio y las ondas con los
// colores de la bandera.
function BarraLateral({
  menu,
  inicio,
  etiquetaMenu,
  onNavegar,
}: Pick<MarcoInstitucionalProps, 'menu' | 'inicio' | 'etiquetaMenu'> & { onNavegar?: () => void }) {
  return (
    <div className="relative flex h-full flex-col overflow-hidden bg-gradient-to-b from-primary via-primary-container to-[#1f4f8f] text-on-primary">
      <div className="relative z-10 flex flex-col gap-space-lg px-5 pt-6">
        <Link to={inicio} onClick={onNavegar} className="block focus:outline-none">
          <img
            src="/img/logo-biess.png"
            alt="BIESS, Banco del Instituto Ecuatoriano de Seguridad Social"
            className="h-auto w-full"
          />
        </Link>
        <span className="h-px w-24 bg-on-primary/30" />
        <nav aria-label={etiquetaMenu} className="flex flex-col gap-1">
          {menu.map((item) => (
            <NavLink
              key={item.ruta}
              to={item.ruta}
              end
              onClick={onNavegar}
              className={({ isActive }) =>
                cx(
                  'flex items-center gap-3 rounded-xl px-4 py-3 text-body-bold focus:outline-none focus-visible:ring-2 focus-visible:ring-on-primary',
                  isActive
                    ? 'bg-[#2f6fc4] text-on-primary shadow-md'
                    : 'text-on-primary/85 hover:bg-on-primary/10',
                )
              }
            >
              <Icon nombre={item.icono} relleno className="text-[24px]" />
              {item.titulo}
            </NavLink>
          ))}
        </nav>
      </div>

      {/* Foto del edificio teñida y fundida con el azul */}
      <img
        src="/img/edificio-biess.webp"
        alt=""
        className="pointer-events-none absolute bottom-24 left-0 h-[36%] w-auto max-w-none object-contain object-left-bottom opacity-50 mix-blend-luminosity [mask-image:linear-gradient(to_top,black_40%,transparent)]"
      />

      <p className="relative z-10 mt-auto px-5 pb-28 text-[22px] font-medium leading-tight">
        Tu bienestar,
        <br />
        nuestra prioridad
        <span className="mt-2 block h-1 w-16 rounded-full bg-[#fcd116]" />
      </p>

      {/* Ondas con los colores de la bandera al pie: amarillo, azul y rojo */}
      <svg
        aria-hidden="true"
        viewBox="0 0 250 110"
        preserveAspectRatio="none"
        className="pointer-events-none absolute inset-x-0 bottom-0 h-24 w-full"
      >
        <path d="M0 62 C80 14 170 24 250 0 V110 H0 Z" fill="#fcd116" />
        <path d="M0 80 C85 36 170 46 250 22 V110 H0 Z" fill="#0b4ea2" />
        <path d="M0 96 C90 58 170 68 250 46 V110 H0 Z" fill="#d52b1e" />
      </svg>
    </div>
  )
}

// Marco de las pantallas con sesión (postulante y panel de administración): barra lateral
// institucional y barra superior con el usuario y Cerrar sesión. El botón de menú oculta la
// barra lateral en escritorio y la abre encima en móvil.
export default function MarcoInstitucional({
  menu,
  inicio,
  etiquetaMenu,
  nombreUsuario,
  detalleUsuario,
  onCerrarSesion,
  children,
}: MarcoInstitucionalProps) {
  const [menuAbierto, setMenuAbierto] = useState(false)
  const [lateralOculta, setLateralOculta] = useState(false)
  const lateral = { menu, inicio, etiquetaMenu }

  return (
    <div className="flex min-h-screen bg-[#f3f6fb]">
      {/* Escritorio: barra fija */}
      <aside
        className={cx(
          'sticky top-0 hidden h-screen w-[250px] shrink-0',
          !lateralOculta && 'lg:block',
        )}
      >
        <BarraLateral {...lateral} />
      </aside>

      {/* Móvil: barra deslizable */}
      {menuAbierto && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <button
            type="button"
            aria-label="Cerrar el menú"
            className="absolute inset-0 bg-primary/40"
            onClick={() => setMenuAbierto(false)}
          />
          <aside className="relative h-full w-[260px] shadow-xl">
            <BarraLateral {...lateral} onNavegar={() => setMenuAbierto(false)} />
          </aside>
        </div>
      )}

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-40 bg-surface-container-lowest/95 shadow-[0_1px_8px_rgba(0,0,0,0.04)] backdrop-blur">
          <div className="flex h-16 items-center justify-between gap-4 px-margin-mobile sm:px-margin">
            <button
              type="button"
              aria-label="Mostrar u ocultar el menú"
              onClick={() => {
                if (window.matchMedia('(min-width: 1024px)').matches) setLateralOculta((v) => !v)
                else setMenuAbierto(true)
              }}
              className="flex h-10 w-10 items-center justify-center rounded-lg text-primary hover:bg-surface-container focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
            >
              <Icon nombre="menu" className="text-[28px]" />
            </button>
            <div className="flex shrink-0 items-center gap-4">
              <div className="flex items-center gap-2">
                <span className="flex h-9 w-9 items-center justify-center rounded-full bg-primary-container text-on-primary">
                  <Icon nombre="person" relleno className="text-[22px]" />
                </span>
                <span className="hidden flex-col sm:flex">
                  <span className="flex items-center gap-1 text-body-bold text-on-surface">
                    {nombreUsuario}
                    <Icon nombre="expand_more" className="text-[20px] text-on-surface-variant" />
                  </span>
                  {detalleUsuario && (
                    <span className="text-legal-code text-on-surface-variant">
                      {detalleUsuario}
                    </span>
                  )}
                </span>
              </div>
              <span className="hidden h-8 w-px bg-outline-variant sm:block" />
              <button
                type="button"
                onClick={() => onCerrarSesion()}
                className="inline-flex items-center gap-1.5 rounded-xl border border-error px-3 py-2 text-label-md text-error hover:bg-error-container hover:text-on-error-container focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
              >
                <Icon nombre="logout" className="text-[20px]" />
                <span className="hidden sm:inline">Cerrar sesión</span>
              </button>
            </div>
          </div>
        </header>
        <div className="flex flex-1 flex-col">{children}</div>
      </div>
    </div>
  )
}

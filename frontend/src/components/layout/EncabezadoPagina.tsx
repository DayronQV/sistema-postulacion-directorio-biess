import type { ReactNode } from 'react'
import Icon from '@/components/ui/Icon'

interface EncabezadoPaginaProps {
  icono: string
  titulo: string
  descripcion?: ReactNode
  // Datos bajo la descripción (código, estado, avisos cortos).
  detalle?: ReactNode
  // Botones a la derecha, sobre la foto (por ejemplo, "Nueva pregunta").
  acciones?: ReactNode
}

// Encabezado de las pantallas con sesión: panel blanco curvo a la izquierda con el icono, el
// título y la descripción, y la foto del edificio con el cielo a la derecha, con las curvas
// amarilla, azul y roja (los colores de la bandera, de arriba hacia abajo).
export default function EncabezadoPagina({
  icono,
  titulo,
  descripcion,
  detalle,
  acciones,
}: EncabezadoPaginaProps) {
  return (
    <header className="relative min-h-[104px] overflow-hidden rounded-[28px] bg-[#eef3fa] shadow-[0_4px_24px_rgba(0,38,77,0.08)] sm:min-h-[118px]">
      <div className="pointer-events-none absolute inset-y-0 left-[max(52%,600px)] right-0 hidden overflow-hidden bg-gradient-to-b from-[#4f8fd8] via-[#9cc4ef] to-[#e3eefb] md:block">
        <img
          src="/img/edificio-biess.webp"
          alt=""
          className="absolute -bottom-[12%] right-[10%] h-[150%] w-auto max-w-none object-contain"
        />
        <svg
          aria-hidden="true"
          viewBox="0 0 400 170"
          preserveAspectRatio="none"
          className="absolute inset-0 h-full w-full"
        >
          <path d="M240 170 C330 150 385 80 400 0 V170 Z" fill="#fcd116" />
          <path d="M300 170 C360 150 390 100 400 45 V170 Z" fill="#0b4ea2" />
          <path d="M350 170 C382 155 396 125 400 95 V170 Z" fill="#d52b1e" />
        </svg>
      </div>
      <div className="absolute inset-y-0 left-0 w-full rounded-r-[999px] bg-gradient-to-r from-surface-container-lowest via-surface-container-lowest to-surface-container-lowest/90 shadow-[10px_0_30px_rgba(0,38,77,0.10)] md:w-[max(58%,660px)]" />

      <div className="relative z-10 flex min-h-[104px] flex-wrap items-center gap-space-md px-space-lg py-3 sm:min-h-[118px] sm:px-space-xl">
        <span className="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-primary text-on-primary shadow-lg ring-4 ring-surface-container-lowest sm:h-[68px] sm:w-[68px]">
          <Icon nombre={icono} className="text-[30px] sm:text-[36px]" />
        </span>
        <div className="flex min-w-0 flex-1 flex-col gap-0.5 md:max-w-[520px]">
          <h1 className="text-headline-lg-mobile text-primary sm:text-headline-lg">{titulo}</h1>
          {descripcion && <p className="text-body-md text-on-surface">{descripcion}</p>}
          {detalle && <div className="flex flex-wrap items-center gap-3">{detalle}</div>}
        </div>
        {acciones && (
          <div className="flex flex-wrap gap-3 md:absolute md:bottom-3 md:right-6">{acciones}</div>
        )}
      </div>
    </header>
  )
}

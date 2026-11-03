import type { ReactNode } from 'react'
import Icon from '@/components/ui/Icon'

interface TarjetaAccesoProps {
  icono: string
  titulo: string
  descripcion?: ReactNode
  // Línea final entre separadores (por ejemplo, "¿Aún no tiene cuenta? Regístrese").
  pie?: ReactNode
  children: ReactNode
}

// Tarjeta de las pantallas de acceso (login, registro, recuperación): círculo con el icono
// sobre el borde, el título entre dos líneas y el contenido debajo.
export default function TarjetaAcceso({
  icono,
  titulo,
  descripcion,
  pie,
  children,
}: TarjetaAccesoProps) {
  return (
    <section className="relative mt-12 flex flex-col gap-space-lg rounded-3xl bg-surface-container-lowest/95 px-6 pb-space-xl pt-16 shadow-[0_10px_40px_rgba(0,38,77,0.12)] backdrop-blur sm:px-14">
      <span className="absolute -top-12 left-1/2 flex h-24 w-24 -translate-x-1/2 items-center justify-center rounded-full bg-primary-container text-on-primary shadow-lg">
        <Icon nombre={icono} className="text-[52px]" />
      </span>
      <div className="flex flex-col items-center gap-space-xs text-center">
        <div className="flex w-full items-center gap-4">
          <span className="h-px flex-1 bg-outline-variant" />
          <h1 className="text-headline-lg-mobile text-primary sm:text-display">{titulo}</h1>
          <span className="h-px flex-1 bg-outline-variant" />
        </div>
        {descripcion && <div className="text-body-lg text-on-surface-variant">{descripcion}</div>}
      </div>

      {children}

      {pie && (
        <div className="flex items-center gap-4">
          <span className="h-px flex-1 bg-outline-variant" />
          <div className="text-center text-body-md text-on-surface-variant">{pie}</div>
          <span className="h-px flex-1 bg-outline-variant" />
        </div>
      )}
    </section>
  )
}

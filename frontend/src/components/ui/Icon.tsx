import { cx } from '@/lib/cx'

interface IconProps {
  nombre: string
  className?: string
  relleno?: boolean
}

// Icono de Material Symbols (se carga desde Google Fonts en index.html).
export default function Icon({ nombre, className, relleno }: IconProps) {
  return (
    <span
      aria-hidden="true"
      className={cx('material-symbols-outlined', className)}
      style={relleno ? { fontVariationSettings: "'FILL' 1" } : undefined}
    >
      {nombre}
    </span>
  )
}

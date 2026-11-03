import Icon from '@/components/ui/Icon'

interface EncabezadoSeccionProps {
  titulo: string
  icono: string
}

// Franja azul con el icono y el título de una sección del formulario, con el subrayado amarillo.
// La curva blanca de la derecha ocupa el 40 % de la franja: en móvil se oculta y desde tableta
// el título se limita al ancho libre, para que un título largo no quede blanco sobre blanco.
export default function EncabezadoSeccion({ titulo, icono }: EncabezadoSeccionProps) {
  return (
    <div
      aria-hidden="true"
      className="relative overflow-hidden rounded-t-3xl bg-gradient-to-r from-primary via-primary-container to-primary-container/80 px-space-lg py-4 text-on-primary"
    >
      <svg
        aria-hidden="true"
        viewBox="0 0 400 80"
        preserveAspectRatio="none"
        className="pointer-events-none absolute inset-y-0 right-0 hidden h-full w-2/5 sm:block"
      >
        <path d="M120 0 H400 V80 H40 C90 60 110 30 120 0 Z" fill="#ffffff" opacity="0.92" />
      </svg>
      <div className="relative flex items-center gap-4 sm:max-w-[62%]">
        <Icon nombre={icono} relleno className="text-[34px]" />
        <p className="flex flex-col gap-1 text-headline-sm">
          {titulo}
          <span className="block h-1 w-14 rounded-full bg-[#fcd116]" />
        </p>
      </div>
    </div>
  )
}

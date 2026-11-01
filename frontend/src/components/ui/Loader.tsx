import Spinner from './Spinner'

interface LoaderProps {
  texto?: string
}

// Estado de carga para bloques completos (pantalla o sección).
export default function Loader({ texto = 'Cargando...' }: LoaderProps) {
  return (
    <div
      role="status"
      className="flex flex-col items-center justify-center gap-space-sm p-space-xl text-primary-container"
    >
      <Spinner className="h-8 w-8" />
      <span className="text-body-md text-on-surface-variant">{texto}</span>
    </div>
  )
}

import { Link } from 'react-router-dom'

// Logotipo textual del portal (el diseño no usa imagen).
export default function Marca() {
  return (
    <Link to="/" className="flex min-w-0 items-center gap-3 focus:outline-none">
      <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary-container text-headline-sm text-on-primary">
        B
      </span>
      <span className="flex min-w-0 flex-col">
        <span className="truncate text-headline-sm tracking-tight text-primary">
          BIESS{' '}
          <span className="hidden sm:inline">
            • Banco del Instituto Ecuatoriano de Seguridad Social
          </span>
        </span>
        <span className="truncate text-legal-code uppercase text-on-surface-variant">
          Concurso Público de Méritos y Oposición – Directorio
        </span>
      </span>
    </Link>
  )
}

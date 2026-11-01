import { Link } from 'react-router-dom'

interface PaginaVaciaProps {
  titulo: string
  descripcion?: string
}

// Página provisional: cada módulo la reemplaza por su pantalla real.
export default function PaginaVacia({ titulo, descripcion }: PaginaVaciaProps) {
  return (
    <main className="mx-auto flex min-h-screen max-w-xl flex-col justify-center gap-space-md p-margin">
      <h1 className="text-headline-lg text-primary">{titulo}</h1>
      {descripcion && <p className="text-body-lg text-on-surface-variant">{descripcion}</p>}
      <nav className="flex flex-wrap gap-space-md text-body-bold text-primary-container">
        <Link to="/">Inicio</Link>
        <Link to="/registro">Registro</Link>
        <Link to="/login">Iniciar sesión</Link>
        <Link to="/recuperar">Recuperar contraseña</Link>
        <Link to="/postulacion">Formulario</Link>
      </nav>
    </main>
  )
}

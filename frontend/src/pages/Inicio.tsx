import { useNavigate } from 'react-router-dom'
import Button from '@/components/ui/Button'

// Entrada al sistema: las dos opciones del requerimiento, "Iniciar sesión" y "Registrarse".
export default function Inicio() {
  const navigate = useNavigate()
  return (
    <main className="mx-auto flex w-full max-w-xl flex-1 flex-col justify-center gap-space-lg px-margin-mobile py-space-xl sm:px-margin">
      <div className="flex flex-col gap-space-sm text-center">
        <h1 className="text-headline-lg-mobile text-primary sm:text-headline-lg">
          Concurso de Méritos y Oposición del Directorio del BIESS
        </h1>
        <p className="text-body-lg text-on-surface-variant">
          Miembros Principales y Alternos en representación de los Afiliados, Jubilados y
          Empleadores.
        </p>
      </div>
      <div className="flex flex-col gap-space-sm sm:flex-row">
        <Button className="sm:flex-1" icono="login" onClick={() => navigate('/login')}>
          Iniciar sesión
        </Button>
        <Button
          className="sm:flex-1"
          variante="secondary"
          icono="person_add"
          onClick={() => navigate('/registro')}
        >
          Registrarse
        </Button>
      </div>
    </main>
  )
}

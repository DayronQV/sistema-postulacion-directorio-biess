import { useNavigate } from 'react-router-dom'
import Button from '@/components/ui/Button'
import TarjetaAcceso from '@/components/layout/TarjetaAcceso'

// Entrada al sistema: las dos opciones del requerimiento, "Iniciar sesión" y "Registrarse".
export default function Inicio() {
  const navigate = useNavigate()
  return (
    <TarjetaAcceso
      icono="how_to_reg"
      titulo="Concurso de Méritos y Oposición del Directorio del BIESS"
      descripcion="Miembros Principales y Alternos en representación de los Afiliados, Jubilados y Empleadores."
    >
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
    </TarjetaAcceso>
  )
}

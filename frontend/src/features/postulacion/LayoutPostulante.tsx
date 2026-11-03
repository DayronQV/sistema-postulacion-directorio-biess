import { useEffect } from 'react'
import LayoutPrivado from '@/components/layout/LayoutPrivado'
import { cerrarSesion } from '@/store/sesion'
import { nombreCorto } from './informacionPersonal'
import { cargarPostulacionSiFalta, reiniciarFormulario, useFormulario } from './store'

// Marco de las pantallas del postulante con su nombre y cédula, que vienen del Registro Civil
// dentro de la postulación. Mientras carga se ve el texto genérico "Postulante".
export default function LayoutPostulante() {
  const { postulacion, cargando } = useFormulario()

  useEffect(() => {
    void cargarPostulacionSiFalta()
  }, [])

  const persona = !cargando ? postulacion?.informacion_personal : undefined
  const usuario = persona
    ? { nombre: nombreCorto(persona.nombres, persona.apellidos), cedula: persona.cedula }
    : undefined

  // Al salir se limpia la postulación cargada: si otra persona entra en la misma pestaña,
  // no debe ver por un instante el nombre anterior.
  function salir() {
    reiniciarFormulario()
    cerrarSesion()
  }

  return <LayoutPrivado usuario={usuario} onCerrarSesion={salir} />
}

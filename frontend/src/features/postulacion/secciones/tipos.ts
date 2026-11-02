import type { ComponentType } from 'react'
import type { DatosPostulacion, Postulacion } from '@/services/postulacion'

// Lo que el contenedor le entrega a cada sección. Una sección solo conoce sus datos y estos
// callbacks, así cada responsable trabaja en su archivo sin tocar el contenedor.
export interface PropsSeccion {
  titulo: string
  datos: Postulacion
  guardando: boolean
  // Guarda los campos de la sección. Devuelve true si el servidor los aceptó.
  onGuardar: (cambios: Partial<DatosPostulacion>) => Promise<boolean>
  // Avisa al contenedor si la sección tiene cambios sin guardar.
  onCambios: (conCambios: boolean) => void
}

export interface DefSeccion {
  id: string
  titulo: string
  icono: string
  Componente: ComponentType<PropsSeccion>
  // Marca de "sección completa" que se ve en el menú.
  estaCompleta: (p: Postulacion) => boolean
}

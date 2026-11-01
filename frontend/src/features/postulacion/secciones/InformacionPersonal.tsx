import type { ReactNode } from 'react'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import type { ApiError } from '@/services/errores'
import type { InformacionPersonal as InformacionPersonalApi } from '@/services/postulacion'
import { estadoCivilConPareja, formatearFecha } from '../informacionPersonal'

export interface InformacionPersonalProps {
  informacion: InformacionPersonalApi
  // Guarda la postulación. Los datos personales no se editan: el guardado envía el
  // formulario tal como está en el servidor.
  onGuardar: () => void | Promise<unknown>
  guardando?: boolean
  guardado?: boolean
  error?: ApiError | null
}

function Dato({ etiqueta, children }: { etiqueta: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <dt className="text-label-md text-on-surface-variant">{etiqueta}</dt>
      <dd className="text-body-md text-on-surface">{children || '—'}</dd>
    </div>
  )
}

// Sección "Información personal" del formulario de postulación (EF-12): datos precargados
// del Registro Civil, de solo lectura. Los de la pareja solo se muestran si la persona es
// casada o vive en unión de hecho.
export default function InformacionPersonal({
  informacion,
  onGuardar,
  guardando = false,
  guardado = false,
  error = null,
}: InformacionPersonalProps) {
  const conPareja = estadoCivilConPareja(informacion.estado_civil)

  return (
    <Card
      titulo="Información personal"
      descripcion="Estos datos vienen del Registro Civil y no se pueden modificar."
    >
      <div className="flex flex-col gap-space-md">
        {error && <Alert tipo="error" titulo={error.message} />}
        {guardado && <Alert tipo="success">La información personal se guardó.</Alert>}

        <dl className="grid grid-cols-1 gap-space-md sm:grid-cols-2">
          <Dato etiqueta="Nombres">{informacion.nombres}</Dato>
          <Dato etiqueta="Apellidos">{informacion.apellidos}</Dato>
          <Dato etiqueta="Cédula">{informacion.cedula}</Dato>
          <Dato etiqueta="Género">{informacion.genero}</Dato>
          <Dato etiqueta="Fecha de nacimiento">{formatearFecha(informacion.fecha_nacimiento)}</Dato>
          <Dato etiqueta="Nacionalidad">{informacion.nacionalidad}</Dato>
          <Dato etiqueta="Estado civil">{informacion.estado_civil}</Dato>
        </dl>

        {conPareja && (
          <section aria-labelledby="datos-pareja" className="flex flex-col gap-space-sm">
            <h3 id="datos-pareja" className="text-body-bold text-primary">
              Datos de su pareja
            </h3>
            <dl className="grid grid-cols-1 gap-space-md sm:grid-cols-2">
              <Dato etiqueta="Nombres y apellidos de la pareja">
                {informacion.pareja?.nombres_apellidos}
              </Dato>
              <Dato etiqueta="Cédula de la pareja">{informacion.pareja?.cedula}</Dato>
            </dl>
          </section>
        )}

        <div>
          <Button
            onClick={() => void onGuardar()}
            cargando={guardando}
            textoCargando="Guardando..."
          >
            Guardar
          </Button>
        </div>
      </div>
    </Card>
  )
}

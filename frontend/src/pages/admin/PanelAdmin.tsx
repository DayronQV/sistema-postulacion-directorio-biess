import Alert from '@/components/ui/Alert'
import Card from '@/components/ui/Card'
import { useSesion } from '@/store/sesion'
import EncabezadoPagina from '@/components/layout/EncabezadoPagina'

// Página de inicio del panel. Las pantallas de cada módulo (fases, dashboard, nómina,
// matriz, méritos) se agregan al menú del panel (features/admin/menu.ts).
export default function PanelAdmin() {
  const rol = useSesion()?.rol
  return (
    <main className="mx-auto flex w-full max-w-[1400px] flex-1 flex-col gap-space-md p-margin-mobile sm:p-margin">
      <EncabezadoPagina
        icono="dashboard"
        titulo="Panel de administración"
        descripcion="Concurso de Méritos y Oposición del Directorio del BIESS."
      />
      <Card titulo="Bienvenido">
        <div className="flex flex-col gap-space-md text-body-md text-on-surface">
          <p>
            Desde aquí se gestiona el Concurso de Méritos y Oposición del Directorio del BIESS:
            fases del concurso, postulantes, documentos y resultados.
          </p>
          {rol === 'analista' ? (
            <Alert tipo="info" titulo="Acceso de solo lectura">
              Como analista puede consultar la información; las acciones que la modifican están
              reservadas para el administrador.
            </Alert>
          ) : (
            <Alert tipo="info" titulo="Acceso de administrador">
              Puede consultar la información y ejecutar las acciones del concurso.
            </Alert>
          )}
        </div>
      </Card>
    </main>
  )
}

import Alert from '@/components/ui/Alert'
import Card from '@/components/ui/Card'
import { useSesion } from '@/store/sesion'

// Página de inicio del panel. Las pantallas de cada módulo (fases, dashboard, nómina,
// matriz, méritos) se agregan al menú del panel (features/admin/menu.ts).
export default function PanelAdmin() {
  const rol = useSesion()?.rol
  return (
    <main className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-space-lg p-margin-mobile sm:p-margin">
      <h1 className="text-headline-lg text-primary">Panel de administración</h1>
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

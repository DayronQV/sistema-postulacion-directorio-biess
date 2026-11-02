import Alert from '@/components/ui/Alert'
import Card from '@/components/ui/Card'
import type { PropsSeccion } from './tipos'

// Sección provisional: quien tome la issue de la sección la reemplaza en secciones/index.ts.
export default function SeccionPendiente({ titulo }: PropsSeccion) {
  return (
    <Card titulo={titulo}>
      <Alert tipo="info">Esta sección se habilita cuando se complete su módulo.</Alert>
    </Card>
  )
}

import Alert from '@/components/ui/Alert'
import Card from '@/components/ui/Card'

interface PostulacionBloqueadaProps {
  motivo: string
  codigo?: string
}

// Pantalla de bloqueo: reemplaza a todo el formulario cuando la postulación está cerrada
// (por ejemplo, por nacionalidad). No ofrece ningún campo ni botón para continuar.
export default function PostulacionBloqueada({ motivo, codigo }: PostulacionBloqueadaProps) {
  return (
    <Card titulo="Postulación cerrada" descripcion={codigo ? `Código ${codigo}` : undefined}>
      <Alert tipo="error" titulo="No puede continuar con la postulación">
        {motivo}
      </Alert>
    </Card>
  )
}

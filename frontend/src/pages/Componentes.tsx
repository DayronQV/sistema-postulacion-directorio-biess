import { useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import {
  CampoCheckbox,
  CampoFecha,
  CampoSelect,
  CampoSiNo,
  CampoTexto,
  CampoTextarea,
} from '@/components/form/campos'
import Alert from '@/components/ui/Alert'
import Badge from '@/components/ui/Badge'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import FileUpload from '@/components/ui/FileUpload'
import Loader from '@/components/ui/Loader'
import { ModalConfirmacion } from '@/components/ui/Modal'
import TablaFilas from '@/components/ui/TablaFilas'

interface Familiar {
  id: number
  nombres: string
  parentesco: string
  institucion: string
}

const familiaresIniciales: Familiar[] = [
  { id: 1, nombres: 'María Elena Morales Zambrano', parentesco: 'Hermano/a', institucion: 'BIESS' },
  { id: 2, nombres: 'Rodrigo Xavier Enríquez Paz', parentesco: 'Cuñado/a', institucion: 'IESS' },
]

interface Demo {
  cedula: string
  provincia: string
  nacimiento: string
  funciones: string
  tieneFamiliares: boolean
  acepto: boolean
}

// Página interna con todos los componentes, para que el equipo copie ejemplos de uso.
export default function Componentes() {
  const metodos = useForm<Demo>({ mode: 'onTouched' })
  const [enviado, setEnviado] = useState(false)
  const [modal, setModal] = useState(false)
  const [familiares, setFamiliares] = useState(familiaresIniciales)
  const [archivo, setArchivo] = useState<File | null>(null)

  return (
    <main className="mx-auto flex w-full max-w-4xl flex-col gap-space-lg p-margin-mobile sm:p-margin">
      <h1 className="text-headline-lg text-primary">Componentes de interfaz</h1>

      <Card titulo="Botones">
        <div className="flex flex-wrap gap-3">
          <Button icono="save">Guardar progreso</Button>
          <Button variante="secondary">Secundario</Button>
          <Button variante="ghost">Texto</Button>
          <Button variante="danger" icono="delete">
            Eliminar
          </Button>
          <Button cargando textoCargando="Ingresando...">
            Ingresar
          </Button>
          <Button disabled>Deshabilitado</Button>
        </div>
      </Card>

      <Card titulo="Avisos y etiquetas">
        <div className="flex flex-col gap-3">
          <Alert tipo="success" titulo="Guardado">
            Su progreso se guardó correctamente.
          </Alert>
          <Alert tipo="error" titulo="Sus credenciales de ingreso son incorrectas">
            Verifique la cédula y la contraseña.
          </Alert>
          <Alert tipo="warning" titulo="Cuenta bloqueada">
            Inténtelo nuevamente en 09:42.
          </Alert>
          <Alert tipo="info" titulo="Información">
            Puede llenar el formulario por partes.
          </Alert>
          <div className="flex flex-wrap gap-2">
            <Badge>Borrador</Badge>
            <Badge tono="primary">Postulante</Badge>
            <Badge tono="success">Enviada</Badge>
            <Badge tono="warning">En revisión</Badge>
            <Badge tono="error">Bloqueado</Badge>
          </div>
          <Loader texto="Cargando postulación..." />
        </div>
      </Card>

      <Card titulo="Campos de formulario" descripcion="Con react-hook-form y validación visible.">
        <FormProvider {...metodos}>
          <form
            noValidate
            className="grid grid-cols-1 gap-space-md md:grid-cols-2"
            onSubmit={metodos.handleSubmit(() => setEnviado(true))}
          >
            <CampoTexto
              name="cedula"
              etiqueta="Cédula"
              requerido
              ayuda="10 dígitos"
              maxLength={10}
              inputMode="numeric"
              reglas={{
                required: 'La cédula es obligatoria',
                pattern: { value: /^\d{10}$/, message: 'Debe tener 10 dígitos' },
              }}
            />
            <CampoSelect
              name="provincia"
              etiqueta="Provincia"
              requerido
              opciones={[
                { valor: 17, etiqueta: 'Pichincha' },
                { valor: 9, etiqueta: 'Guayas' },
              ]}
              reglas={{ required: 'Seleccione una provincia' }}
            />
            <CampoFecha
              name="nacimiento"
              etiqueta="Fecha"
              requerido
              reglas={{ required: 'Indique la fecha' }}
            />
            <CampoSiNo name="tieneFamiliares" etiqueta="¿Tiene familiares en el IESS?" requerido />
            <div className="md:col-span-2">
              <CampoTextarea
                name="funciones"
                etiqueta="Descripción de funciones"
                maxLength={1000}
                ayuda="Máximo 1000 caracteres"
              />
            </div>
            <div className="md:col-span-2">
              <CampoCheckbox
                name="acepto"
                etiqueta="Acepto la declaración"
                reglas={{ required: 'Debe aceptar la declaración' }}
              />
            </div>
            <div className="md:col-span-2">
              <Button type="submit" cargando={metodos.formState.isSubmitting}>
                Validar
              </Button>
            </div>
            {enviado && (
              <div className="md:col-span-2">
                <Alert tipo="success">Formulario válido.</Alert>
              </div>
            )}
          </form>
        </FormProvider>
      </Card>

      <Card titulo="Subida de archivo">
        <FileUpload
          archivo={archivo ? { nombre: archivo.name, tamano: archivo.size } : null}
          maxMb={2}
          estado={archivo ? 'exito' : 'inactivo'}
          onSeleccionar={setArchivo}
          onQuitar={() => setArchivo(null)}
        />
      </Card>

      <TablaFilas
        titulo="Familiares declarados"
        columnas={[
          { clave: 'nombres', titulo: 'Nombres y apellidos' },
          { clave: 'parentesco', titulo: 'Parentesco' },
          { clave: 'institucion', titulo: 'Institución' },
        ]}
        filas={familiares}
        obtenerClave={(f) => f.id}
        textoAgregar="Agregar familiar"
        onAgregar={() =>
          setFamiliares((f) => [
            ...f,
            {
              id: Date.now(),
              nombres: 'Nuevo familiar',
              parentesco: 'Hijo/a',
              institucion: 'IESS',
            },
          ])
        }
        onEditar={() => undefined}
        onEliminar={() => setModal(true)}
      />

      <ModalConfirmacion
        abierto={modal}
        titulo="Eliminar registro"
        mensaje="¿Desea eliminar este registro? Esta acción no se puede deshacer."
        textoConfirmar="Eliminar"
        peligroso
        onConfirmar={() => {
          setFamiliares((f) => f.slice(1))
          setModal(false)
        }}
        onCancelar={() => setModal(false)}
      />
    </main>
  )
}

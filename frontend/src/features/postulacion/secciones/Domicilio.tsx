import { useEffect, useId, useRef, type ReactNode } from 'react'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Campo from '@/components/form/Campo'
import { CampoTexto } from '@/components/form/campos'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import Loader from '@/components/ui/Loader'
import Select from '@/components/ui/Select'
import { aplicarErroresDeCampo } from '@/features/auth/errores'
import type { ApiError } from '@/services/errores'
import { CAMPOS_API_DOMICILIO, crearEsquemaDomicilio, type ValoresDomicilio } from '../domicilio'
import { useCatalogosGeo, type ListaCatalogo } from '../useCatalogosGeo'

interface DomicilioProps {
  valoresIniciales: ValoresDomicilio
  // Correo del registro: se muestra, no se edita.
  correoPrincipal: string
  onGuardar: (valores: ValoresDomicilio) => void | Promise<unknown>
  guardando?: boolean
  guardado?: boolean
  // Error del servidor al guardar; los errores por campo se muestran junto a cada campo.
  error?: ApiError | null
}

// Sección "Información domiciliaria y de contacto". Es de presentación: recibe los valores
// iniciales y avisa con onGuardar; solo lee por su cuenta los catálogos geográficos.
export default function Domicilio({
  valoresIniciales,
  correoPrincipal,
  onGuardar,
  guardando = false,
  guardado = false,
  error = null,
}: DomicilioProps) {
  // El esquema depende de si el país es Ecuador, que se sabe al cargar el catálogo.
  const esEcuadorRef = useRef(false)
  const form = useForm<ValoresDomicilio>({
    resolver: (valores, contexto, opciones) =>
      zodResolver(crearEsquemaDomicilio(esEcuadorRef.current))(valores, contexto, opciones),
    defaultValues: valoresIniciales,
  })
  const { register, setValue, clearErrors, setError, control, formState } = form
  const { errors } = formState

  const paisId = useWatch({ control, name: 'paisId' })
  const provinciaId = useWatch({ control, name: 'provinciaId' })
  const { paises, provincias, cantones, esEcuador, hayPais } = useCatalogosGeo(paisId, provinciaId)
  useEffect(() => {
    esEcuadorRef.current = esEcuador
  }, [esEcuador])

  useEffect(() => {
    if (error) aplicarErroresDeCampo(error, setError, CAMPOS_API_DOMICILIO)
  }, [error, setError])

  const idPais = useId()
  const idProvincia = useId()
  const idCanton = useId()
  const idCorreo = useId()

  const registroPais = register('paisId')
  const registroProvincia = register('provinciaId')
  const registroCanton = register('cantonId')

  // Al cambiar el país se limpian provincia y cantón (ids y textos); al cambiar la
  // provincia, el cantón. Se hace en el evento y no en un efecto: así los valores que
  // llegan del servidor al abrir la sección no se borran.
  function alCambiarPais() {
    for (const campo of ['provinciaId', 'cantonId', 'provinciaTexto', 'ciudadTexto'] as const) {
      setValue(campo, '')
    }
    clearErrors(['provinciaId', 'cantonId', 'provinciaTexto', 'ciudadTexto'])
  }
  function alCambiarProvincia() {
    setValue('cantonId', '')
    clearErrors('cantonId')
  }

  const enviar = form.handleSubmit(async (valores) => {
    await onGuardar(valores)
  })

  return (
    <FormProvider {...form}>
      <form noValidate onSubmit={enviar} className="flex flex-col gap-space-md">
        <h2 className="text-headline-sm text-on-surface">Información domiciliaria y de contacto</h2>

        {error && (
          <Alert tipo="error" titulo="No se pudo guardar">
            {error.message}
          </Alert>
        )}
        {guardado && !error && <Alert tipo="success">Domicilio guardado.</Alert>}

        <fieldset disabled={guardando} className="flex flex-col gap-space-md">
          <div className="grid gap-space-md sm:grid-cols-2">
            <Campo id={idPais} etiqueta="País" requerido error={errors.paisId?.message}>
              <ListaCargable lista={paises} textoCarga="Cargando países..." nombre="los países">
                <Select
                  id={idPais}
                  opciones={paises.items.map((p) => ({ valor: p.id, etiqueta: p.nombre }))}
                  placeholder="Seleccione el país"
                  invalido={Boolean(errors.paisId)}
                  {...registroPais}
                  onChange={(e) => {
                    void registroPais.onChange(e)
                    alCambiarPais()
                  }}
                />
              </ListaCargable>
            </Campo>

            {hayPais && esEcuador && (
              <>
                <Campo
                  id={idProvincia}
                  etiqueta="Provincia"
                  requerido
                  error={errors.provinciaId?.message}
                >
                  <ListaCargable
                    lista={provincias}
                    textoCarga="Cargando provincias..."
                    nombre="las provincias"
                  >
                    <Select
                      id={idProvincia}
                      opciones={provincias.items.map((p) => ({ valor: p.id, etiqueta: p.nombre }))}
                      placeholder="Seleccione la provincia"
                      invalido={Boolean(errors.provinciaId)}
                      {...registroProvincia}
                      onChange={(e) => {
                        void registroProvincia.onChange(e)
                        alCambiarProvincia()
                      }}
                    />
                  </ListaCargable>
                </Campo>
                <Campo
                  id={idCanton}
                  etiqueta="Ciudad / Cantón"
                  requerido
                  error={errors.cantonId?.message}
                >
                  {provinciaId ? (
                    <ListaCargable
                      lista={cantones}
                      textoCarga="Cargando cantones..."
                      nombre="los cantones"
                    >
                      <Select
                        id={idCanton}
                        opciones={cantones.items.map((c) => ({ valor: c.id, etiqueta: c.nombre }))}
                        placeholder="Seleccione el cantón"
                        invalido={Boolean(errors.cantonId)}
                        {...registroCanton}
                      />
                    </ListaCargable>
                  ) : (
                    <Select
                      id={idCanton}
                      opciones={[]}
                      placeholder="Seleccione primero la provincia"
                      disabled
                    />
                  )}
                </Campo>
              </>
            )}

            {hayPais && !esEcuador && (
              <>
                <CampoTexto
                  name="provinciaTexto"
                  etiqueta="Provincia / Estado"
                  requerido
                  autoComplete="address-level1"
                />
                <CampoTexto
                  name="ciudadTexto"
                  etiqueta="Ciudad"
                  requerido
                  autoComplete="address-level2"
                />
              </>
            )}
          </div>

          <div className="grid gap-space-md sm:grid-cols-2">
            <CampoTexto
              name="telefonoCelular"
              etiqueta="Teléfono celular"
              requerido
              type="tel"
              inputMode="numeric"
              autoComplete="tel-national"
              placeholder="0999999999"
              ayuda="10 dígitos, empieza en 09."
            />
            <Campo
              id={idCorreo}
              etiqueta="Correo electrónico 1"
              ayuda="Es el correo de su registro."
            >
              <Input id={idCorreo} value={correoPrincipal} readOnly />
            </Campo>
            <CampoTexto
              name="emailSecundario"
              etiqueta="Correo electrónico 2"
              type="email"
              autoComplete="email"
              placeholder="otro@correo.com"
            />
            <CampoTexto
              name="telefonoDomicilio"
              etiqueta="Teléfono de domicilio"
              type="tel"
              inputMode="numeric"
              autoComplete="tel-national"
              ayuda="Solo dígitos, con código de área."
            />
            <CampoTexto name="callePrincipal" etiqueta="Calle principal" />
            <CampoTexto name="numero" etiqueta="Número" />
            <CampoTexto name="calleSecundaria" etiqueta="Calle secundaria" />
            <CampoTexto name="sector" etiqueta="Sector" />
          </div>
          <CampoTexto name="referencia" etiqueta="Referencia" />
        </fieldset>

        <div className="flex justify-end">
          <Button type="submit" cargando={guardando} textoCargando="Guardando...">
            Guardar
          </Button>
        </div>
      </form>
    </FormProvider>
  )
}

interface ListaCargableProps {
  lista: ListaCatalogo
  textoCarga: string
  nombre: string
  children: ReactNode
}

// Muestra el control cuando la lista está lista; mientras carga, un Loader, y si falla, el
// error con un botón para reintentar. El control se monta recién cuando llegan las
// opciones, para que muestre el valor que ya tenía el formulario.
function ListaCargable({ lista, textoCarga, nombre, children }: ListaCargableProps) {
  if (lista.cargando) return <Loader texto={textoCarga} />
  if (lista.error) {
    return (
      <div className="flex flex-col gap-space-sm">
        <Alert tipo="error">{`No se pudieron cargar ${nombre}. ${lista.error.message}`}</Alert>
        <Button variante="secondary" onClick={lista.reintentar}>
          Reintentar
        </Button>
      </div>
    )
  }
  return <>{children}</>
}

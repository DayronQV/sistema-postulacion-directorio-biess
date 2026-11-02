import { marcarPostulacionCreada } from '@/store/sesion'
import { http } from './http'

// Llamadas de la postulación (docs/api.md, sección Postulación).
//
// PUT /postulacion REEMPLAZA el estado completo del formulario: lo que no se envía se
// borra. Quien guarde una sección debe leer la postulación con obtenerPostulacion, cambiar
// solo su parte y enviar el formulario entero con guardarPostulacion.

export type Institucion = 'IESS' | 'BIESS'

export interface Familiar {
  nombres_apellidos: string
  parentesco: string
  institucion: Institucion
  area: string
  cargo: string
}

export interface ReferenciaPersonal {
  nombres_apellidos: string
  relacion: string
  telefono: string
  correo: string
}

export type NivelFormacion = 'tercer_nivel' | 'especialidad' | 'maestria' | 'phd'
export type TipoInstitucion = 'nacional' | 'extranjera'
export type AreaFormacion =
  'finanzas' | 'mercado_capitales' | 'administracion_empresas' | 'economia' | 'derecho' | 'afines'

export interface FormacionAcademica {
  nivel: NivelFormacion
  institucion: string
  tipo: TipoInstitucion
  titulo: string
  area: AreaFormacion
  senescyt_numero: string
  senescyt_fecha: string
  documento_id: string
}

export type TipoExperiencia = 'general' | 'direccion'
export type ActividadExperiencia =
  | 'finanzas'
  | 'mercado_capitales'
  | 'administracion_empresas'
  | 'economia'
  | 'derecho'
  | 'afines'

export interface Duracion {
  anios: number
  meses: number
  dias: number
}

export interface ExperienciaProfesional {
  tipo: TipoExperiencia
  cargo: string
  institucion: string
  fecha_inicio: string
  fecha_fin: string
  actividades: ActividadExperiencia
  descripcion: string
  documento_id: string
  duracion: Duracion
}

export type TipoPublicacion = 'publicacion' | 'investigacion'

export interface PublicacionAcademica {
  titulo: string
  tipo: TipoPublicacion
  medio: string
  fecha: string
  relacionada: boolean
  documento_id: string
}

export type TipoCapacitacion = 'curso' | 'seminario' | 'taller'

export interface Capacitacion {
  tipo: TipoCapacitacion
  nombre: string
  institucion: string
  fecha_inicio: string
  fecha_fin: string
  horas: number
  documento_id: string
}

export interface Domicilio {
  pais_id: number | null
  provincia_id: number | null
  canton_id: number | null
  provincia_texto: string
  ciudad_texto: string
  telefono_celular: string
  email_secundario: string
  calle_principal: string
  numero: string
  calle_secundaria: string
  sector: string
  referencia: string
  telefono_domicilio: string
}

export interface InformacionPersonal {
  cedula: string
  email: string // correo del registro (correo 1 del domicilio); solo lectura
  nombres: string
  apellidos: string
  genero: string
  fecha_nacimiento: string // AAAA-MM-DD
  nacionalidad: string
  estado_civil: string
  pareja?: { nombres_apellidos: string; cedula: string }
}

export interface RemovidoDetalle {
  institucion: string
  cargo: string
  fecha_fin: string
  motivo: string
}

// Claves "1" a "13" con la respuesta de cada pregunta, más el detalle de la 7.
export type Inhabilidades = Record<string, boolean | RemovidoDetalle | undefined> & {
  removido_detalle?: RemovidoDetalle
}

export type TipoCargo = 'afiliados' | 'jubilados' | 'empleador'

// Resultado de validar el cargo con el Historial Laboral. Si el postulante no cumple la
// calidad, el servidor responde 422 con el mensaje del requerimiento y cierra la postulación.
export interface ResultadoCargo {
  cargo: TipoCargo
  elegible: boolean
  mensaje: string
}

export interface Postulacion {
  id: string
  codigo: string
  estado: 'borrador' | 'enviada' | 'cerrada'
  informacion_personal: InformacionPersonal
  domicilio: Domicilio
  tiene_familiares: boolean
  familiares: Familiar[]
  referencias: ReferenciaPersonal[]
  formacion: FormacionAcademica[]
  experiencia: ExperienciaProfesional[]
  duracion_experiencia_total: Duracion
  publicaciones: PublicacionAcademica[]
  capacitaciones: Capacitacion[]
  cargo: string
  respaldo_organizacion: { nombre: string; documento_id: string | null }
  tiene_publicaciones: boolean
  consentimiento: boolean
  updated_at: string
  // Lo calcula el servidor a partir de las respuestas de Inhabilidades.
  inhabilidades?: Inhabilidades
  elegible?: boolean
}

// Lo que se edita y se guarda con PUT /postulacion: todo salvo lo que fija el servidor.
export type DatosPostulacion = Partial<
  Omit<Postulacion, 'id' | 'codigo' | 'estado' | 'informacion_personal' | 'updated_at' | 'elegible'>
>

export type EstadoPostulacion = Postulacion['estado']

// GET /postulacion: la postulación del usuario (el servidor la crea en borrador la primera vez).
// Con la postulación ya creada se levanta la restricción del primer ingreso (EF-11).
export async function obtenerPostulacion(signal?: AbortSignal): Promise<Postulacion> {
  const { data } = await http.get<Postulacion>('/postulacion', { signal })
  marcarPostulacionCreada()
  return data
}

// Avisos de guardado: cada sección guarda por su cuenta, y así el contenedor del formulario
// se entera (actualiza el avance y quita la marca de cambios sin guardar).
const alGuardar = new Set<(postulacion: Postulacion) => void>()

export function suscribirGuardado(escucha: (postulacion: Postulacion) => void): () => void {
  alGuardar.add(escucha)
  return () => {
    alGuardar.delete(escucha)
  }
}

// PUT /postulacion: guarda el formulario completo en una transacción y devuelve la
// postulación guardada. El servidor ignora los campos de solo lectura (id, código, estado...).
export async function guardarPostulacion(
  formulario: Postulacion,
  signal?: AbortSignal,
): Promise<Postulacion> {
  const { data } = await http.put<Postulacion>('/postulacion', formulario, { signal })
  alGuardar.forEach((f) => f(data))
  return data
}

// POST /postulacion/cargo/validar: consulta el Historial Laboral para el cargo elegido. No
// guarda el cargo (eso lo hace guardarPostulacion). 422 si no cumple la calidad: el servidor
// cierra la postulación y el mensaje del ApiError es el del requerimiento.
export async function validarCargo(cargo: TipoCargo): Promise<ResultadoCargo> {
  const { data } = await http.post<ResultadoCargo>('/postulacion/cargo/validar', { cargo })
  return data
}

// POST /postulacion/enviar: valida que esté completa, la marca "enviada" y manda el formulario
// al correo del postulante. 422 con `fields` por sección incompleta; 409 si ya fue enviada.
export async function enviarPostulacion(): Promise<Postulacion> {
  const { data } = await http.post<Postulacion>('/postulacion/enviar')
  return data
}

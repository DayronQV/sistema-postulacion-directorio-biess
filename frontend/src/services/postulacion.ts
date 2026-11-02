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
  cargo: string
  respaldo_organizacion: { nombre: string; documento_id: string | null }
  tiene_publicaciones: boolean
  consentimiento: boolean
  updated_at: string
}

// GET /postulacion: la postulación del usuario (el servidor la crea en borrador la primera vez).
// Con la postulación ya creada se levanta la restricción del primer ingreso (EF-11).
export async function obtenerPostulacion(): Promise<Postulacion> {
  const { data } = await http.get<Postulacion>('/postulacion')
  marcarPostulacionCreada()
  return data
}

// PUT /postulacion: guarda el formulario completo en una transacción y devuelve la
// postulación guardada. El servidor ignora los campos de solo lectura (id, código, estado...).
export async function guardarPostulacion(formulario: Postulacion): Promise<Postulacion> {
  const { data } = await http.put<Postulacion>('/postulacion', formulario)
  return data
}

// POST /postulacion/cargo/validar: consulta el Historial Laboral para el cargo elegido. No
// guarda el cargo (eso lo hace guardarPostulacion). 422 si no cumple la calidad: el servidor
// cierra la postulación y el mensaje del ApiError es el del requerimiento.
export async function validarCargo(cargo: TipoCargo): Promise<ResultadoCargo> {
  const { data } = await http.post<ResultadoCargo>('/postulacion/cargo/validar', { cargo })
  return data
}

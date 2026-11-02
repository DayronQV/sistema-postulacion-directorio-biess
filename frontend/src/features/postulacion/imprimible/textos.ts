// Texto de consentimiento y declaración que consta en el formulario final
// (docs/requerimiento-biess.md, 1.10.1).
export const CONSENTIMIENTO: string[] = [
  'Autorizo de manera libre y voluntaria, emitiendo consentimiento explícito al Instituto Ecuatoriano de Seguridad Social, para el uso de mis datos personales dentro del concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores, de conformidad con lo prescrito en el artículo 7 de la Ley Orgánica de Protección de Datos Personales.',
  'Declaro que la información del presente formulario y documentación adjunta corresponde a la verdad y autorizo al Instituto Ecuatoriano de Seguridad Social para que se compruebe ante cualquier organismo público o privado su veracidad.',
  'Acepto expresamente cumplir con las normas aplicables al concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores; así como las resoluciones y disposiciones impartidas por el Instituto Ecuatoriano de Seguridad Social. En caso de incumplimiento de requisitos o encontrarme inmerso/a en las prohibiciones o inhabilidades acepto mi descalificación del proceso.',
  'Acepto que mi nombre como postulante al concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores; sea publicado en la página web del IESS - BIESS.',
  'El presente formulario, debe ser impreso y con firma manuscrita, entregado junto con los requisitos solicitados, en el lugar, fecha y horario establecido por el Instituto Ecuatoriano de Seguridad Social.',
]

export const CARGOS: Record<string, string> = {
  afiliados:
    'Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS',
  jubilados:
    'Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS',
  empleador:
    'Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS',
}

// Etiquetas de los valores que guarda la API, como las muestran las secciones del formulario.
export const ETIQUETAS: Record<string, string> = {
  // Formación académica
  tercer_nivel: 'Tercer Nivel de Grado',
  especialidad: 'Cuarto Nivel - Especialidad Académica',
  maestria: 'Cuarto Nivel - Maestría Académica',
  phd: 'Cuarto Nivel PhD - Doctorado',
  nacional: 'Nacional',
  extranjera: 'Extranjera',
  // Áreas de estudio y actividades relacionadas
  finanzas: 'Finanzas',
  mercado_capitales: 'Mercado de Capitales',
  administracion_empresas: 'Administración de Empresas',
  economia: 'Economía',
  derecho: 'Derecho',
  afines: 'Materias afines',
  // Experiencia, publicaciones y capacitación
  general: 'General',
  direccion: 'Dirección',
  publicacion: 'Publicación',
  investigacion: 'Trabajo de investigación',
  curso: 'Curso',
  seminario: 'Seminario',
  taller: 'Taller',
}

// Etiqueta legible de un valor de la API; si no se conoce, el valor tal cual.
export function etiqueta(valor: string | null | undefined): string {
  if (!valor) return '—'
  return ETIQUETAS[valor] ?? valor
}

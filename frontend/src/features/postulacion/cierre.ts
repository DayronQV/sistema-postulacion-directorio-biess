// Sección del formulario a la que corresponde cada clave de `fields` que responde
// POST /postulacion/enviar con 422 (docs/api.md).
export const SECCION_DEL_ENVIO: Record<string, { id: string; titulo: string }> = {
  domicilio: { id: 'domicilio', titulo: 'Domicilio y contacto' },
  cargo: { id: 'cargo', titulo: 'Selección del cargo' },
  formacion: { id: 'formacion', titulo: 'Formación académica' },
  experiencia: { id: 'experiencia', titulo: 'Experiencia profesional' },
  capacitaciones: { id: 'capacitacion', titulo: 'Capacitación' },
  referencias: { id: 'referencias', titulo: 'Referencias personales' },
  inhabilidades: { id: 'inhabilidades', titulo: 'Inhabilidades' },
  consentimiento: { id: 'cierre', titulo: 'Cierre' },
}

export interface SeccionFaltante {
  id: string
  titulo: string
  motivo?: string
}

// Secciones que el servidor dijo que faltan, sin repetir una y en el orden en que llegaron.
export function faltantesDelServidor(
  fields: Record<string, string> | undefined,
): SeccionFaltante[] {
  const vistas = new Set<string>()
  const salida: SeccionFaltante[] = []
  for (const [campo, motivo] of Object.entries(fields ?? {})) {
    const seccion = SECCION_DEL_ENVIO[campo]
    if (!seccion || vistas.has(seccion.id)) continue
    vistas.add(seccion.id)
    salida.push({ ...seccion, motivo })
  }
  return salida
}

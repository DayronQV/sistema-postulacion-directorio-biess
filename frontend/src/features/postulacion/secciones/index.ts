import CapacitacionSeccion from './CapacitacionSeccion'
import CargoSeccion from './CargoSeccion'
import DomicilioSeccion from './DomicilioSeccion'
import ExperienciaSeccion from './ExperienciaSeccion'
import FamiliaresSeccion from './FamiliaresSeccion'
import FormacionSeccion from './FormacionSeccion'
import InformacionPersonalSeccion from './InformacionPersonalSeccion'
import PublicacionesSeccion from './PublicacionesSeccion'
import ReferenciasSeccion from './ReferenciasSeccion'
import SeccionPendiente from './SeccionPendiente'
import type { DefSeccion } from './tipos'

// Registro de secciones en el orden del requerimiento (docs/especificacion.md, EF-12).
// Para agregar o reemplazar una sección basta con cambiar su `Componente` aquí.
export const SECCIONES: DefSeccion[] = [
  {
    id: 'personal',
    titulo: 'Información personal',
    icono: 'badge',
    Componente: InformacionPersonalSeccion,
    estaCompleta: (p) => Boolean(p.informacion_personal),
  },
  {
    id: 'domicilio',
    titulo: 'Domicilio y contacto',
    icono: 'home',
    Componente: DomicilioSeccion,
    estaCompleta: (p) => Boolean(p.domicilio?.pais_id && p.domicilio.telefono_celular),
  },
  {
    id: 'adicional',
    titulo: 'Información adicional',
    icono: 'family_restroom',
    Componente: FamiliaresSeccion,
    estaCompleta: (p) => p.tiene_familiares === false || (p.familiares?.length ?? 0) > 0,
  },
  {
    id: 'cargo',
    titulo: 'Selección del cargo',
    icono: 'how_to_vote',
    Componente: CargoSeccion,
    estaCompleta: (p) => Boolean(p.cargo),
  },
  {
    id: 'formacion',
    titulo: 'Formación académica',
    icono: 'school',
    Componente: FormacionSeccion,
    estaCompleta: (p) => (p.formacion?.length ?? 0) > 0,
  },
  {
    id: 'experiencia',
    titulo: 'Experiencia profesional',
    icono: 'work',
    Componente: ExperienciaSeccion,
    estaCompleta: (p) => (p.experiencia?.length ?? 0) > 0,
  },
  {
    id: 'publicaciones',
    titulo: 'Publicaciones',
    icono: 'menu_book',
    Componente: PublicacionesSeccion,
    estaCompleta: (p) => p.tiene_publicaciones != null,
  },
  {
    id: 'capacitacion',
    titulo: 'Capacitación',
    icono: 'workspace_premium',
    Componente: CapacitacionSeccion,
    estaCompleta: (p) => (p.capacitaciones?.length ?? 0) > 0,
  },
  {
    id: 'referencias',
    titulo: 'Referencias personales',
    icono: 'contacts',
    Componente: ReferenciasSeccion,
    estaCompleta: (p) => (p.referencias?.length ?? 0) > 0,
  },
  {
    id: 'inhabilidades',
    titulo: 'Inhabilidades',
    icono: 'gavel',
    Componente: SeccionPendiente,
    estaCompleta: () => false,
  },
  {
    id: 'cierre',
    titulo: 'Cierre',
    icono: 'draw',
    Componente: SeccionPendiente,
    estaCompleta: (p) => p.consentimiento === true,
  },
]

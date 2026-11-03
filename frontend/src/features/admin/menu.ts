import type { Rol } from '@/store/sesion'

export interface ItemMenuAdmin {
  ruta: string
  titulo: string
  icono: string
  // Roles que ven el ítem; el servidor vuelve a validar en cada petición.
  roles: Rol[]
}

const TODOS: Rol[] = ['analista', 'administrador']
const ADMIN: Rol[] = ['administrador']

// Menú del panel de administración. Cada módulo de EF-13 (fases, dashboard, nómina, matriz,
// méritos) agrega aquí su pantalla y la registra en routes/index.tsx bajo /admin.
// Las pantallas de escritura van con roles: ['administrador'].
export const MENU_ADMIN: ItemMenuAdmin[] = [
  { ruta: '/admin', titulo: 'Inicio', icono: 'dashboard', roles: TODOS },
  // El analista las consulta; la pantalla deshabilita los controles si no es administrador.
  { ruta: '/admin/fases', titulo: 'Fases', icono: 'event', roles: TODOS },
  { ruta: '/admin/dashboard', titulo: 'Dashboard', icono: 'monitoring', roles: TODOS },
  { ruta: '/admin/postulaciones', titulo: 'Postulaciones', icono: 'folder_open', roles: TODOS },
  // La evaluación teórica (EF-14) es solo del administrador: el servidor responde 403 al analista.
  { ruta: '/admin/evaluacion/banco', titulo: 'Banco de preguntas', icono: 'quiz', roles: ADMIN },
  {
    ruta: '/admin/evaluacion/configuracion',
    titulo: 'Configurar examen',
    icono: 'tune',
    roles: ADMIN,
  },
  {
    ruta: '/admin/evaluacion/resultados',
    titulo: 'Resultados del examen',
    icono: 'grading',
    roles: ADMIN,
  },
]

export function menuParaRol(rol: Rol | undefined): ItemMenuAdmin[] {
  return rol ? MENU_ADMIN.filter((i) => i.roles.includes(rol)) : []
}

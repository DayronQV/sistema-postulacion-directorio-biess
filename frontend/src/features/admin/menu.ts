import type { Rol } from '@/store/sesion'

export interface ItemMenuAdmin {
  ruta: string
  titulo: string
  icono: string
  // Roles que ven el ítem; el servidor vuelve a validar en cada petición.
  roles: Rol[]
}

const TODOS: Rol[] = ['analista', 'administrador']

// Menú del panel de administración. Cada módulo de EF-13 (fases, dashboard, nómina, matriz,
// méritos) agrega aquí su pantalla y la registra en routes/index.tsx bajo /admin.
// Las pantallas de escritura van con roles: ['administrador'].
export const MENU_ADMIN: ItemMenuAdmin[] = [
  { ruta: '/admin', titulo: 'Inicio', icono: 'dashboard', roles: TODOS },
  { ruta: '/admin/dashboard', titulo: 'Dashboard', icono: 'monitoring', roles: TODOS },
]

export function menuParaRol(rol: Rol | undefined): ItemMenuAdmin[] {
  return rol ? MENU_ADMIN.filter((i) => i.roles.includes(rol)) : []
}

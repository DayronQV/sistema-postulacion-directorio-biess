// Único módulo que lee variables de entorno de Vite.
export const env = {
  apiUrl: import.meta.env.VITE_API_URL ?? 'http://localhost:8080/api/v1',
}

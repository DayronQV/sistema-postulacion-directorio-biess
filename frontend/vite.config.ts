import { fileURLToPath } from 'node:url'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  // Un solo .env en la raíz del repositorio (ver .env.example); Vite solo expone las VITE_*.
  envDir: '..',
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    css: false,
    // El límite por defecto (5 s) se queda corto cuando la suite corre en paralelo y una
    // prueba llena y guarda un formulario: fallaba una distinta en cada corrida aunque el
    // código estuviera bien. Con 15 s pasan y un bloqueo real sigue fallando.
    testTimeout: 15000,
  },
})

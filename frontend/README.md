# Frontend

SPA con React, TypeScript, Vite y Tailwind CSS. Estructura y reglas en `docs/arquitectura.md`.

## Comandos

```bash
npm install
npm run dev      # servidor de desarrollo en http://localhost:5173
npm run lint     # ESLint
npm run format   # Prettier
npm test         # Vitest + Testing Library
npm run build    # comprobación de tipos y build de producción
```

## Configuración

- `VITE_API_URL` se toma del `.env` de la raíz del repositorio (copia de `.env.example`; `envDir: '..'` en `vite.config.ts`) y se lee solo desde `src/config/env.ts`.
- El alias `@/` apunta a `src/`.
- Los colores, la tipografía y el espaciado del diseño institucional están en `tailwind.config.ts`.

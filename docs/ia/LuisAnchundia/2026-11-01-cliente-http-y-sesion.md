# 2026-11-01: cliente HTTP, sesión y rutas protegidas

- **Autor (usuario de GitHub):** LuisAnchundia
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #9 (cliente Axios, estado de sesión y rutas protegidas)

## Prompt
```
Sí, sigue con #9
(Contexto: cliente Axios con token Bearer, errores normalizados al formato de docs/api.md,
sesión en el cliente, ruta protegida por sesión y rol, y un hook para llamadas con loading/error.)
```

## Resultado
- Archivos generados o modificados:
  - `frontend/src/services/http.ts` y `errores.ts` (instancia de Axios, interceptores y `ApiError`).
  - `frontend/src/store/sesion.ts` (token en `sessionStorage`, vencimiento y rol leído del JWT).
  - `frontend/src/routes/RutaProtegida.tsx` y `routes/index.tsx` (guardia de las rutas privadas).
  - `frontend/src/hooks/useApi.ts` (estados cargando/éxito/error y cancelación con `AbortController`).
  - `frontend/src/components/layout/LayoutPrivado.tsx` (Cerrar sesión usa la sesión real).
  - Pruebas de cada pieza y `axios` como dependencia nueva.
- Resumen de lo que hizo: el cliente agrega el token a cada petición y convierte cualquier fallo en un `ApiError` con `code`, `message`, `fields` y `retryAfterSeconds`. Un 401 con sesión abierta la cierra y la guardia redirige al login. `/postulacion` ahora exige sesión.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npm run lint`, `npm test` (40 pruebas) y `npm run build`.
- Qué se corrigió: la firma de `useApi` pasó a recibir la señal como primer argumento porque TypeScript no inferia bien los argumentos con la señal al final.
- Pendiente: revisión del código por la persona autora. El rol se lee del payload del JWT (`rol`); falta confirmar con el backend (#20) el nombre exacto del claim.

# 2026-11-01: inicio de sesión con feedback de bloqueo

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #17

## Prompt
```
podemos continuar con la pantalla de incio de sesion, el issue #17
```

## Resultado
- Archivos generados o modificados:
  - `frontend/src/services/auth.ts` y `auth.test.ts`: `iniciarSesionPostulante` (`POST /auth/login`).
  - `frontend/src/features/auth/FormularioLogin.tsx`: formulario de cédula y contraseña con el aviso de bloqueo.
  - `frontend/src/features/auth/esquemas.ts` y `esquemas.test.ts`: `esquemaLogin`.
  - `frontend/src/features/auth/useCuentaRegresiva.ts` y `PasoIdentidad.tsx`: `BLOQUEO_POR_DEFECTO_S` pasa al hook para compartirlo entre registro y login.
  - `frontend/src/pages/Login.tsx` y `Login.test.tsx`.
- Resumen de lo que hizo: formulario de cédula y contraseña validado con zod. Con credenciales incorrectas muestra "sus credenciales de ingreso son incorrectas" y limpia la contraseña; ante el 429 muestra el bloqueo con cuenta regresiva desde `retry_after_seconds` y el formulario deshabilitado, y al terminar deja reintentar. Con un login correcto guarda la sesión (`iniciarSesion(token, expires_in)`) y redirige: si `primer_ingreso` es verdadero, obligatoriamente al formulario de postulación; si no, a la ruta privada que pidió el login. Incluye el enlace "¿Olvidaste tu contraseña?" y el aviso de registro exitoso al llegar desde el registro.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npm run lint`, `npm test` (13 archivos, 98 pruebas, dos corridas seguidas) y `npm run build` en verde. La prueba principal hace tres intentos fallidos, comprueba el bloqueo con la cuenta regresiva y el formulario deshabilitado, y tras la cuenta regresiva vuelve a entrar.
- Qué se corrigió: dos pruebas fallaban porque `beforeEach(() => mockLogin.mockReset())` devolvía el mock y Vitest lo ejecutaba como limpieza al final de cada prueba, dejando una promesa rechazada sin manejar; se escribió el callback con llaves.

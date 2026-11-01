# 2026-11-01: pantalla de inicio y registro en dos pasos

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #16

## Prompt
```
Sí, empezar la #16
```
(Ante la pregunta de cómo validar, ya que zod no estaba instalado, se eligió instalar zod.)

## Resultado
- Archivos generados o modificados:
  - `frontend/package.json` y `package-lock.json`: `zod@4.6.7` y `@hookform/resolvers@5.9.2`, que piden la issue y `docs/arquitectura.md`.
  - `frontend/src/services/auth.ts` y `auth.test.ts`: `validarIdentidad` y `registrarCuenta`.
  - `frontend/src/features/auth/`: `cedula.ts` (módulo 10), `esquemas.ts` (zod), `fortaleza.ts`, `useCuentaRegresiva.ts`, `errores.ts`, `SelectorDato.tsx`, `PasoIdentidad.tsx`, `PasoCuenta.tsx`, `IndicadorFortaleza.tsx` y sus pruebas.
  - `frontend/src/pages/Inicio.tsx`, `Registro.tsx` y `Registro.test.tsx`.
- Resumen de lo que hizo: la página de inicio muestra "Iniciar sesión" y "Registrarse". El registro tiene dos pasos: (1) cédula y fecha de expedición o código dactilar, validados con zod y luego con `POST /auth/identidad`; muestra "la información ingresada es incorrecta" y, ante el 429, "registro fallido, inténtelo nuevamente en 10 minutos" con cuenta regresiva desde `retry_after_seconds` y el formulario deshabilitado; (2) correo, contraseña y confirmación con indicador de fortaleza contra `POST /auth/registro`. Los errores por campo del servidor se muestran junto a cada campo, un ticket vencido vuelve al paso 1 y el registro correcto muestra un mensaje de éxito y lleva al login.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npm run lint`, `npm test` (12 archivos, 88 pruebas) y `npm run build` en verde. Las pruebas recorren el flujo completo con el servicio simulado, el bloqueo con cuenta regresiva (deshabilita y luego habilita el formulario), la cédula ya registrada, el correo repetido, el ticket vencido y la redirección automática. Se revisó la pantalla en el navegador con el servidor de desarrollo.
- Qué se corrigió:

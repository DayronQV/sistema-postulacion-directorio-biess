# 2026-11-01: recuperar y restablecer contraseña (frontend)

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #18

## Prompt
```
el pr 75 ya fue mergeado, si es asi asi podemos continuar con la issue #18
```

## Resultado
- Archivos generados o modificados:
  - `frontend/src/services/auth.ts` y `auth.test.ts`: `solicitarRecuperacion` (`POST /auth/recuperar`) y `restablecerPassword` (`POST /auth/restablecer`).
  - `frontend/src/features/auth/FormularioIdentidad.tsx`: el formulario de identidad del registro pasa a ser un componente reutilizable (recibe la llamada, el texto del botón y el mensaje de bloqueo); `PasoIdentidad.tsx` queda como envoltorio del registro.
  - `frontend/src/features/auth/esquemas.ts` y `esquemas.test.ts`: `esquemaRestablecer`, que comparte la regla de contraseña con el registro.
  - `frontend/src/pages/Recuperar.tsx`, `Restablecer.tsx` y sus pruebas; ruta `/restablecer` en `frontend/src/routes/index.tsx` y su prueba.
- Resumen de lo que hizo: `/recuperar` pide la cédula y la fecha de expedición o el código dactilar, con el mismo bloqueo y cuenta regresiva que el registro, y al responder `202` muestra un mensaje neutro (no revela si la cédula tiene cuenta). `/restablecer?token=` pide la nueva contraseña y su confirmación con indicador de fortaleza; con `204` muestra el éxito y lleva al login; con `401` (enlace vencido o usado) muestra un aviso claro con el enlace para pedir otro; sin token en la URL no muestra el formulario.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npm run lint`, `npm test` (15 archivos, 114 pruebas, dos corridas) y `npm run build` en verde. Las pruebas del registro siguen pasando sin cambios tras extraer `FormularioIdentidad`.
- Qué se corrigió: Prettier sobre la carpeta reformateaba líneas de la validación de fecha de Gualter; se restauraron sus archivos y solo se formatearon los propios.

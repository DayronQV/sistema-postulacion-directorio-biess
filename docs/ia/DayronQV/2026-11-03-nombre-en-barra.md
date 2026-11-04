# 2026-11-03: nombre del postulante en la barra y estado "En proceso"

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** sin issue (ajustes de presentación detectados al probar en producción)

## Prompt
```
Actúa como desarrollador frontend sénior (React, TypeScript y Tailwind) del sistema del
concurso del Directorio del BIESS.

Probando en producción vi dos cosas:
1. La barra superior del postulante dice "Postulante". Debe mostrar el nombre de quien
   ingresó (por ejemplo "Dayron Quiñonez") y su cédula.
2. La etiqueta del estado dice "Borrador". Cámbiala por "En proceso", que es más claro para
   el postulante.

Requisitos:
- Sin tocar la API ni el backend: el nombre ya viene del Registro Civil en la postulación.
- No repitas la carga de la postulación que ya hace el formulario, y que al cerrar sesión
  no quede el nombre de la persona anterior para quien entre después en la misma pestaña.
- Componentes pequeños; las funciones auxiliares fuera del componente.
- Pruebas para lo nuevo; después rompe el código a propósito, una pieza a la vez, y
  comprueba que las pruebas lo detectan.
- Corre tipos, lint, todas las pruebas de Vitest y el build, y compruébalo en el navegador.
- Commits en mi nombre, en inglés y con Conventional Commits, y PR con la plantilla.
```

## Resultado
- Archivos generados o modificados: `features/postulacion/LayoutPostulante.tsx` (nuevo, con sus pruebas), `features/postulacion/store.ts` (`cargarPostulacionSiFalta`, con su prueba), `features/postulacion/informacionPersonal.ts` (`nombreCorto`), `routes/index.tsx`, `pages/FormularioPostulacion.tsx` y su prueba, y un comentario de `EncabezadoFormulario.tsx`.
- Resumen de lo que hizo: un marco del postulante que toma el primer nombre y el primer apellido de la postulación y los pasa a la barra con la cédula; carga la postulación solo si la pantalla no la pidió, oculta el nombre mientras se vuelve a cargar y la limpia al cerrar sesión. La etiqueta del formulario pasa de "Borrador" a "En proceso" (el panel de administración sigue diciendo "Borrador").

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npx tsc -b`, `npm run lint`, `npx vitest run` (645 pruebas) y `npm run build`; 9 mutaciones a mano sobre el marco, el nombre corto y la carga condicional, todas detectadas; en el navegador local, con una cuenta de prueba, la barra mostró "María Zambrano" y su cédula, en celular sin desborde, y al cerrar sesión volvió al login sin sesión.
- Qué se corrigió: la función `nombreCorto` se movió del componente a `informacionPersonal.ts` porque el lint avisó que rompía la recarga en caliente; una mutación sobrevivió al principio (mostrar el nombre durante una nueva carga) y se agregó su prueba.

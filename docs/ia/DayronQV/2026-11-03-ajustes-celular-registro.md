# 2026-11-03: ajustes de registro y de celular detectados en producción

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** sin issue (ajustes de presentación detectados al probar en producción)

## Prompt
```
Actúa como desarrollador frontend sénior (React, TypeScript y Tailwind) del sistema del
concurso del Directorio del BIESS. Probando en producción, sobre todo en el celular, vi:

1. En el registro, al escribir la contraseña la página salta: el panel de la izquierda y
   el formulario se mueven, como si cambiara el tamaño de la pantalla.
2. Las contraseñas del registro no tienen el ojo para mostrarlas (el login sí).
3. En las secciones con tabla (formación, experiencia, etc.), al tocar "Agregar" en el
   celular el formulario aparece fuera de la vista y parece que no pasó nada.
4. En el pie de las pantallas de acceso, el escudo azul se ve cortado en el celular.

Corrige las cuatro cosas en todas las pantallas que tengan el mismo problema, con
componentes pequeños y reutilizables. Sin tocar la API.

Requisitos: pruebas para cada cambio; rompe el código a propósito, una pieza a la vez, y
comprueba que las pruebas lo detectan; corre tipos, lint, todas las pruebas y el build, y
compruébalo en el navegador en tamaño de celular. Commits en mi nombre, en inglés y con
Conventional Commits, y PR con la plantilla hacia develop.
```

## Resultado
- Archivos generados o modificados: `features/auth/CampoClave.tsx` y `components/ui/PanelEditor.tsx` (nuevos, con sus pruebas), `features/auth/IndicadorFortaleza.tsx`, `PasoCuenta.tsx`, `FormularioLogin.tsx`, `pages/Restablecer.tsx`, las secciones Formación, Experiencia, Publicaciones, Capacitación y Referencias (y sus pruebas), `components/layout/LayoutAcceso.tsx` y su prueba, y `pages/Registro.test.tsx`.
- Resumen de lo que hizo:
  1. El indicador de seguridad de la contraseña reserva su lugar invisible desde el inicio: antes aparecía con la primera letra, la tarjeta crecía, salía la barra de desplazamiento y todo se corría.
  2. `CampoClave` (campo de contraseña con su ojo) se usa en el registro, en restablecer contraseña y en el login, que dejó de tener su botón propio.
  3. `PanelEditor` envuelve el formulario de fila de 5 secciones: al abrirse baja la pantalla hasta él (sin animación si el sistema pide reducir el movimiento) y pone el cursor en el primer campo. Familiares no lo necesitaba porque usa una ventana.
  4. El escudo y la línea del pie ya no se encogen (`shrink-0`) cuando el texto ocupa todo el ancho.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npx tsc -b`, `npm run lint`, `npx vitest run` (654 pruebas) y `npm run build`; 13 mutaciones a mano sobre los cuatro cambios, todas detectadas; en el navegador local: el alto de la tarjeta y de la página y la posición del formulario iguales antes y después de escribir la contraseña, el ojo mostrando la contraseña, "Agregar experiencia" en 375 px bajando al formulario con el cursor en "Tipo", y el escudo completo (22 × 22 px) en 360 px.
- Qué se corrigió: una mutación sobrevivió al principio (pasar de editar una fila a editar otra sin volver a bajar) y se agregó su prueba; dos mutaciones no se aplicaban por los finales de línea de Windows y se repitieron bien. El servidor de desarrollo de Vite se quedó dos veces con módulos viejos tras cambiar de rama y hubo que reiniciarlo; no era un error del código.

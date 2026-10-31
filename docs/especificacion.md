# Especificación condensada (requerimiento completo)

Resumen fiel del requerimiento funcional del IESS v1.4. **Se implementa todo el documento.** Las prioridades indican el orden de trabajo, no qué se omite:

- **P1 (lo que califica la rúbrica del examen):** EF-11 y EF-12 completos.
- **P2:** módulo de administración (EF-14).
- **P3:** banco de preguntas y evaluación teórica.

## Roles
- **Postulante** (usuario externo).
- **Analista** (Subdirección Nacional de Gestión de Talento Humano).
- **Administrador** (Comisión de Selección).

Analista y administrador son usuarios sembrados en la base de datos, sin registro público.

## Convenciones del requerimiento
- `*` = respuesta obligatoria.
- Todo ítem del formulario tiene botones **Editar** y **Eliminar**; cada sección tiene **Guardar** y se puede llenar por partes sin perder datos ("Guardar Progreso").
- El formulario tiene un **código de numeración único**.
- Límite de documentos: **PDF**, máximo **50 MB por postulante** en total.
- Integraciones **simuladas** con adaptador `mock`: Registro Civil, Historial Laboral del IESS y correo electrónico (se escribe en log).

---

## P1. EF-11: Registro y autenticación

### Registro
1. Cédula + a elegir **fecha de expedición** (DD/MM/AAAA) o **código dactilar** (10 dígitos).
2. Validación contra el Registro Civil (simulado). Si coincide, se piden **correo** y **contraseña**.
3. Si no coincide: "la información ingresada es incorrecta".
4. **3 intentos fallidos** → "registro fallido, inténtelo nuevamente en 10 minutos"; pasados los 10 minutos se puede reintentar.
5. Registro correcto → pantalla de inicio de sesión (credenciales: cédula y contraseña).

### Inicio de sesión
- Contraseña incorrecta: "sus credenciales de ingreso son incorrectas".
- **3 intentos fallidos consecutivos → bloqueo de 10 minutos** (Redis), con feedback de bloqueo en pantalla.
- "¿Olvidaste tu contraseña?": cédula + fecha de expedición o código dactilar; si valida, se envía un correo de restablecimiento (simulado).
- Primer ingreso: se dirige obligatoriamente al formulario de postulación.
- Rutas del formulario protegidas por sesión.

---

## P1. EF-12: Formulario de postulación

Secciones, en orden: Información personal, Información domiciliaria y de contacto, Información adicional, Selección del cargo, Formación académica, Experiencia profesional, Publicaciones, Capacitación, Inhabilidades, y cierre (consentimiento, guardar y enviar).

### 1. Información personal (precargada del Registro Civil, solo lectura)
Nombres, apellidos, cédula, género, fecha de nacimiento, nacionalidad, estado civil.
- **Nacionalidad no ecuatoriana → no se permite continuar y se cierra la postulación.**
- Estado civil casado/a o unión de hecho → se piden nombres y apellidos y cédula de la pareja.

### 2. Información domiciliaria y de contacto
- País (lista de ~190 países, incluido Ecuador).
- Provincia/Estado: lista de las 24 provincias si el país es Ecuador; texto libre si es otro país.
- Ciudad: lista de cantones de la provincia elegida si el país es Ecuador; texto libre si es otro.
- Teléfono celular, correo electrónico 1 (precargado del registro), correo 2, calle principal, número, calle secundaria, sector, referencia, teléfono de domicilio.

### 3. Información adicional
"¿Tiene familiares hasta el 2.º grado de consanguinidad o 4.º de afinidad trabajando en IESS o BIESS?" Sí/No. Si es Sí, filas con: nombres y apellidos, parentesco (conviviente, cónyuge, cuñado/a, hermano/a, hijo político, hijo/a, madre, nieto/a, nuera, padre, primo hermano/a, sobrino/a, suegro/a, tío/a, yerno), institución (IESS, BIESS), área y cargo. Botones Agregar, Editar, Eliminar y Guardar.

### 4. Selección del cargo (elegir uno)
- Miembro Principal y Alterno, sector **Afiliados**.
- Miembro Principal y Alterno, sector **Jubilados**.
- Miembro Principal y Alterno, sector **Empleador**.

Elegibilidad validada con Historial Laboral (simulado):
| Cargo | Resultado válido | Resultado inválido (mensaje y cierre del sistema) |
| --- | --- | --- |
| Afiliados | "Postulante afiliado al IESS" | "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS, NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS" |
| Jubilados | "Postulante acredita jubilación" | "…Sector de los Jubilados…, NO puede continuar, por cuanto no acredita jubilación" |
| Empleador | "Postulante es empleador" | "…Sector Empleador…, NO puede continuar, por cuanto no es empleador" |

Si es válido, se pregunta si lo respalda una organización de representación nacional legalmente constituida: **organización de trabajadores** (afiliados), **organización de jubilados** (jubilados) o **federación de cámara** (empleador). Si responde Sí: nombre de la organización `*` y documento de respaldo `*` (PDF, **máx. 4 MB**). Botón "Guardar Postulación".

### 5. Formación académica
Botón Agregar. Nivel de instrucción: Tercer Nivel de Grado, Cuarto Nivel - Especialidad Académica, Cuarto Nivel - Maestría Académica, Cuarto Nivel PhD - Doctorado. Luego: institución educativa, tipo (Nacional/Extranjera), título obtenido, área de estudios (Finanzas, Mercado de Capitales, Administración de Empresas, Economía, Derecho o materias afines), número y fecha de registro SENESCYT, documento de respaldo (PDF, **máx. 2 MB**).

### 6. Experiencia profesional
Botón Agregar. Tipo (General, Dirección), cargo o puesto, institución o empresa, fecha inicio, fecha finalización, actividades relacionadas a (finanzas, mercado de capitales, administración de empresas, economía, derecho o materias afines), descripción de funciones (**máx. 1000 caracteres**), documento de respaldo `*` por experiencia (PDF, **máx. 2 MB**).
- **Calculadora de experiencia:** a partir de las fechas, se muestra el tiempo de cada experiencia en **años, meses y días**.

### 7. Publicaciones
"¿Cuenta con Publicaciones e Investigaciones académicas?" Sí/No. Si es Sí: título, tipo (Publicación / Trabajo de investigación), medio de publicación, fecha, "¿guarda relación con su ejercicio o formación profesional?" (Sí/No), documento de respaldo (PDF, **máx. 2 MB**). Agregar, Editar, Eliminar, Guardar.

### 8. Capacitación
Botón Agregar. Tipo de evento (Curso, Seminario, Taller), nombre, institución, fecha inicio, fecha fin, número de horas, documento de respaldo (PDF, **máx. 2 MB**).

### 9. Inhabilidades (preguntas Sí/No)
1. ¿Es ecuatoriano/a y está en pleno goce de los derechos de participación política?
2. ¿Se encuentra inhabilitado para ejercer el comercio?
3. ¿Está en mora, directa o indirectamente, con instituciones del Estado o bajo control de la Superintendencia de Bancos, incluidas las off-shore?
4. ¿Mantiene vínculo, por propiedad o administración, con instituciones del sistema financiero y de seguros privados, bolsas de valores, casas de valores y administradoras de fondos y fideicomisos?
5. ¿Es funcionario o empleado del IESS - BIESS?
6. ¿Mantiene interés propio o representa a terceros en la propiedad, dirección o gestión de aseguradoras u otras personas del sistema nacional de seguridad social?
7. ¿En los últimos 5 años ha sido removido por algún organismo de control? Si es Sí: institución `*`, cargo, fecha de finalización y motivo de finalización `*`.
8. ¿Mantiene obligaciones pendientes con el SRI?
9. ¿Ha incurrido en los últimos 5 años en castigo de sus obligaciones por parte de una institución financiera?
10. ¿Actualmente litiga contra el IESS - BIESS?
11. ¿Está procesado, enjuiciado o condenado por delito de corrupción o crimen organizado?
12. ¿Mantiene en firme responsabilidades de la Contraloría General del Estado?
13. ¿Está registrado en la base de datos de la Unidad de Análisis Financiero y Económico?

### 10. Cierre
- **Guardar Postulación:** muestra el texto de consentimiento y declaración (5 numerales, transcritos en `docs/requerimiento-biess.md`), que consta en el formulario final. Guardar persiste todo en **una transacción**.
- El formulario incluye al final un apartado con la **firma** del postulante, sus nombres y apellidos y su número de cédula.
- **Enviar Postulación:** el formulario completo se envía al correo del postulante (simulado).

---

## P2. EF-14: Módulo de administración
Para Analista y Administrador.
1. **Publicar cada fase** del concurso.
2. **Habilitar** el apartado EF-12 "Formulario de Postulación".
3. **Dashboard** en tiempo real con el número de postulantes por cargo.
4. **Detalle de postulantes históricos:** al abrir una postulación cerrada, se muestra la nómina completa y el detalle íntegro del formulario de cada postulante, con **descarga de certificados** y **calculadora de experiencia**.
5. **Generación de matriz:** exportar a **Excel (.xlsx)** y **PDF** el cuadro comparativo de postulantes.

## P3. Banco de preguntas y evaluación teórica
- El campo está **bloqueado y oculto** al público; solo accede quien superó la **Fase de Méritos** (su estado pasa a "Evaluación Técnica"), en la fecha y hora que el administrador habilite.
- El administrador **carga el banco de preguntas** de opción múltiple y la matriz de respuestas correctas.
- Al iniciar sesión el postulante habilitado en el horario programado, el motor **sortea un número parametrizado** de preguntas del banco: examen único y en distinto orden por postulante.
- "Iniciar Evaluación": preguntas **una por una**, con **reloj en cuenta regresiva** arriba (tiempo configurable).
- **Seguridad:** se bloquea el botón "Atrás" del navegador y las funciones copiar y pegar.
- "Finalizar y Calificar" envía el examen. Si el reloj llega a 00:00, cierre automático y envío de lo respondido hasta ese segundo.
- Calificación automática contra la matriz de respuestas. La nota se guarda en el expediente digital, se genera un **reporte PDF**, se **notifica por correo** (simulado) y se almacena en el módulo de administración.

## Contexto general (EF-1)
El sistema debe además: separar postulantes por sector, verificar requisitos e inhabilidades, calificar méritos y oposición, gestionar reconsideraciones con plazos legales, notificar por correo cada etapa y generar reportes y listas de elegibles con trazabilidad auditable. El documento solo detalla las pantallas anteriores; estos puntos se refinan en issues propias cuando el grupo los priorice.

## Restricciones
Cumplimiento de la normativa de protección de datos personales de Ecuador, periodos definidos de publicación y cierre de postulación, y límites de tamaño y formato de documentos.

## Preguntas abiertas (decidir como grupo o consultar al profesor)
1. **"Formación adicional" y "Referencias personales"** aparecen en la lista de campos del formulario pero el documento no define sus campos. Propuesta: "formación adicional" se cubre con Capacitación; "referencias personales" (nombre, parentesco o relación, teléfono, correo) queda por definir.
2. El bloqueo de 3 intentos: el requerimiento lo pide en el registro y el examen en el login. Propuesta: aplicarlo a ambos, con claves separadas en Redis.
3. Validación de identidad simulada: propuesta aceptar fecha de expedición **o** código dactilar.
4. ¿Cómo se determina que un postulante "superó la Fase de Méritos"? Propuesta: el administrador lo marca manualmente.

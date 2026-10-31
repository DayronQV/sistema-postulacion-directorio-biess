# Especificación condensada (EF-11 y EF-12)

Resumen fiel del requerimiento funcional del IESS, limitado a lo que evalúa el examen. Fuente: documento de requerimientos v1.4 y enunciado del examen.

## Actores
- **Postulante** (usuario externo). Es el único rol que se implementa.

## EF-11: Registro y autenticación

### Registro
1. El postulante ingresa su **cédula** y, a elegir, su **fecha de expedición** (DD/MM/AAAA) o su **código dactilar** (10 dígitos).
2. El sistema valida contra el Registro Civil (**simulado**: adaptador `mock` con datos semilla).
3. Si coincide, pide **correo electrónico** y **contraseña** y crea la cuenta.
4. Si no coincide: mensaje "la información ingresada es incorrecta".
5. **3 intentos fallidos** → mensaje "registro fallido, inténtelo nuevamente en 10 minutos". Pasados 10 minutos se permite reintentar.

### Inicio de sesión
- Credenciales: **cédula + contraseña**.
- Contraseña incorrecta: "sus credenciales de ingreso son incorrectas".
- **3 intentos fallidos consecutivos → bloqueo de 10 minutos** (Redis). La pantalla muestra el feedback de bloqueo.
- Enlace "¿Olvidaste tu contraseña?": pide cédula + fecha de expedición o código dactilar; si valida, se envía un correo de restablecimiento (**simulado**: se registra en log).
- Las rutas del formulario son **protegidas** (requieren sesión).

## EF-12: Formulario de postulación
Al entrar por primera vez, el postulante es dirigido obligatoriamente al formulario. Cada sección tiene "Guardar" y se puede llenar por partes sin perder datos. Todo ítem se puede editar y eliminar. `*` = obligatorio.

### Núcleo obligatorio (cubre la rúbrica)
1. **Información personal** (precargada desde el Registro Civil simulado, solo lectura): nombres, apellidos, cédula, género, fecha de nacimiento, nacionalidad, estado civil.
   - **Nacionalidad distinta de ecuatoriana → no se permite continuar y la postulación se cierra.**
   - Si el estado civil es casado/a o unión de hecho: se piden nombres y apellidos y cédula de la pareja.
2. **Información domiciliaria y de contacto:**
   - País (lista), Provincia/Estado y Ciudad.
   - **Cascada geográfica:** si el país es Ecuador, provincia y ciudad son listas dependientes (País → Provincia → Cantón); si es otro país, se escriben a mano.
   - Teléfono celular, correo 1 (precargado), correo 2, calle principal, número, calle secundaria, sector, referencia, teléfono de domicilio.
3. **Información adicional:** "¿Tiene familiares hasta el 2.º grado de consanguinidad o 4.º de afinidad trabajando en IESS/BIESS?" Sí/No. Si es Sí, se agregan filas: nombres y apellidos, parentesco (lista), institución (IESS/BIESS), área, cargo.
4. **Selección del cargo:** Afiliados, Jubilados o Empleador. La elegibilidad la valida Historial Laboral (**simulado**). Si no cumple, se muestra el mensaje de cierre del requerimiento y se cierra el proceso.
5. **Guardar Postulación** persiste todo en **una sola transacción** (varias tablas). **Enviar Postulación** marca la postulación como enviada y simula el correo.

### Secciones del requerimiento NO incluidas por defecto
Formación académica, experiencia profesional (con cálculo en años/meses/días), publicaciones, capacitación e inhabilidades. Se agregan **solo si el grupo lo decide** (ver "Preguntas abiertas") y se detallan en una issue propia.

## Reglas transversales
- Documentos de respaldo: PDF, máximo 2 MB (4 MB para organizaciones) y 50 MB por postulante. Solo aplica si se implementan secciones con archivos.
- El formulario tiene un **código de numeración único**.
- Texto de consentimiento y declaración al guardar: se toma del requerimiento (5 numerales) si se implementa el cierre del formulario.
- Todo error de validación se muestra por campo, de forma inmediata.

## Preguntas abiertas (decidir como grupo o consultar al profesor)
1. ¿Qué secciones de EF-12 más allá del núcleo se implementan? El examen evalúa cascada, nacionalidad y transacción multi-tabla; el núcleo lo cubre.
2. ¿El bloqueo de 3 intentos aplica al registro, al login o a ambos? El requerimiento lo pide en el registro y el examen en el login. Propuesta: ambos, con claves separadas en Redis.
3. ¿Cuál identificador se usa en la validación simulada: fecha de expedición, código dactilar o ambos? Propuesta: aceptar cualquiera de los dos.

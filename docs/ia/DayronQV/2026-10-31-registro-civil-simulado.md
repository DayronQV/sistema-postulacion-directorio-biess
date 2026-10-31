# 2026-10-31: simulador del Registro Civil

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #5

## Prompt (versión redactada)
```
Actúa como desarrollador backend en Go.

Necesito simular el Registro Civil del Ecuador, porque no hay acceso real, para
validar la identidad en el registro y precargar los datos personales del formulario.

Crea:
1. Una interfaz RegistroCivil que consulte a una persona por su cédula.
2. La entidad Ciudadano: datos personales, nacionalidad, estado civil, pareja (si es
   casado/a o unión de hecho), fecha de expedición y código dactilar.
3. Una función que valide la cédula ecuatoriana: 10 dígitos, provincia del 01 al 24,
   tercer dígito menor que 6 y dígito verificador por módulo 10.
4. Un adaptador mock con 7 ciudadanos de prueba: flujo feliz, casada, unión de hecho,
   extranjero, sin calidad de afiliado, bloqueo por intentos y recuperación de
   contraseña. Una cédula que no exista debe devolver "no encontrado".

Agrega pruebas y documenta los datos de prueba en el README.

La issue #5 del repositorio tiene el detalle completo (tabla de ciudadanos de prueba,
archivos a tocar y criterios de terminado): revísala para cumplirla exactamente.
```

## Para qué se usó
Construir el simulador del Registro Civil (interfaz, validación de cédula y 7 ciudadanos de prueba) con sus pruebas.

# 2026-11-01: autenticación con JWT y roles

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #4

## Prompt (versión redactada)
```
Actúa como desarrollador backend en Go.

Necesito proteger las rutas privadas del backend con un token JWT y restringirlas por
rol (postulante, analista y administrador), para que cada módulo lo use sin esperar al
login.

Crea:
1. Una interfaz TokenIssuer para emitir y validar tokens firmados con JWT_SECRET, con
   la expiración de JWT_TTL_MINUTES, y su implementación con JWT.
2. Un middleware que exija "Authorization: Bearer <token>" y guarde la identidad del
   usuario en el contexto. Sin token o con uno inválido debe responder 401.
3. Un middleware que restrinja por rol y responda 403 si el rol no coincide.
4. Un helper para leer el usuario actual desde un handler.
5. La inyección del emisor de tokens en main.go, y un mínimo de 32 caracteres para
   JWT_SECRET.

Agrega pruebas: token válido, vencido, firma alterada y rol incorrecto.

La issue #4 del repositorio tiene el detalle completo (archivos a tocar y criterios de
terminado): revísala para cumplirla exactamente.
```

## Para qué se usó
Construir la autenticación con JWT y los middlewares de sesión y de rol, con sus pruebas.

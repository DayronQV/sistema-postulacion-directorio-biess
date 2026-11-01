package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

const claimsKey = "auth_claims"

// RequireAuth exige un token válido en la cabecera "Authorization: Bearer <token>".
// Si falta o no sirve responde 401 UNAUTHORIZED; si sirve, guarda la identidad en el
// contexto para leerla con ClaimsFrom.
func RequireAuth(tokens domain.TokenIssuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			Fail(c, domain.Unauthorized("Debe iniciar sesión para continuar"))
			return
		}
		claims, err := tokens.Validar(raw)
		if err != nil {
			Fail(c, err)
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

// RequireRole deja pasar solo a los usuarios con alguno de los roles indicados; a los
// demás les responde 403 FORBIDDEN. Debe ir después de RequireAuth: si no hay sesión
// responde 401.
func RequireRole(roles ...domain.Rol) gin.HandlerFunc {
	if len(roles) == 0 {
		// Error de programación al armar las rutas: se detecta al arrancar.
		panic("middleware.RequireRole necesita al menos un rol")
	}
	permitidos := make(map[domain.Rol]bool, len(roles))
	for _, r := range roles {
		permitidos[r] = true
	}
	return func(c *gin.Context) {
		claims, ok := ClaimsFrom(c)
		if !ok {
			Fail(c, domain.Unauthorized("Debe iniciar sesión para continuar"))
			return
		}
		if !permitidos[claims.Rol] {
			Fail(c, domain.Forbidden("No tiene permiso para realizar esta acción"))
			return
		}
		c.Next()
	}
}

// ClaimsFrom devuelve el usuario autenticado de la petición actual. El segundo valor es
// false si la ruta no pasó por RequireAuth.
func ClaimsFrom(c *gin.Context) (domain.Claims, bool) {
	v, ok := c.Get(claimsKey)
	if !ok {
		return domain.Claims{}, false
	}
	claims, ok := v.(domain.Claims)
	return claims, ok
}

// bearerToken extrae el token de una cabecera "Bearer <token>".
func bearerToken(header string) (string, bool) {
	scheme, token, found := strings.Cut(strings.TrimSpace(header), " ")
	token = strings.TrimSpace(token)
	if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}

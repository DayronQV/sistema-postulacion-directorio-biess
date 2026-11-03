package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// MaxLargoUsuarioAdmin es el largo máximo del usuario: evita claves de Redis de tamaño
// arbitrario en el contador de intentos.
const MaxLargoUsuarioAdmin = 50

// AdminLoginInput son las credenciales de un analista o administrador.
type AdminLoginInput struct {
	Usuario  string
	Password string
}

// AdminSesion es la sesión emitida para el panel de administración.
type AdminSesion struct {
	Token    string
	ExpiraEn time.Duration
	Usuario  string
	Rol      domain.Rol
}

// AdminAuthService implementa el inicio de sesión del panel de administración, con el
// mismo bloqueo que el login del postulante: al tercer fallo, 10 minutos.
type AdminAuthService struct {
	usuarios domain.UsuariosAdmin
	intentos domain.Intentos // instancia propia: auth:*:admin:{usuario}
	tokens   domain.TokenIssuer
	// hashFicticio se compara cuando el usuario no existe, para que esa respuesta tarde lo
	// mismo que una contraseña incorrecta y no delate qué usuarios existen.
	hashFicticio []byte
}

// NewAdminAuthService crea el servicio con sus dependencias inyectadas.
func NewAdminAuthService(usuarios domain.UsuariosAdmin, intentos domain.Intentos, tokens domain.TokenIssuer) *AdminAuthService {
	hash, err := bcrypt.GenerateFromPassword([]byte("contraseña-ficticia-sin-usuario"), bcrypt.DefaultCost)
	if err != nil {
		// Solo falla si el costo no es válido, y DefaultCost lo es.
		panic(fmt.Sprintf("generando el hash ficticio: %v", err))
	}
	return &AdminAuthService{usuarios: usuarios, intentos: intentos, tokens: tokens, hashFicticio: hash}
}

// Login valida usuario y contraseña y emite un token con el rol del usuario (analista o
// administrador). Un usuario inexistente responde igual que una contraseña incorrecta y
// también cuenta como intento fallido; al tercero el usuario queda bloqueado 10 minutos.
func (s *AdminAuthService) Login(ctx context.Context, in AdminLoginInput) (AdminSesion, error) {
	usuario := strings.ToLower(strings.TrimSpace(in.Usuario))
	fields := map[string]string{}
	if usuario == "" {
		fields["usuario"] = "Campo obligatorio"
	} else if len([]rune(usuario)) > MaxLargoUsuarioAdmin {
		fields["usuario"] = fmt.Sprintf("Máximo %d caracteres", MaxLargoUsuarioAdmin)
	}
	if in.Password == "" {
		fields["password"] = "Campo obligatorio"
	}
	if len(fields) > 0 {
		return AdminSesion{}, domain.Validation("Datos inválidos", fields)
	}

	restante, err := s.intentos.Bloqueo(ctx, usuario)
	if err != nil {
		return AdminSesion{}, err
	}
	if restante > 0 {
		return AdminSesion{}, domain.Locked(MsgLoginBloqueado, restante)
	}

	u, err := s.usuarios.BuscarPorUsuario(ctx, usuario)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		_ = bcrypt.CompareHashAndPassword(s.hashFicticio, []byte(in.Password))
		return AdminSesion{}, s.fallo(ctx, usuario)
	case err != nil:
		return AdminSesion{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)) != nil {
		return AdminSesion{}, s.fallo(ctx, usuario)
	}
	// La tabla solo admite analista y administrador; si llegara otro valor no se emite
	// un token que el resto del sistema no sabría tratar.
	if u.Rol != domain.RolAnalista && u.Rol != domain.RolAdministrador {
		return AdminSesion{}, fmt.Errorf("rol de administración no válido: %q", u.Rol)
	}

	if err := s.intentos.Reiniciar(ctx, usuario); err != nil {
		return AdminSesion{}, err
	}
	emitido, err := s.tokens.Emitir(domain.Claims{UsuarioID: u.ID, Usuario: u.Usuario, Rol: u.Rol})
	if err != nil {
		return AdminSesion{}, err
	}
	return AdminSesion{Token: emitido.Token, ExpiraEn: emitido.ExpiraEn, Usuario: u.Usuario, Rol: u.Rol}, nil
}

// fallo registra un intento fallido y bloquea al usuario al llegar al máximo.
func (s *AdminAuthService) fallo(ctx context.Context, usuario string) error {
	n, err := s.intentos.RegistrarFallo(ctx, usuario, VentanaIntentosLogin)
	if err != nil {
		return err
	}
	if n >= MaxIntentosLogin {
		if err := s.intentos.Bloquear(ctx, usuario, BloqueoLogin); err != nil {
			return err
		}
		return domain.Locked(MsgLoginBloqueado, BloqueoLogin)
	}
	return domain.Unauthorized(MsgCredencialesIncorrectas)
}

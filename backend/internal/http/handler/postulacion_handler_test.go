package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
)

type fakePostulacion struct {
	obtenerID string
	guardarID string
	guardado  domain.Formulario
	resultado domain.Postulacion
	err       error
}

func (f *fakePostulacion) Obtener(_ context.Context, id string) (domain.Postulacion, error) {
	f.obtenerID = id
	return f.resultado, f.err
}

func (f *fakePostulacion) Guardar(_ context.Context, id string, fm domain.Formulario) (domain.Postulacion, error) {
	f.guardarID = id
	f.guardado = fm
	return f.resultado, f.err
}

// fakeTokens valida dos tokens fijos: "postulante" y "analista".
type fakeTokens struct{}

func (fakeTokens) Emitir(domain.Claims) (domain.TokenEmitido, error) {
	return domain.TokenEmitido{}, errors.New("no se usa")
}

func (fakeTokens) Validar(token string) (domain.Claims, error) {
	switch token {
	case "postulante":
		return domain.Claims{UsuarioID: "p-1", Cedula: "1712345675", Rol: domain.RolPostulante}, nil
	case "analista":
		return domain.Claims{UsuarioID: "a-1", Usuario: "analista1", Rol: domain.RolAnalista}, nil
	}
	return domain.Claims{}, domain.Unauthorized("token inválido")
}

func newPostulacionRouter(uc handler.PostulacionUseCase) *gin.Engine {
	return apihttp.NewRouter(apihttp.Deps{
		Config:  config.Config{RequestTimeout: time.Second},
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health:  handler.NewHealthHandler(),
		Modules: []apihttp.RouteRegistrar{apihttp.PostulacionRoutes(handler.NewPostulacionHandler(uc), fakeTokens{})},
	})
}

func send(r *gin.Engine, method, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/v1/postulacion", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func postulacionDeEjemplo() domain.Postulacion {
	return domain.Postulacion{
		ID: "post-1", Codigo: "BIESS-2026-000001", Estado: domain.EstadoBorrador,
		Personal: domain.InformacionPersonal{
			Nombres: "Ana", Apellidos: "Pérez", Genero: "Femenino", Nacionalidad: "Ecuatoriana",
			FechaNacimiento: time.Date(1990, 3, 4, 0, 0, 0, 0, time.UTC), EstadoCivil: domain.EstadoSoltero,
		},
		Formulario: domain.Formulario{Cargo: domain.CargoAfiliados},
	}
}

func TestPostulacionObtener_OK(t *testing.T) {
	uc := &fakePostulacion{resultado: postulacionDeEjemplo()}
	w := send(newPostulacionRouter(uc), "GET", "postulante", "")
	if w.Code != 200 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	var got struct {
		Codigo              string `json:"codigo"`
		Estado              string `json:"estado"`
		Cargo               string `json:"cargo"`
		Familiares          []any  `json:"familiares"`
		Referencias         []any  `json:"referencias"`
		Formacion           []any  `json:"formacion"`
		InformacionPersonal struct {
			Nombres         string `json:"nombres"`
			FechaNacimiento string `json:"fecha_nacimiento"`
		} `json:"informacion_personal"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Codigo != "BIESS-2026-000001" || got.Estado != "borrador" || got.Cargo != "afiliados" ||
		got.InformacionPersonal.Nombres != "Ana" || got.InformacionPersonal.FechaNacimiento != "1990-03-04" {
		t.Fatalf("cuerpo inesperado: %s", w.Body.String())
	}
	if got.Familiares == nil {
		t.Fatalf("familiares debe ser [], no null: %s", w.Body.String())
	}
	if got.Referencias == nil {
		t.Fatalf("referencias debe ser [], no null: %s", w.Body.String())
	}
	if got.Formacion == nil {
		t.Fatalf("formacion debe ser [], no null: %s", w.Body.String())
	}
	if uc.obtenerID != "p-1" {
		t.Fatalf("el handler no pasó el postulante autenticado: %q", uc.obtenerID)
	}
}

func TestPostulacion_ExigeSesionDePostulante(t *testing.T) {
	casos := []struct {
		nombre string
		token  string
		status int
		code   string
	}{
		{"sin token", "", 401, "UNAUTHORIZED"},
		{"token inválido", "otro", 401, "UNAUTHORIZED"},
		{"rol incorrecto", "analista", 403, "FORBIDDEN"},
	}
	for _, c := range casos {
		for _, method := range []string{"GET", "PUT"} {
			t.Run(c.nombre+" "+method, func(t *testing.T) {
				uc := &fakePostulacion{}
				w := send(newPostulacionRouter(uc), method, c.token, `{}`)
				if w.Code != c.status || !strings.Contains(w.Body.String(), `"code":"`+c.code+`"`) {
					t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
				}
				if uc.obtenerID != "" || uc.guardarID != "" {
					t.Fatalf("sin sesión de postulante no debe llegar al servicio")
				}
			})
		}
	}
}

func TestPostulacionGuardar_PasaElFormularioAlServicio(t *testing.T) {
	uc := &fakePostulacion{resultado: postulacionDeEjemplo()}
	body := `{
		"domicilio": {"pais_id": 1, "provincia_id": 17, "canton_id": 170, "telefono_celular": "0999999999"},
		"tiene_familiares": true,
		"familiares": [{"nombres_apellidos": "Luis", "parentesco": "primo", "institucion": "BIESS", "area": "TI", "cargo": "Jefe"}],
		"referencias": [{"nombres_apellidos": "Ana Pérez", "relacion": "Amiga", "telefono": "0999999999", "correo": "ana@example.com"}],
		"cargo": "afiliados",
		"respaldo_organizacion": {"nombre": "Asociación", "documento_id": null},
		"formacion": [{"nivel": "tercer_nivel", "institucion": "UPSE", "tipo": "nacional", "titulo": "Economista", "area": "economia", "senescyt_numero": "1001", "senescyt_fecha": "2020-01-31", "documento_id": "doc-1"}],
		"consentimiento": true
	}`
	w := send(newPostulacionRouter(uc), "PUT", "postulante", body)
	if w.Code != 200 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	g := uc.guardado
	if uc.guardarID != "p-1" || g.Cargo != "afiliados" || !g.TieneFamiliares || !g.Consentimiento ||
		g.Domicilio.PaisID == nil || *g.Domicilio.CantonID != 170 || g.Domicilio.TelefonoCelular != "0999999999" ||
		len(g.Familiares) != 1 || g.Familiares[0].Institucion != "BIESS" || len(g.Referencias) != 1 ||
		g.Referencias[0].Correo != "ana@example.com" || len(g.Formacion) != 1 ||
		g.Formacion[0].DocumentoID != "doc-1" || g.RespaldoOrganizacion.Nombre != "Asociación" {
		t.Fatalf("formulario mal traducido: %+v", g)
	}
}

func TestPostulacionGuardar_ErroresConElFormatoDeLaApi(t *testing.T) {
	casos := []struct {
		nombre string
		body   string
		err    error
		status int
		code   string
	}{
		{"conflicto", `{}`, domain.Conflict("ya enviada", nil), 409, "CONFLICT"},
		{"nacionalidad", `{}`, domain.Unprocessable("solo ecuatorianos", nil), 422, "VALIDATION_ERROR"},
		{"validación del servicio", `{}`, domain.Validation("Datos inválidos", map[string]string{"cargo": "Cargo no válido"}), 400, "VALIDATION_ERROR"},
		{"error interno", `{}`, errors.New("boom con detalles de SQL"), 500, "INTERNAL_ERROR"},
		{"JSON mal formado", `{`, nil, 400, "VALIDATION_ERROR"},
		{"tipo incorrecto", `{"tiene_familiares": "si"}`, nil, 400, "VALIDATION_ERROR"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			uc := &fakePostulacion{err: c.err}
			w := send(newPostulacionRouter(uc), "PUT", "postulante", c.body)
			if w.Code != c.status || !strings.Contains(w.Body.String(), `"code":"`+c.code+`"`) {
				t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
			}
			if strings.Contains(w.Body.String(), "SQL") {
				t.Fatalf("un 500 no debe exponer detalles internos: %s", w.Body.String())
			}
		})
	}
}

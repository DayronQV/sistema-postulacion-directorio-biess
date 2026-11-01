package domain

import "context"

// Calidad es lo que informa el Historial Laboral del IESS sobre una persona. Puede
// cumplirse más de una condición a la vez, o ninguna.
type Calidad struct {
	EsAfiliado  bool // "Postulante afiliado al IESS" / "Postulante No Afiliado al IESS"
	EsJubilado  bool // "Postulante acredita jubilación" / "Postulante no acredita jubilación"
	EsEmpleador bool // "Postulante es empleador" / "Postulante no es empleador"
}

// HistorialLaboral es el puerto hacia el Historial Laboral del IESS. Hoy se simula
// con un adaptador mock (internal/integration/historiallaboral); si algún día hubiera
// acceso real, solo se cambia el adaptador.
type HistorialLaboral interface {
	// ConsultarCalidad devuelve si la cédula es de un afiliado, un jubilado o un
	// empleador. Quien lo use decide qué condición exige según el cargo elegido.
	ConsultarCalidad(ctx context.Context, cedula string) (Calidad, error)
}

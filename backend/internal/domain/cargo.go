package domain

// ResultadoCargo es el resultado de validar con el Historial Laboral si el postulante
// puede aspirar al cargo elegido.
type ResultadoCargo struct {
	Cargo    string
	Elegible bool
	// Mensaje es el texto del requerimiento: "Postulante afiliado al IESS" (y sus
	// equivalentes) si es elegible, o el mensaje de cierre si no lo es.
	Mensaje string
}

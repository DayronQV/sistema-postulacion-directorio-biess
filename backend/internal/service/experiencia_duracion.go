package service

import (
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// completarDuraciones calcula con la calculadora del dominio el tiempo de cada experiencia
// y el total acumulado. Es lógica de negocio: el repositorio solo carga las fechas y el
// servicio la aplica después de leer las secciones (GET /postulacion y la respuesta del PUT).
func completarDuraciones(f *domain.Formulario) {
	duraciones := make([]domain.Duracion, 0, len(f.Experiencia))
	for i := range f.Experiencia {
		e := &f.Experiencia[i]
		e.Duracion = domain.Duracion{}
		inicio, errInicio := time.Parse(time.DateOnly, e.FechaInicio)
		fin, errFin := time.Parse(time.DateOnly, e.FechaFin)
		if errInicio != nil || errFin != nil {
			continue // no ocurre con datos guardados: la base exige fechas válidas
		}
		e.Duracion.Anios, e.Duracion.Meses, e.Duracion.Dias = domain.CalcularDuracion(inicio, fin)
		duraciones = append(duraciones, e.Duracion)
	}
	f.DuracionExperiencia = domain.SumarDuraciones(duraciones)
}

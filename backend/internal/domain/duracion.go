package domain

import "time"

// Duracion expresa tiempo calendario completo, sin aproximar los meses a una cantidad
// fija de días.
type Duracion struct {
	Anios int
	Meses int
	Dias  int
}

// CalcularDuracion devuelve la diferencia calendario entre dos fechas. La validación
// de la sección impide que fin sea anterior a inicio; ante ese dato inválido devuelve
// cero para mantener esta función total y segura.
func CalcularDuracion(inicio, fin time.Time) (anios, meses, dias int) {
	inicio = soloFecha(inicio)
	fin = soloFecha(fin)
	if fin.Before(inicio) {
		return 0, 0, 0
	}

	anios = fin.Year() - inicio.Year()
	meses = int(fin.Month()) - int(inicio.Month())
	dias = fin.Day() - inicio.Day()
	if dias < 0 {
		mesAnterior := fin.AddDate(0, 0, -fin.Day())
		dias += mesAnterior.Day()
		meses--
	}
	if meses < 0 {
		meses += 12
		anios--
	}
	return anios, meses, dias
}

func soloFecha(fecha time.Time) time.Time {
	return time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, time.UTC)
}

// SumarDuraciones acumula los periodos mostrados. Para normalizar el total, 30 días
// forman un mes y 12 meses forman un año; esto no altera el cálculo calendario de cada
// experiencia.
func SumarDuraciones(duraciones []Duracion) Duracion {
	var total Duracion
	for _, duracion := range duraciones {
		total.Anios += duracion.Anios
		total.Meses += duracion.Meses
		total.Dias += duracion.Dias
	}
	total.Meses += total.Dias / 30
	total.Dias %= 30
	total.Anios += total.Meses / 12
	total.Meses %= 12
	return total
}

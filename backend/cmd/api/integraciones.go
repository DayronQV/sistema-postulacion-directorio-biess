package main

import (
	"log/slog"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/historiallaboral"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
)

// integraciones crea los simuladores del Registro Civil y del Historial Laboral. Sin
// PADRON_ARCHIVO usan sus ciudadanos de prueba; con él, solo las personas de ese archivo.
// Un archivo ausente o inválido detiene el arranque.
func integraciones(archivo string, logger *slog.Logger) (domain.RegistroCivil, domain.HistorialLaboral, error) {
	if archivo == "" {
		return registrocivil.NewMock(), historiallaboral.NewMock(), nil
	}
	registro, err := registrocivil.NewDesdeArchivo(archivo)
	if err != nil {
		return nil, nil, err
	}
	historial, err := historiallaboral.NewDesdeArchivo(archivo)
	if err != nil {
		return nil, nil, err
	}
	logger.Info("simuladores con las personas del archivo de padrón", "archivo", archivo)
	return registro, historial, nil
}

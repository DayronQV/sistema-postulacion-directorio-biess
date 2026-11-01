package domain

import "context"

// Tipos de documento admitidos por las secciones de la postulación.
const (
	DocumentoOrganizacion = "organizacion"
	DocumentoFormacion    = "formacion"
	DocumentoExperiencia  = "experiencia"
	DocumentoPublicacion  = "publicacion"
	DocumentoCapacitacion = "capacitacion"
)

// Documento es un archivo de respaldo perteneciente a una postulación.
type Documento struct {
	ID        string
	Tipo      string
	Nombre    string
	Tamano    int64
	Contenido []byte
}

// Documentos es el puerto de persistencia de los respaldos PDF.
type Documentos interface {
	Guardar(ctx context.Context, postulanteID string, documento Documento, limiteTotal int64) (Documento, error)
	ObtenerPropio(ctx context.Context, postulanteID, documentoID string) (Documento, error)
	EliminarPropioNoReferenciado(ctx context.Context, postulanteID, documentoID string) error
}

package domain

// Límites de la sección de referencias personales (issue #31): de 1 a 3 contactos. El
// mínimo se exige al enviar la postulación; en borrador la lista puede estar vacía.
const (
	MaxReferencias                = 3
	MaxLongitudNombreReferencia   = 150
	MaxLongitudRelacionReferencia = 60
	MaxLongitudCorreoReferencia   = 254
)

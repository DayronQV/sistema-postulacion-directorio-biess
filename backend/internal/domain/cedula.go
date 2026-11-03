package domain

// CedulaValida indica si es una cédula ecuatoriana bien formada:
//   - 10 dígitos;
//   - código de provincia entre 01 y 24;
//   - tercer dígito menor que 6;
//   - dígito verificador correcto (módulo 10).
//
// Solo comprueba el formato y el dígito verificador; no consulta si la persona existe.
func CedulaValida(cedula string) bool {
	if len(cedula) != 10 {
		return false
	}
	var d [10]int
	for i := 0; i < 10; i++ {
		c := cedula[i]
		if c < '0' || c > '9' {
			return false
		}
		d[i] = int(c - '0')
	}

	if provincia := d[0]*10 + d[1]; provincia < 1 || provincia > 24 {
		return false
	}
	if d[2] >= 6 {
		return false
	}

	// Módulo 10: los dígitos en posición impar (1.º, 3.º, ...) se multiplican por 2 y
	// los demás por 1; si el producto pasa de 9 se le resta 9.
	suma := 0
	for i := 0; i < 9; i++ {
		n := d[i]
		if i%2 == 0 {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		suma += n
	}
	verificador := (10 - suma%10) % 10
	return verificador == d[9]
}

// ValidarCedula devuelve nil si la cédula es válida, o un error de validación (400)
// con el campo `cedula`, en el formato de docs/api.md.
func ValidarCedula(cedula string) error {
	if CedulaValida(cedula) {
		return nil
	}
	return Validation("Datos inválidos", map[string]string{"cedula": "Cédula no válida"})
}

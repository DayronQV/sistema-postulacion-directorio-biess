// Validación de cédula ecuatoriana, la misma regla que el servidor
// (backend/internal/domain/cedula.go): 10 dígitos, provincia 01-24, tercer dígito
// menor que 6 y dígito verificador por módulo 10. Solo da aviso inmediato: el
// servidor vuelve a validar.
export function cedulaValida(cedula: string): boolean {
  if (!/^\d{10}$/.test(cedula)) return false
  const d = [...cedula].map(Number)
  const provincia = d[0] * 10 + d[1]
  if (provincia < 1 || provincia > 24 || d[2] >= 6) return false

  let suma = 0
  for (let i = 0; i < 9; i++) {
    let n = d[i]
    if (i % 2 === 0) {
      n *= 2
      if (n > 9) n -= 9
    }
    suma += n
  }
  return (10 - (suma % 10)) % 10 === d[9]
}

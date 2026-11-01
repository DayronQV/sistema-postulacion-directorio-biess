// Une clases CSS ignorando valores falsos.
export function cx(...partes: Array<string | false | null | undefined>): string {
  return partes.filter(Boolean).join(' ')
}

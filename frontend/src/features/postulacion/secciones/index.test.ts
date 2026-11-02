import type { Postulacion } from '@/services/postulacion'
import { SECCIONES } from '.'

const completa = (id: string, p: Partial<Postulacion>) =>
  SECCIONES.find((s) => s.id === id)!.estaCompleta(p as Postulacion)

describe('marcas de sección completa', () => {
  it('Publicaciones: "No" basta, "Sí" exige al menos una publicación', () => {
    expect(completa('publicaciones', { tiene_publicaciones: false, publicaciones: [] })).toBe(true)
    expect(completa('publicaciones', { tiene_publicaciones: true, publicaciones: [] })).toBe(false)
    expect(
      completa('publicaciones', {
        tiene_publicaciones: true,
        publicaciones: [{} as Postulacion['publicaciones'][number]],
      }),
    ).toBe(true)
  })

  it('Información adicional: "Sí" exige al menos un familiar', () => {
    expect(completa('adicional', { tiene_familiares: true, familiares: [] })).toBe(false)
    expect(completa('adicional', { tiene_familiares: false, familiares: [] })).toBe(true)
  })
})

import { renderHook } from '@testing-library/react'
import type { InternalAxiosRequestConfig } from 'axios'
import { http } from '@/services/http'
import { guardarPostulacion, type Postulacion } from '@/services/postulacion'
import { cargarPostulacion, marcarCambios, reiniciarFormulario, useFormulario } from './store'

// Sin simular el servicio: el guardado real avisa al store, como cuando guarda una sección.
const inicial = { id: 'p1', codigo: 'A', estado: 'borrador', cargo: '' } as Postulacion
const guardada = { ...inicial, cargo: 'afiliados' } as Postulacion

function responder(config: InternalAxiosRequestConfig) {
  const data = config.method === 'put' ? guardada : inicial
  return Promise.resolve({ status: 200, statusText: '', headers: {}, config, data })
}

function estado() {
  return renderHook(() => useFormulario()).result.current
}

beforeEach(() => {
  reiniciarFormulario()
  http.defaults.adapter = responder
})

afterEach(() => {
  delete http.defaults.adapter
})

describe('store del formulario ante el guardado de una sección', () => {
  it('actualiza la postulación y quita las marcas de cambios sin guardar', async () => {
    await cargarPostulacion()
    marcarCambios('cargo', true)
    expect(estado().seccionesConCambios).toEqual(['cargo'])

    await guardarPostulacion({ ...inicial, cargo: 'afiliados' })

    expect(estado().postulacion?.cargo).toBe('afiliados')
    expect(estado().seccionesConCambios).toEqual([])
  })

  it('sin postulación cargada ignora el aviso', async () => {
    await guardarPostulacion(guardada)
    expect(estado().postulacion).toBeNull()
  })
})

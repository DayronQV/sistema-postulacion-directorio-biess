import type { InternalAxiosRequestConfig } from 'axios'
import {
  descargarDocumentoAdmin,
  descargarMatrizAdmin,
  listarPostulacionesAdmin,
  marcarMerito,
  obtenerPostulacionAdmin,
} from './adminPostulaciones'
import { http } from './http'

afterEach(() => {
  delete http.defaults.adapter
})

describe('servicio de postulaciones del panel', () => {
  it('marcarMerito envía PUT /admin/postulaciones/{id}/merito con supero', async () => {
    const peticiones: InternalAxiosRequestConfig[] = []
    const respuesta = { id: 'p1', supero_merito: true, estado_visible: 'Evaluación Técnica' }
    http.defaults.adapter = (config) => {
      peticiones.push(config)
      return Promise.resolve({ status: 200, statusText: '', headers: {}, config, data: respuesta })
    }

    const r = await marcarMerito(new AbortController().signal, 'p1', true)

    expect(r).toEqual(respuesta)
    expect(peticiones[0]).toMatchObject({ method: 'put', url: '/admin/postulaciones/p1/merito' })
    expect(JSON.parse(peticiones[0].data)).toEqual({ supero: true })
  })

  it('lista con filtros y página codificados en la URL', async () => {
    const peticiones: InternalAxiosRequestConfig[] = []
    const respuesta = { items: [], total: 0, pagina: 2, por_pagina: 20 }
    http.defaults.adapter = (config) => {
      peticiones.push(config)
      return Promise.resolve({ status: 200, statusText: '', headers: {}, config, data: respuesta })
    }

    await expect(
      listarPostulacionesAdmin(new AbortController().signal, {
        cargo: 'afiliados',
        estado: 'enviada',
        elegible: 'true',
        pagina: 2,
      }),
    ).resolves.toEqual(respuesta)
    expect(peticiones[0].url).toBe(
      '/admin/postulaciones?cargo=afiliados&estado=enviada&elegible=true&pagina=2',
    )
  })

  it('obtiene el detalle completo por id', async () => {
    const peticiones: InternalAxiosRequestConfig[] = []
    const respuesta = { id: 'p 1', codigo: 'BIESS-2026-000001' }
    http.defaults.adapter = (config) => {
      peticiones.push(config)
      return Promise.resolve({ status: 200, statusText: '', headers: {}, config, data: respuesta })
    }

    await expect(obtenerPostulacionAdmin(new AbortController().signal, 'p 1')).resolves.toEqual(
      respuesta,
    )
    expect(peticiones[0].url).toBe('/admin/postulaciones/p%201')
  })

  it('descarga el certificado y conserva el nombre de la cabecera', async () => {
    const peticiones: InternalAxiosRequestConfig[] = []
    const pdf = new Blob(['%PDF'])
    http.defaults.adapter = (config) => {
      peticiones.push(config)
      return Promise.resolve({
        status: 200,
        statusText: '',
        headers: { 'content-disposition': "attachment; filename*=UTF-8''t%C3%ADtulo.pdf" },
        config,
        data: pdf,
      })
    }

    await expect(
      descargarDocumentoAdmin(new AbortController().signal, 'p 1', 'd 1'),
    ).resolves.toEqual({ contenido: pdf, nombre: 'título.pdf' })
    expect(peticiones[0]).toMatchObject({
      method: 'get',
      url: '/admin/postulaciones/p%201/documentos/d%201',
      responseType: 'blob',
    })
  })

  it('descarga la matriz con el formato y los filtros activos', async () => {
    const peticiones: InternalAxiosRequestConfig[] = []
    const archivo = new Blob(['xlsx'])
    http.defaults.adapter = (config) => {
      peticiones.push(config)
      return Promise.resolve({
        status: 200,
        statusText: '',
        headers: { 'content-disposition': 'attachment; filename="matriz.xlsx"' },
        config,
        data: archivo,
      })
    }

    await expect(
      descargarMatrizAdmin(new AbortController().signal, 'xlsx', {
        cargo: 'afiliados',
        estado: 'enviada',
        elegible: 'true',
        pagina: 3,
      }),
    ).resolves.toEqual({ contenido: archivo, nombre: 'matriz.xlsx' })
    expect(peticiones[0]).toMatchObject({
      method: 'get',
      url: '/admin/matriz?formato=xlsx&cargo=afiliados&estado=enviada&elegible=true',
      responseType: 'blob',
    })
  })
})

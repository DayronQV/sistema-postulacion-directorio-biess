import indexHtml from '../../../index.html?raw'

// index.html pide a Google Fonts solo los iconos que usa el sistema (icon_names), para que la
// fuente pese ~65 KB y no ~4 MB. Un icono que falte en esa lista se vería escrito como texto
// ("badge", "home"), así que esta prueba recorre el código y compara.

// Todo el código del frontend como texto (Vite lo lee al correr la prueba), sin las pruebas.
const fuentes = Object.entries(
  import.meta.glob<string>(['/src/**/*.{ts,tsx}', '!/src/**/*.test.{ts,tsx}'], {
    query: '?raw',
    import: 'default',
    eager: true,
  }),
)
const NOMBRE_ICONO = /^[a-z][a-z0-9_]*$/

function literales(expresion: string): string[] {
  return [...expresion.matchAll(/'([^']+)'/g)].map((m) => m[1])
}

// Las cuatro formas en que el código nombra un icono.
function iconosUsados(): string[] {
  const usados = new Set<string>()
  for (const [ruta, texto] of fuentes) {
    // <Icon nombre="x" /> o <Icon nombre={cond ? 'x' : 'y'} />
    for (const m of texto.matchAll(/<Icon\b[^>]*?\bnombre=(?:"([^"]+)"|\{([^}]*)\})/gs)) {
      if (m[1]) usados.add(m[1])
      else literales(m[2]).forEach((n) => usados.add(n))
    }
    // Props que los componentes pasan a <Icon>: icono="x", iconoEtiqueta="x", icono={cond ? ...}
    for (const m of texto.matchAll(/\b(?:icono|iconoEtiqueta)=(?:"([^"]+)"|\{([^}]*)\})/g)) {
      if (m[1]) usados.add(m[1])
      else literales(m[2]).forEach((n) => usados.add(n))
    }
    // Listas como el menú o las secciones: { icono: 'x' }
    for (const m of texto.matchAll(/\bicono:\s*'([^']+)'/g)) usados.add(m[1])
    // Los avisos (Alert) eligen el icono por tipo: { nombre: 'x' }
    if (ruta.endsWith('Alert.tsx')) {
      for (const m of texto.matchAll(/\bnombre:\s*'([^']+)'/g)) usados.add(m[1])
    }
  }
  // Se descartan las clases de color que también se llaman "icono" (text-success, ...).
  return [...usados].filter((n) => NOMBRE_ICONO.test(n)).sort()
}

function iconosPedidos(): string[] {
  const m = indexHtml.match(/Material\+Symbols\+Outlined[^"]*[?&]icon_names=([^&"]+)/)
  if (!m) throw new Error('index.html no pide los iconos con icon_names')
  return m[1].split(',')
}

describe('iconos de Material Symbols', () => {
  it('index.html pide todos los iconos que usa el código', () => {
    const pedidos = new Set(iconosPedidos())
    const faltan = iconosUsados().filter((n) => !pedidos.has(n))
    expect(faltan, 'agregue estos iconos a icon_names en index.html').toEqual([])
  })

  it('la lista va en orden alfabético y sin repetidos, como exige Google Fonts', () => {
    const pedidos = iconosPedidos()
    expect(pedidos).toEqual([...new Set(pedidos)].sort())
  })

  it('reconoce las cuatro formas de nombrar un icono', () => {
    const usados = iconosUsados()
    // Una de cada forma: <Icon nombre>, una condición en <Icon>, un prop icono, una lista, Alert.
    for (const n of ['logout', 'visibility_off', 'refresh', 'undo', 'home', 'warning']) {
      expect(usados).toContain(n)
    }
    expect(usados).not.toContain('text-success')
  })
})

// Temas del IDE. Cada uno define la paleta completa (interfaz, editor y el
// color de cada tipo de dato) y el Pokémon que aparece en la pantalla de
// título. Se aplican como variables CSS sobre :root.

export const TEMAS = {
  'rojo-fuego': {
    nombre: 'Rojo Fuego',
    oscuro: false,
    escena: { fondo: 'volcanocave', particulas: 'brasas', mascota: 'charizard', vuela: true },
    colores: {
      fondo: '#f3e6cc',
      'fondo-2': '#ead8b4',
      panel: '#fffaf0',
      'panel-2': '#fbefd9',
      borde: '#3b2a2a',
      'borde-suave': '#d9c29a',
      texto: '#2d2222',
      'texto-suave': '#7a6452',
      acento: '#e3223a',
      'acento-texto': '#ffffff',
      'acento-2': '#ffcb05',
      exito: '#3fa34d',
      error: '#d62f2f',
      aviso: '#e08a00',
      'editor-fondo': '#fffdf7',
      'editor-gutter': '#f6ead2',
      'editor-texto': '#2d2222',
      'editor-linea': '#fff3dc',
      'editor-seleccion': '#ffd9a8',
      comentario: '#9a8a74',
      palabra: '#c0182c',
      control: '#8a3ab9',
      numero: '#b05a00',
      texto_lit: '#2e8b3e',
      constante: '#0f6fb0',
      nombre_tipo: '#a0522d',
      't-roca': '#8a6d2b',
      't-agua': '#2f6fd8',
      't-fuego': '#e2561b',
      't-planta': '#2f9a3a',
      't-electrico': '#c49a00',
      cielo1: '#ffd9a0',
      cielo2: '#ff8a5b',
      cielo3: '#b8384a',
    },
  },
  'game-boy': {
    nombre: 'Game Boy',
    oscuro: false,
    escena: { fondo: 'route', particulas: 'hojas', mascota: 'pikachu' },
    filtroSprites:
      'grayscale(1) sepia(1) hue-rotate(38deg) saturate(2.2) brightness(0.78) contrast(1.1)',
    colores: {
      fondo: '#9bbc0f',
      'fondo-2': '#8bac0f',
      panel: '#c4d86a',
      'panel-2': '#b3cc4a',
      borde: '#0f380f',
      'borde-suave': '#306230',
      texto: '#0f380f',
      'texto-suave': '#306230',
      acento: '#306230',
      'acento-texto': '#c4d86a',
      'acento-2': '#0f380f',
      exito: '#306230',
      error: '#0f380f',
      aviso: '#306230',
      'editor-fondo': '#c4d86a',
      'editor-gutter': '#b3cc4a',
      'editor-texto': '#0f380f',
      'editor-linea': '#b8d055',
      'editor-seleccion': '#8bac0f',
      comentario: '#56812a',
      palabra: '#0f380f',
      control: '#0f380f',
      numero: '#306230',
      texto_lit: '#306230',
      constante: '#0f380f',
      nombre_tipo: '#306230',
      't-roca': '#0f380f',
      't-agua': '#0f380f',
      't-fuego': '#0f380f',
      't-planta': '#0f380f',
      't-electrico': '#0f380f',
      cielo1: '#c4d86a',
      cielo2: '#9bbc0f',
      cielo3: '#306230',
    },
  },
  'torre-lavanda': {
    nombre: 'Torre Lavanda',
    oscuro: true,
    escena: {
      fondo: 'forest',
      particulas: 'luciernagas',
      mascota: 'gengar',
      filtro: 'hue-rotate(150deg) brightness(0.5) saturate(1.4)',
    },
    colores: {
      fondo: '#150f22',
      'fondo-2': '#1d1530',
      panel: '#231a38',
      'panel-2': '#2b2044',
      borde: '#08050f',
      'borde-suave': '#4a3a6e',
      texto: '#efe6ff',
      'texto-suave': '#a797c8',
      acento: '#b784f5',
      'acento-texto': '#1a1029',
      'acento-2': '#ff7ab0',
      exito: '#6ee7a0',
      error: '#ff6b8a',
      aviso: '#ffc46b',
      'editor-fondo': '#1a1329',
      'editor-gutter': '#1f1731',
      'editor-texto': '#e8ddff',
      'editor-linea': '#241a3a',
      'editor-seleccion': '#46336e',
      comentario: '#7a6c9c',
      palabra: '#ff7ab0',
      control: '#c89bff',
      numero: '#ffb86b',
      texto_lit: '#8ef0b4',
      constante: '#7ad4ff',
      nombre_tipo: '#ffd27a',
      't-roca': '#d8b878',
      't-agua': '#6ab4ff',
      't-fuego': '#ff8c5a',
      't-planta': '#7ee07e',
      't-electrico': '#ffe066',
      cielo1: '#3a2560',
      cielo2: '#1d1236',
      cielo3: '#0b0716',
    },
  },
  'cueva-celeste': {
    nombre: 'Cueva Celeste',
    oscuro: true,
    escena: {
      fondo: 'icecave',
      particulas: 'nieve',
      mascota: 'mewtwo',
      filtro: 'brightness(0.62) saturate(1.2) hue-rotate(15deg)',
    },
    colores: {
      fondo: '#0c1526',
      'fondo-2': '#111d33',
      panel: '#15233e',
      'panel-2': '#1b2c4c',
      borde: '#050a14',
      'borde-suave': '#2e4670',
      texto: '#e2ecff',
      'texto-suave': '#8ea4cc',
      acento: '#58c4ff',
      'acento-texto': '#08121f',
      'acento-2': '#c77dff',
      exito: '#5ee6a8',
      error: '#ff6b7a',
      aviso: '#ffc86b',
      'editor-fondo': '#0f1a2e',
      'editor-gutter': '#121f36',
      'editor-texto': '#dbe7ff',
      'editor-linea': '#16253f',
      'editor-seleccion': '#28406a',
      comentario: '#5f7599',
      palabra: '#c77dff',
      control: '#58c4ff',
      numero: '#ffb070',
      texto_lit: '#7eeab0',
      constante: '#ff9ad0',
      nombre_tipo: '#ffd580',
      't-roca': '#d0b27a',
      't-agua': '#58a8ff',
      't-fuego': '#ff8a5c',
      't-planta': '#74dc84',
      't-electrico': '#ffe066',
      cielo1: '#28345e',
      cielo2: '#141d3a',
      cielo3: '#070b1a',
    },
  },
  'pueblo-paleta': {
    nombre: 'Pueblo Paleta',
    oscuro: false,
    escena: { fondo: 'meadow', particulas: 'petalos', mascota: 'bulbasaur' },
    colores: {
      fondo: '#e6f0d2',
      'fondo-2': '#d6e6bc',
      panel: '#fffdf4',
      'panel-2': '#f3f7e6',
      borde: '#34402c',
      'borde-suave': '#b9cf95',
      texto: '#27301f',
      'texto-suave': '#66775a',
      acento: '#3f9d4a',
      'acento-texto': '#ffffff',
      'acento-2': '#f5c542',
      exito: '#2f9a3a',
      error: '#d6453a',
      aviso: '#d98a00',
      'editor-fondo': '#fffef8',
      'editor-gutter': '#f2f6e4',
      'editor-texto': '#27301f',
      'editor-linea': '#f3f8e2',
      'editor-seleccion': '#d6ebb0',
      comentario: '#8f9c7e',
      palabra: '#2f7d3a',
      control: '#8a4ab9',
      numero: '#b85c14',
      texto_lit: '#1f7aa8',
      constante: '#c03a6a',
      nombre_tipo: '#8a5a1a',
      't-roca': '#8a6d2b',
      't-agua': '#1f78d8',
      't-fuego': '#e0501a',
      't-planta': '#23903a',
      't-electrico': '#b88f00',
      cielo1: '#fffbe0',
      cielo2: '#cfe8a8',
      cielo3: '#7fb865',
    },
  },
};

// marco devuelve un SVG de 12x12 píxeles con esquinas escalonadas, para
// usarlo como border-image (se estira el centro y se conservan las esquinas).
function marco(borde, linea, relleno) {
  const r = (x, y, w, h, c) => `<rect x='${x}' y='${y}' width='${w}' height='${h}' fill='${c}'/>`;
  const svg =
    `<svg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12' shape-rendering='crispEdges'>` +
    r(2, 2, 8, 8, relleno) +
    r(2, 1, 8, 1, linea) +
    r(1, 2, 1, 8, linea) +
    r(10, 2, 1, 8, linea) +
    r(2, 10, 8, 1, linea) +
    r(2, 0, 8, 1, borde) +
    r(0, 2, 1, 8, borde) +
    r(11, 2, 1, 8, borde) +
    r(2, 11, 8, 1, borde) +
    r(1, 1, 1, 1, borde) +
    r(10, 1, 1, 1, borde) +
    r(1, 10, 1, 1, borde) +
    r(10, 10, 1, 1, borde) +
    '</svg>';
  return `url("data:image/svg+xml,${encodeURIComponent(svg)}")`;
}

export function aplicarTema(id) {
  const tema = temaActual(id);
  const raiz = document.documentElement;
  for (const [k, v] of Object.entries(tema.colores)) raiz.style.setProperty(`--${k}`, v);
  raiz.style.setProperty('--filtro-sprites', tema.filtroSprites ?? 'none');
  raiz.style.setProperty('--filtro-escena', tema.escena.filtro ?? tema.filtroSprites ?? 'none');
  const c = tema.colores;
  raiz.style.setProperty('--marco', marco(c.borde, c['borde-suave'], c.panel));
  raiz.style.setProperty('--marco-acento', marco(c.borde, c['acento-2'], c.acento));
  raiz.style.setProperty('--marco-oscuro', marco('#101018', '#58587a', '#282838'));
  raiz.dataset.tema = id;
  raiz.dataset.oscuro = tema.oscuro ? 'si' : 'no';
}

// temaActual devuelve el tema pedido o el de siempre si no existe (por
// ejemplo, un perfil guardado con un tema que ya se quitó).
export function temaActual(id) {
  return TEMAS[id] ?? TEMAS['rojo-fuego'];
}

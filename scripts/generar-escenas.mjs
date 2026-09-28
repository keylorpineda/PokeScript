// Genera las escenas animadas del README: la Pokédex, el combate de tipos,
// las medallas de la hoja de ruta, el recorrido de un programa por el
// compilador y los Pokémon que explican cada sección.
// Uso: pnpm docs:svg (escribe en docs/assets/).
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { inflateSync } from 'node:zlib';

const OUT = process.argv[2] ?? fileURLToPath(new URL('../docs/assets/', import.meta.url));
const MONO = "Consolas, 'Courier New', monospace";
const SANS = "'Trebuchet MS', 'Segoe UI', Verdana, sans-serif";

// El arte de PokeAPI (dream world) se copia como vectores dentro de cada
// escena: GitHub sirve las imágenes con una política que bloquea cualquier
// imagen incrustada, así que un <image> se vería vacío. Se descarga con
// scripts/descargar-sprites.mjs.
const FUENTES = fileURLToPath(new URL('../docs/assets/fuentes/', import.meta.url));
function pokemon(id, x, y, w, h) {
  const src = readFileSync(join(FUENTES, 'pokemon', `${id}.svg`), 'utf8');
  const caja = src.match(/viewBox="([^"]+)"/)[1];
  const inicio = src.indexOf('>', src.indexOf('<svg')) + 1;
  return `<svg x="${x}" y="${y}" width="${w}" height="${h}" viewBox="${caja}">${src.slice(inicio, src.lastIndexOf('</svg>')).trim()}</svg>`;
}

// leerPng decodifica un PNG sin entrelazar (paleta o RGB/RGBA de 8 bits) y
// devuelve { ancho, alto, color(x, y) } con el color como '#rrggbb' o null
// si el píxel es transparente.
function leerPng(archivo) {
  const b = readFileSync(archivo);
  let pos = 8;
  let ancho = 0;
  let alto = 0;
  let prof = 8;
  let tipo = 6;
  let paleta = [];
  let alfa = [];
  const idat = [];
  while (pos < b.length) {
    const largoChunk = b.readUInt32BE(pos);
    const nombre = b.toString('ascii', pos + 4, pos + 8);
    const datos = b.subarray(pos + 8, pos + 8 + largoChunk);
    if (nombre === 'IHDR') {
      ancho = datos.readUInt32BE(0);
      alto = datos.readUInt32BE(4);
      prof = datos[8];
      tipo = datos[9];
      if (datos[12]) throw new Error(`${archivo}: PNG entrelazado`);
    } else if (nombre === 'PLTE') {
      for (let i = 0; i < datos.length; i += 3) paleta.push([datos[i], datos[i + 1], datos[i + 2]]);
    } else if (nombre === 'tRNS') {
      alfa = [...datos];
    } else if (nombre === 'IDAT') {
      idat.push(datos);
    }
    pos += 12 + largoChunk;
  }
  const canales = { 0: 1, 2: 3, 3: 1, 4: 2, 6: 4 }[tipo];
  const bitsPixel = canales * prof;
  const bpp = Math.max(1, bitsPixel >> 3);
  const fila = Math.ceil((ancho * bitsPixel) / 8);
  const crudo = inflateSync(Buffer.concat(idat));
  const px = Buffer.alloc(fila * alto);
  for (let y = 0; y < alto; y++) {
    const filtro = crudo[y * (fila + 1)];
    for (let i = 0; i < fila; i++) {
      const v = crudo[y * (fila + 1) + 1 + i];
      const a = i >= bpp ? px[y * fila + i - bpp] : 0;
      const u = y ? px[(y - 1) * fila + i] : 0;
      const c = y && i >= bpp ? px[(y - 1) * fila + i - bpp] : 0;
      let pred = 0;
      if (filtro === 1) pred = a;
      else if (filtro === 2) pred = u;
      else if (filtro === 3) pred = (a + u) >> 1;
      else if (filtro === 4) {
        const p = a + u - c;
        const pa = Math.abs(p - a);
        const pb = Math.abs(p - u);
        const pc = Math.abs(p - c);
        pred = pa <= pb && pa <= pc ? a : pb <= pc ? u : c;
      }
      px[y * fila + i] = (v + pred) & 255;
    }
  }
  const hex = (r, g, bl) => `#${[r, g, bl].map((n) => n.toString(16).padStart(2, '0')).join('')}`;
  return {
    ancho,
    alto,
    color(x, y) {
      if (tipo === 3) {
        const bit = x * prof;
        const byte = px[y * fila + (bit >> 3)];
        const idx = (byte >> (8 - prof - (bit & 7))) & ((1 << prof) - 1);
        if ((alfa[idx] ?? 255) < 128) return null;
        return hex(...paleta[idx]);
      }
      const o = y * fila + x * canales;
      if (canales === 4 && px[o + 3] < 128) return null;
      return hex(px[o], px[o + 1], px[o + 2]);
    },
  };
}

// sprite dibuja un sprite de los juegos como cuadritos, recortado a lo que
// tiene color y ajustado a la caja (x, y, w, h), apoyado en el borde de abajo.
function sprite(nombre, x, y, w, h) {
  const img = leerPng(join(FUENTES, 'pixeles', `${nombre}.png`));
  let [x0, y0, x1, y1] = [img.ancho, img.alto, 0, 0];
  for (let j = 0; j < img.alto; j++) {
    for (let i = 0; i < img.ancho; i++) {
      if (img.color(i, j))
        [x0, y0, x1, y1] = [Math.min(x0, i), Math.min(y0, j), Math.max(x1, i), Math.max(y1, j)];
    }
  }
  const cw = x1 - x0 + 1;
  const ch = y1 - y0 + 1;
  const k = Math.max(1, Math.floor(Math.min(w / cw, h / ch)));
  const ox = x + (w - cw * k) / 2;
  const oy = y + h - ch * k;
  let rects = '';
  for (let j = y0; j <= y1; j++) {
    for (let i = x0; i <= x1;) {
      const c = img.color(i, j);
      let n = 1;
      while (i + n <= x1 && img.color(i + n, j) === c) n++;
      if (c) rects += `<rect x="${i - x0}" y="${j - y0}" width="${n}" height="1" fill="${c}"/>`;
      i += n;
    }
  }
  return `<g transform="translate(${ox.toFixed(1)} ${oy.toFixed(1)}) scale(${k.toFixed(3)})" shape-rendering="crispEdges">${rects}</g>`;
}

const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
const largo = (s) => [...s].length;

// escritor arma líneas de texto que aparecen letra por letra: el texto se
// estira a un ancho fijo por carácter y una tapa del color del fondo se
// corre a saltos, una letra por paso.
function escritor(prefijo) {
  let css = '';
  let svg = '';
  let n = 0;
  return {
    linea(
      x,
      y,
      texto,
      { cw = 10, fs = 17, color, fondo, t, vel = 0.035, peso = 400, codigo = '#2458c8' },
    ) {
      // Lo que va entre comillas invertidas se pinta como código.
      const partes = texto.split('`');
      const contenido = partes
        .map((p, i) =>
          i % 2 ? `<tspan fill="${codigo}" font-weight="700">${esc(p)}</tspan>` : esc(p),
        )
        .join('');
      const chars = largo(partes.join(''));
      const ancho = chars * cw;
      const dur = Math.max(0.2, chars * vel);
      const id = `${prefijo}${n++}`;
      css += `    @keyframes ${id} { to { transform: translateX(${ancho + 4}px); } }\n`;
      svg += `    <text x="${x}" y="${y}" font-family="${MONO}" font-size="${fs}" font-weight="${peso}" fill="${color}" textLength="${ancho}" lengthAdjust="spacingAndGlyphs">${contenido}</text>\n`;
      svg += `    <rect x="${x - 2}" y="${y - fs}" width="${ancho + 6}" height="${fs + 8}" fill="${fondo}" style="animation: ${id} ${dur.toFixed(2)}s steps(${chars}, end) ${t.toFixed(2)}s both"/>\n`;
      return t + dur;
    },
    get css() {
      return css;
    },
    get svg() {
      return svg;
    },
  };
}

function pokebola(cx, cy, r, arriba, extra = '') {
  return `<g ${extra}>
      <circle cx="${cx}" cy="${cy}" r="${r}" fill="#ffffff"/>
      <path d="M${cx - r} ${cy} a${r} ${r} 0 0 1 ${2 * r} 0 z" fill="${arriba}"/>
      <rect x="${cx - r}" y="${cy - r * 0.1}" width="${2 * r}" height="${r * 0.2}" fill="#1b1b1b"/>
      <circle cx="${cx}" cy="${cy}" r="${r}" fill="none" stroke="#1b1b1b" stroke-width="${r * 0.1}"/>
      <circle cx="${cx}" cy="${cy}" r="${r * 0.33}" fill="#1b1b1b"/>
      <circle cx="${cx}" cy="${cy}" r="${r * 0.2}" fill="#ffffff"/>
    </g>`;
}

// ─── 1. Pokédex ────────────────────────────────────────────────────────────
// Una Pokédex abierta: a la izquierda escanea a Mewtwo, a la derecha escribe
// la entrada de PokeScript y llena las barras de sus fases.
{
  const W = 900;
  const H = 440;
  const e = escritor('pdx');
  const fondo = '#10231a';
  let t = 2.9;
  const lineas = [
    ['Nº 000 · POKESCRIPT', '#ffcb05', 700],
    ['TIPO: LENGUAJE / EDUCATIVO', '#7fd8a0', 700],
    ['', '', 400],
    ['Un lenguaje en español para aprender', '#d8f5e0', 400],
    ['a programar. Cada dato tiene un tipo', '#d8f5e0', 400],
    ['y, como en los combates, no todos los', '#d8f5e0', 400],
    ['tipos se llevan bien. Si algo falla,', '#d8f5e0', 400],
    ['te dice qué pasó y cómo arreglarlo.', '#d8f5e0', 400],
  ];
  lineas.forEach(([texto, color, peso], i) => {
    if (texto) {
      t =
        e.linea(500, 122 + i * 22, texto, { color, fondo, t, peso, cw: 8.6, fs: 15, vel: 0.025 }) +
        0.1;
    }
  });
  const fases = [
    ['LEXER', '#f8d030'],
    ['PARSER', '#f08030'],
    ['ANALIZADOR', '#a060d8'],
    ['INTÉRPRETE', '#58b848'],
  ];
  const barras = fases
    .map(
      ([nombre, color], i) => `
  <text x="500" y="${339 + i * 20}" font-family="${MONO}" font-size="12" font-weight="700" fill="#ffd9dd">${nombre}</text>
  <rect x="600" y="${329 + i * 20}" width="240" height="11" rx="5" fill="#7a0a15"/>
  <rect class="barra" style="animation-delay:${(t + 0.2 + i * 0.25).toFixed(2)}s" x="600" y="${329 + i * 20}" width="240" height="11" rx="5" fill="${color}"/>`,
    )
    .join('');
  const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="Pokédex abierta: escanea a Mewtwo y muestra la entrada de PokeScript, un lenguaje en español para aprender a programar">
  <defs>
    <clipPath id="datos"><rect x="486" y="96" width="364" height="190" rx="6"/></clipPath>
    <clipPath id="visor"><rect x="84" y="146" width="292" height="164" rx="6"/></clipPath>
    <radialGradient id="lente" cx="0.35" cy="0.35" r="0.7">
      <stop offset="0" stop-color="#b8e4ff"/><stop offset="0.45" stop-color="#3d9cf0"/><stop offset="1" stop-color="#15508f"/>
    </radialGradient>
    <linearGradient id="cuerpo" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0" stop-color="#e3223a"/><stop offset="1" stop-color="#b3101f"/>
    </linearGradient>
    <radialGradient id="aura" cx="0.5" cy="0.5" r="0.5">
      <stop offset="0" stop-color="#c77dff" stop-opacity="0.55"/><stop offset="1" stop-color="#c77dff" stop-opacity="0"/>
    </radialGradient>
  </defs>
  <style>
    .brillo { animation: brillo 2.4s ease-in-out infinite; }
    @keyframes brillo { 0%, 100% { opacity: 0.25; } 50% { opacity: 0.8; } }
    .led { animation: led 1.8s steps(1) infinite; }
    .l2 { animation-delay: 0.6s; } .l3 { animation-delay: 1.2s; }
    @keyframes led { 0% { opacity: 1; } 33% { opacity: 0.25; } }
    .escaneo { animation: escaneo 1.2s linear 0.3s 2 both; }
    @keyframes escaneo { from { transform: translateY(0); opacity: 1; } to { transform: translateY(160px); opacity: 1; } }
    .fin-escaneo { opacity: 1; animation: apaga 0.1s linear 2.7s forwards; }
    @keyframes apaga { to { opacity: 0; } }
    .silueta * { fill: #07140d !important; stroke: none !important; }
    .silueta { animation: revela 0.7s ease-out 2.5s forwards; }
    @keyframes revela { to { opacity: 0; } }
    .registrado { opacity: 0; animation: aparece 0.3s ease-out 2.7s forwards; }
    .escaneando { animation: parpadea 0.5s steps(1) 0s 5 both, apaga 0.1s linear 2.6s forwards; }
    @keyframes parpadea { 50% { opacity: 0.2; } }
    @keyframes aparece { to { opacity: 1; } }
    .flota { animation: flota 3s ease-in-out 3s infinite; }
    @keyframes flota { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-6px); } }
    .aura { transform-box: fill-box; transform-origin: center; opacity: 0; animation: aura 2.4s ease-in-out 3s infinite; }
    @keyframes aura { 0%, 100% { opacity: 0.4; transform: scale(0.9); } 50% { opacity: 1; transform: scale(1.08); } }
    .barra { transform-box: fill-box; transform-origin: left; transform: scaleX(0); animation: barra 0.9s cubic-bezier(.2,.8,.2,1) both; }
    @keyframes barra { to { transform: scaleX(1); } }
${e.css}  </style>

  <!-- mitad izquierda -->
  <path d="M44 20 H440 V420 H44 A24 24 0 0 1 20 396 V44 A24 24 0 0 1 44 20 Z" fill="url(#cuerpo)" stroke="#7a0a15" stroke-width="4"/>
  <path d="M20 112 H220 L262 76 H440" fill="none" stroke="#7a0a15" stroke-width="4"/>
  <circle cx="84" cy="64" r="38" fill="#f4f4f4" stroke="#7a0a15" stroke-width="4"/>
  <circle cx="84" cy="64" r="29" fill="url(#lente)"/>
  <circle class="brillo" cx="84" cy="64" r="29" fill="#b8e4ff"/>
  <circle cx="75" cy="54" r="8" fill="#ffffff" opacity="0.8"/>
  <circle class="led" cx="148" cy="42" r="8" fill="#ff5f57" stroke="#7a0a15" stroke-width="2"/>
  <circle class="led l2" cx="174" cy="42" r="8" fill="#ffcb05" stroke="#7a0a15" stroke-width="2"/>
  <circle class="led l3" cx="200" cy="42" r="8" fill="#5fd35f" stroke="#7a0a15" stroke-width="2"/>
  <path d="M60 128 H400 V318 H86 L60 292 Z" fill="#dcdcdc" stroke="#7a0a15" stroke-width="3"/>
  <circle cx="200" cy="137" r="3.5" fill="#e3223a"/><circle cx="260" cy="137" r="3.5" fill="#e3223a"/>
  <rect x="84" y="146" width="292" height="164" rx="6" fill="${fondo}"/>
  <g clip-path="url(#visor)">
    <g stroke="#1f4a33" stroke-width="1">
      ${Array.from({ length: 14 }, (_, i) => `<path d="M${84 + i * 22} 146 V310"/>`).join('')}
      ${Array.from({ length: 8 }, (_, i) => `<path d="M84 ${146 + i * 22} H376"/>`).join('')}
    </g>
    <g class="flota">
      <ellipse class="aura" cx="230" cy="232" rx="78" ry="70" fill="url(#aura)"/>
      ${pokemon(150, 168, 158, 124, 146)}
      <g class="silueta">${pokemon(150, 168, 158, 124, 146)}</g>
    </g>
    <g class="fin-escaneo"><rect class="escaneo" x="84" y="146" width="292" height="4" fill="#7fffb0" opacity="0.9"/></g>
    <text class="escaneando" x="94" y="166" font-family="${MONO}" font-size="12" font-weight="700" fill="#7fffb0">ESCANEANDO…</text>
    <text class="registrado" x="94" y="166" font-family="${MONO}" font-size="12" font-weight="700" fill="#ffcb05">Nº 150 · MEWTWO</text>
    <text class="registrado" x="366" y="302" font-family="${MONO}" font-size="11" fill="#7fd8a0" text-anchor="end">¡REGISTRADO!</text>
  </g>
  <circle cx="82" cy="372" r="20" fill="#2a2a2a" stroke="#111" stroke-width="3"/>
  <circle cx="76" cy="366" r="6" fill="#555"/>
  <rect x="130" y="340" width="48" height="10" rx="5" fill="#ff5f57" stroke="#7a0a15" stroke-width="2"/>
  <rect x="194" y="340" width="48" height="10" rx="5" fill="#3a6ee8" stroke="#1a2a6a" stroke-width="2"/>
  <rect x="130" y="362" width="112" height="36" rx="5" fill="#5fd35f" stroke="#2a6a2a" stroke-width="2"/>
  <text x="186" y="385" font-family="${MONO}" font-size="13" font-weight="700" fill="#1a4a1a" text-anchor="middle">PKS · 2026</text>
  <path d="M322 356 h20 v-20 h18 v20 h20 v18 h-20 v20 h-18 v-20 h-20 z" fill="#2a2a2a" stroke="#111" stroke-width="2"/>
  <circle cx="351" cy="365" r="5" fill="#444"/>

  <!-- bisagra -->
  <rect x="440" y="40" width="22" height="380" fill="#8e0c1a" stroke="#7a0a15" stroke-width="3"/>
  <path d="M440 70 H462 M440 390 H462" stroke="#5a0610" stroke-width="3"/>

  <!-- mitad derecha -->
  <path d="M462 70 H856 A24 24 0 0 1 880 94 V396 A24 24 0 0 1 856 420 H462 Z" fill="url(#cuerpo)" stroke="#7a0a15" stroke-width="4"/>
  <rect x="480" y="90" width="376" height="202" rx="10" fill="#dcdcdc" stroke="#7a0a15" stroke-width="3"/>
  <rect x="486" y="96" width="364" height="190" rx="6" fill="${fondo}"/>
  <g clip-path="url(#datos)">
${e.svg}  </g>
  <circle cx="490" cy="310" r="5" fill="#ffcb05" class="led"/><circle cx="508" cy="310" r="5" fill="#ffcb05" class="led l2"/>
  <text x="840" y="314" font-family="${SANS}" font-size="13" font-weight="700" fill="#ffd9dd" text-anchor="end">PokeScript · UNA 2026</text>
${barras}
</svg>
`;
  writeFileSync(join(OUT, 'pokedex.svg'), s);
}

// ─── 2. Combate: un error de tipos contado como un combate ─────────────────
{
  const W = 900;
  const H = 440;
  const fase1 = escritor('cmbA');
  const fase2 = escritor('cmbB');
  const blanco = '#fbfbf5';
  const linea = (e, y, texto, t, color = '#282828') =>
    e.linea(52, y, texto, { cw: 11, fs: 19, color, fondo: blanco, t, vel: 0.04 });

  let t = linea(fase1, 388, '¡ROCA usó + contra ELECTRICO!', 0.7);
  const ataque = t + 0.2;
  t = linea(fase1, 418, 'No es muy efectivo…', ataque + 1.0, '#c0392b');
  const borrar = t + 1.4;
  t = linea(fase2, 388, 'No se puede sumar un roca con un electrico.', borrar + 0.3);
  linea(fase2, 418, 'Revisa la tabla de efectividades.', t + 0.3, '#2a5caa');

  const vida = (x, y, nombre, extra) => `<g>
    <rect x="${x}" y="${y}" width="300" height="${extra ? 96 : 76}" rx="10" fill="${blanco}" stroke="#383838" stroke-width="4"/>
    <text x="${x + 20}" y="${y + 32}" font-family="${SANS}" font-size="20" font-weight="700" fill="#282828">${nombre}</text>
    <text x="${x + 280}" y="${y + 32}" font-family="${SANS}" font-size="17" font-weight="700" fill="#282828" text-anchor="end">Nv.5</text>
    <text x="${x + 20}" y="${y + 58}" font-family="${SANS}" font-size="13" font-weight="700" fill="#e08a00">PS</text>
    <rect x="${x + 46}" y="${y + 48}" width="232" height="12" rx="6" fill="#484848"/>
    <rect x="${x + 49}" y="${y + 51}" width="226" height="6" rx="3" fill="#48d048"/>
    ${extra ? `<text x="${x + 280}" y="${y + 86}" font-family="${SANS}" font-size="16" font-weight="700" fill="#282828" text-anchor="end">35/35</text>` : ''}
  </g>`;

  const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="Un error de tipos contado como un combate: ROCA usa + contra ELECTRICO y no es muy efectivo">
  <defs>
    <linearGradient id="campo" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0" stop-color="#c8e8f8"/><stop offset="0.55" stop-color="#e8f4d8"/><stop offset="1" stop-color="#a8d890"/>
    </linearGradient>
    <clipPath id="marco"><rect width="${W}" height="${H}" rx="16"/></clipPath>
    <clipPath id="dialogo"><rect x="36" y="362" width="${W - 72}" height="62"/></clipPath>
  </defs>
  <style>
    .entra-rival { animation: entraRival 0.6s ease-out both; }
    @keyframes entraRival { from { transform: translateX(260px); } }
    .entra-propio { animation: entraPropio 0.6s ease-out both; }
    @keyframes entraPropio { from { transform: translateX(-260px); } }
    .embiste { animation: embiste 0.5s ease-in-out ${ataque.toFixed(2)}s both; }
    @keyframes embiste { 50% { transform: translate(60px, -30px); } }
    .sacude { transform-box: fill-box; transform-origin: center; animation: sacude 0.5s ease-in-out ${(ataque + 0.25).toFixed(2)}s both; }
    @keyframes sacude { 20% { transform: translateX(-10px); } 40% { transform: translateX(10px); } 60% { transform: translateX(-6px); } 80% { transform: translateX(6px); } }
    .chispa { opacity: 0; animation: chispa 0.5s ease-out ${(ataque + 0.25).toFixed(2)}s both; }
    @keyframes chispa { 30% { opacity: 1; } 100% { opacity: 0; transform: scale(1.4); } }
    .limpiar { opacity: 0; animation: limpiar 0.01s ${borrar.toFixed(2)}s both; }
    @keyframes limpiar { to { opacity: 1; } }
    .flecha { animation: flecha 0.8s steps(1) infinite; }
    @keyframes flecha { 50% { opacity: 0; } }
${fase1.css}${fase2.css}  </style>
  <g clip-path="url(#marco)">
    <rect width="${W}" height="${H}" fill="url(#campo)"/>
    <ellipse cx="660" cy="176" rx="150" ry="30" fill="#b8d888" stroke="#88b060" stroke-width="3"/>
    <ellipse cx="230" cy="306" rx="170" ry="34" fill="#b8d888" stroke="#88b060" stroke-width="3"/>
    ${vida(40, 36, 'ELECTRICO', false)}
    ${vida(560, 212, 'ROCA', true)}
    <g class="entra-rival"><g class="sacude">${sprite('25', 596, 52, 128, 120)}</g></g>
    <g class="chispa" transform-origin="660 118"><circle cx="660" cy="118" r="70" fill="none" stroke="#ffffff" stroke-width="6" stroke-dasharray="10 14"/></g>
    <g class="entra-propio"><g class="embiste">${sprite('74-espalda', 120, 140, 220, 160)}</g></g>
    <rect x="20" y="350" width="${W - 40}" height="82" rx="12" fill="${blanco}" stroke="#383838" stroke-width="5"/>
    <rect x="28" y="358" width="${W - 56}" height="66" rx="8" fill="none" stroke="#c8c8c0" stroke-width="2"/>
    <g clip-path="url(#dialogo)">
${fase1.svg}      <g class="limpiar">
        <rect x="36" y="362" width="${W - 72}" height="62" fill="${blanco}"/>
${fase2.svg}      </g>
    </g>
    <path class="flecha" d="M${W - 58} 408 l10 0 l-5 8 z" fill="#c0392b"/>
  </g>
</svg>
`;
  writeFileSync(join(OUT, 'combate.svg'), s);
}

// ─── 3. Medallas: la hoja de ruta ──────────────────────────────────────────
// Los doce hitos son las ocho medallas de Kanto, dibujadas con la forma que
// tienen en Rojo Fuego, y los cuatro del Alto Mando con su Pokémon.
{
  const W = 900;
  const H = 380;
  const punto = (r, a) => `${(r * Math.cos(a)).toFixed(1)} ${(r * Math.sin(a)).toFixed(1)}`;
  const poligono = (lados, r, giro = -Math.PI / 2) => {
    let d = '';
    for (let i = 0; i < lados; i++)
      d += `${i ? 'L' : 'M'}${punto(r, (2 * Math.PI * i) / lados + giro)} `;
    return `${d}Z`;
  };
  const trazo = 'stroke="#1a1a1a" stroke-width="2.5" stroke-linejoin="round"';
  const brillo = (x, y, rx = 7, ry = 4) =>
    `<ellipse cx="${x}" cy="${y}" rx="${rx}" ry="${ry}" fill="#ffffff" opacity="0.75" transform="rotate(-35 ${x} ${y})"/>`;

  // Cada medalla va centrada en (0, 0) con radio ~34.
  const roca = () => {
    const g = Math.PI / 8;
    let facetas = '';
    for (let i = 0; i < 8; i++) {
      const a = (2 * Math.PI * i) / 8 + g;
      facetas += `<path d="M${punto(33, a)} L${punto(19, a)}" stroke="#55555e" stroke-width="1.5"/>`;
    }
    return `<path d="${poligono(8, 33, g)}" fill="#8e8e98" ${trazo}/>
      <path d="${poligono(8, 33, g)}" fill="none" stroke="#c4c4cc" stroke-width="3" transform="scale(0.86)"/>
      <path d="${poligono(8, 19, g)}" fill="#d6d6de" stroke="#55555e" stroke-width="1.5"/>${facetas}
      <path d="M-12 -4 L4 -16 L10 -12 L-6 0 Z" fill="#ffffff" opacity="0.8"/>`;
  };
  const cascada = () => `<g transform="rotate(-28)">
      <path d="M0 -38 C14 -18 26 -2 26 12 A26 26 0 0 1 -26 12 C-26 -2 -14 -18 0 -38 Z" fill="#35a4ea" ${trazo}/>
      <path d="M0 -26 C9 -12 18 0 18 12 A18 18 0 0 1 -18 12 C-18 0 -9 -12 0 -26 Z" fill="#7cd0ff"/>
      <path d="M8 26 A18 18 0 0 0 18 12" fill="none" stroke="#1a6aa8" stroke-width="4" stroke-linecap="round"/>
      ${brillo(-8, 4, 5, 9)}</g>`;
  const trueno = () => {
    let petalos = '';
    for (let i = 0; i < 8; i++) {
      petalos += `<path d="M0 -37 L11 -22 L0 -11 L-11 -22 Z" transform="rotate(${i * 45})" fill="${i % 2 ? '#ffb020' : '#ffd84a'}" ${trazo}/>`;
    }
    return `${petalos}<circle r="17" fill="#f07818" ${trazo}/><circle r="11" fill="#ff9a3a"/>${brillo(-5, -6, 5, 3)}`;
  };
  const arcoiris = () => {
    const colores = [
      '#f03a3a',
      '#ff9a1f',
      '#ffe23a',
      '#3ad05a',
      '#3ad0e0',
      '#3a7af0',
      '#9a4af0',
      '#f04ab0',
    ];
    const gemas = colores
      .map((c, i) => {
        const a = (2 * Math.PI * i) / 8 - Math.PI / 2;
        const [x, y] = punto(23, a).split(' ');
        return `<g transform="translate(${x} ${y})"><path d="${poligono(6, 12, 0)}" fill="${c}" ${trazo}/>${brillo(-3, -4, 3, 2)}</g>`;
      })
      .join('');
    const flor = [0, 90, 180, 270]
      .map(
        (g) =>
          `<circle cx="0" cy="-5" r="5" transform="rotate(${g + 45})" fill="#f2f2f2" stroke="#8a8a8a" stroke-width="1.5"/>`,
      )
      .join('');
    return `${gemas}<circle r="10" fill="#bdbdbd" ${trazo}/>${flor}<circle r="3" fill="#8a8a8a"/>`;
  };
  const corazon =
    'M0 32 C-10 24 -32 8 -32 -10 C-32 -24 -20 -32 -10 -30 C-4 -29 0 -24 0 -20 C0 -24 4 -29 10 -30 C20 -32 32 -24 32 -10 C32 8 10 24 0 32 Z';
  const alma = () =>
    `<path d="${corazon}" fill="#ff5ec8" ${trazo}/><path d="${corazon}" transform="scale(0.62) translate(0 2)" fill="#ffa4e4"/>${brillo(-16, -16, 6, 4)}`;
  const pantano = () =>
    `<circle r="33" fill="#f59a1a" ${trazo}/><circle r="24" fill="#ffc83a" stroke="#d86a08" stroke-width="4"/><circle r="15" fill="#ffe46a"/>${brillo(-14, -16, 7, 4)}`;
  const volcan = () =>
    `<path d="M-26 -2 L-31 -30 L-13 -18 L0 -37 L13 -18 L31 -30 L26 -2 C26 18 14 32 0 34 C-14 32 -26 18 -26 -2 Z" fill="#e8304a" ${trazo}/>
      <path d="M-18 0 L-21 -16 L-9 -9 L0 -22 L9 -9 L21 -16 L18 0 C18 14 10 24 0 25 C-10 24 -18 14 -18 0 Z" fill="#ff5a6e"/>
      <path d="M0 -6 L11 9 L0 24 L-11 9 Z" fill="#ffa8bc" stroke="#9a1028" stroke-width="2"/>${brillo(-3, 3, 3, 2)}`;
  const tierra = () => {
    const hojas = [
      [12, -26],
      [24, -16],
      [18, -4],
      [4, -14],
      [26, -30],
    ]
      .map(
        ([x, y], i) =>
          `<path d="${poligono(6, 8, 0)}" transform="translate(${x} ${y})" fill="${i % 2 ? '#2eaa3e' : '#5ad86a'}" ${trazo}/>`,
      )
      .join('');
    return `<path d="M-14 14 L14 -14" stroke="#1a1a1a" stroke-width="9" stroke-linecap="round"/>
      <path d="M-14 14 L14 -14" stroke="#58c048" stroke-width="5" stroke-linecap="round"/>
      ${hojas}<circle cx="-16" cy="17" r="14" fill="#c8f5a8" ${trazo}/><circle cx="-16" cy="17" r="8" fill="#e8ffd8"/>${brillo(-21, 12, 4, 3)}`;
  };

  const hitos = [
    ['Roca', 'Lexer', roca, true],
    ['Cascada', 'Expresiones', cascada, true],
    ['Trueno', 'Bloques', trueno, true],
    ['Arcoíris', 'Primer programa', arcoiris, true],
    ['Alma', 'Tipos', alma, true],
    ['Pantano', 'Movimientos', pantano, true],
    ['Volcán', 'Colecciones', volcan, true],
    ['Tierra', 'Segun', tierra, true],
    ['Lorelei', 'Importaciones', 131, true, '#58b8d8'],
    ['Bruno', 'Posible', 68, true, '#b04a30'],
    ['Agatha', 'Fichas', 94, true, '#6a4a9a'],
    ['Lance', 'Asistente', 149, true, '#5a5ad8'],
  ];
  const ganadas = hitos.filter((h) => h[3]).length;

  let cuerpo = '';
  hitos.forEach(([nombre, hito, dibujo, ganada, color], i) => {
    const enCaja = i < 8;
    const x = enCaja ? 102 + (i % 4) * 132 : 680 + ((i - 8) % 2) * 130;
    const y = 150 + Math.floor((enCaja ? i : i - 8) / (enCaja ? 4 : 2)) * 118;
    const d = (0.25 + i * 0.12).toFixed(2);
    const figura = enCaja
      ? `<g transform="translate(${x} ${y})">${dibujo()}</g>`
      : pokemon(dibujo, x - 32, y - 32, 64, 64);
    const fondo = enCaja
      ? `<circle cx="${x}" cy="${y}" r="42" fill="#6e1d1d"/>`
      : `<circle cx="${x}" cy="${y}" r="42" fill="${color}" stroke="#141630" stroke-width="3"/><circle cx="${x}" cy="${y}" r="36" fill="none" stroke="#ffffff" stroke-opacity="0.5" stroke-width="2"/>`;
    const pieza = ganada
      ? `<g class="pop" style="animation-delay:${d}s">${figura}</g>`
      : `<g class="silueta">${figura}</g><circle cx="${x}" cy="${y}" r="42" fill="none" stroke="#ffe9c8" stroke-opacity="0.6" stroke-width="2" stroke-dasharray="6 5"/>`;
    const claro = enCaja ? '#ffe9c8' : '#e6e9ff';
    const tenue = enCaja ? '#e6a88c' : '#8b90c0';
    cuerpo += `    ${fondo}
    ${pieza}
    <text x="${x}" y="${y + 60}" font-family="${SANS}" font-size="14" font-weight="700" fill="${claro}" text-anchor="middle">${nombre}</text>
    <text x="${x}" y="${y + 76}" font-family="${SANS}" font-size="11" fill="${tenue}" text-anchor="middle">${ganada ? `${i + 1} · ${hito}` : 'falta el editor'}</text>
${ganada ? `    <path class="destello" style="animation-delay:${(1.8 + i * 0.37).toFixed(2)}s" d="M${x + 26} ${y - 36} l3 8 l8 3 l-8 3 l-3 8 l-3 -8 l-8 -3 l8 -3 z" fill="#ffffff"/>\n` : ''}`;
  });

  const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="Estuche de medallas: ${ganadas} de 12. Las ocho medallas de Kanto y el Alto Mando, una por hito">
  <style>
    .pop { opacity: 0; transform-box: fill-box; transform-origin: center; animation: pop 0.45s cubic-bezier(.34,1.56,.64,1) both; }
    @keyframes pop { from { opacity: 0; transform: scale(0.3); } to { opacity: 1; transform: scale(1); } }
    .destello { opacity: 0; transform-box: fill-box; transform-origin: center; animation: destello 4.4s ease-in-out infinite; }
    @keyframes destello { 0%, 100% { opacity: 0; transform: scale(0.4); } 8% { opacity: 1; transform: scale(1.2); } 16% { opacity: 0; transform: scale(0.4); } }
    .silueta * { fill: #3a1010 !important; stroke: #4a1818 !important; opacity: 1 !important; }
  </style>
  <rect width="${W}" height="${H}" rx="16" fill="#1b1d36"/>
  <text x="36" y="44" font-family="${SANS}" font-size="22" font-weight="900" fill="#ffcb05">Estuche de medallas: ${ganadas} de 12</text>
  <text x="${W - 36}" y="44" font-family="${SANS}" font-size="14" fill="#8b90c0" text-anchor="end">un hito de la especificación por medalla</text>
  <rect x="24" y="70" width="562" height="290" rx="18" fill="#b8322a" stroke="#6e1d1d" stroke-width="4"/>
  <rect x="36" y="82" width="538" height="266" rx="12" fill="#8e2420"/>
  <text x="305" y="102" font-family="${SANS}" font-size="12" font-weight="700" letter-spacing="3" fill="#ffcfb8" text-anchor="middle">GIMNASIOS DE KANTO</text>
  <rect x="606" y="70" width="270" height="290" rx="18" fill="#2e3260" stroke="#141630" stroke-width="4"/>
  <text x="741" y="102" font-family="${SANS}" font-size="12" font-weight="700" letter-spacing="3" fill="#b8bdf0" text-anchor="middle">ALTO MANDO</text>
${cuerpo}</svg>
`;
  writeFileSync(join(OUT, 'medallas.svg'), s);
}

// ─── 4. Por dentro: el viaje de un programa por el compilador ──────────────
{
  const W = 900;
  const H = 280;
  const Y = 110;
  const estaciones = [
    ['principal.pks', 'tu código'],
    ['Lexer', 'palabras y símbolos'],
    ['Parser', 'el árbol'],
    ['Analizador', 'tipos y reglas'],
    ['Intérprete', 'lo ejecuta'],
    ['Salida', 'gritar · capturar'],
  ];
  const xs = estaciones.map((_, i) => 80 + i * 148);
  const total = 10; // segundos por vuelta
  const paso = 90 / estaciones.length; // % del ciclo por estación
  let bolaKf = '';
  let css = '';
  let nodos = '';
  estaciones.forEach(([titulo, sub], i) => {
    const llega = (i * paso).toFixed(1);
    const sale = (i * paso + paso * 0.55).toFixed(1);
    bolaKf += `${llega}%, ${sale}% { transform: translateX(${xs[i] - xs[0]}px); } `;
    const fin = Math.min(99, i * paso + paso * 0.9).toFixed(1);
    css += `    @keyframes luz${i} { 0%, ${Math.max(0, i * paso - 0.1).toFixed(1)}% { stroke: #3a3f7a; fill: #23264a; } ${llega}%, ${fin}% { stroke: #ffcb05; fill: #2d3160; } ${(Number(fin) + 0.1).toFixed(1)}%, 100% { stroke: #3a3f7a; fill: #23264a; } }\n`;
    nodos += `    <rect x="${xs[i] - 62}" y="${Y - 38}" width="124" height="76" rx="12" stroke-width="3" style="animation: luz${i} ${total}s linear infinite"/>
    <text x="${xs[i]}" y="${Y - 6}" font-family="${SANS}" font-size="16" font-weight="700" fill="#e6e9ff" text-anchor="middle">${esc(titulo)}</text>
    <text x="${xs[i]}" y="${Y + 18}" font-family="${SANS}" font-size="12" fill="#8b90c0" text-anchor="middle">${esc(sub)}</text>\n`;
  });
  const enAnalizador = (3 * paso).toFixed(1);
  const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="El recorrido de un programa: principal.pks pasa por el lexer, el parser, el analizador y el intérprete hasta la salida; el asistente ayuda cuando el analizador encuentra un error">
  <style>
    .viaje { animation: viaje ${total}s cubic-bezier(.5,0,.5,1) infinite; }
    @keyframes viaje { ${bolaKf}100% { transform: translateX(${xs[xs.length - 1] - xs[0]}px); } }
    .rueda { transform-box: fill-box; transform-origin: center; animation: rueda 1.2s linear infinite; }
    @keyframes rueda { to { transform: rotate(360deg); } }
    .asistente { animation: asistente ${total}s linear infinite; }
    @keyframes asistente { 0%, ${enAnalizador}% { opacity: 0.35; } ${(Number(enAnalizador) + 2).toFixed(1)}%, ${(Number(enAnalizador) + paso).toFixed(1)}% { opacity: 1; } ${(Number(enAnalizador) + paso + 2).toFixed(1)}%, 100% { opacity: 0.35; } }
${css}  </style>
  <rect width="${W}" height="${H}" rx="16" fill="#1b1d36"/>
  <line x1="${xs[0]}" y1="${Y + 52}" x2="${xs[xs.length - 1]}" y2="${Y + 52}" stroke="#3a3f7a" stroke-width="4" stroke-dasharray="3 9" stroke-linecap="round"/>
${nodos}  <g class="viaje"><g class="rueda">${pokebola(xs[0], Y + 52, 13, '#ee1515')}</g></g>
  <path d="M${xs[3]} ${Y + 66} V${Y + 104}" stroke="#ffcb05" stroke-width="2" stroke-dasharray="4 4"/>
  <g class="asistente">
    <rect x="${xs[3] - 150}" y="${Y + 106}" width="300" height="46" rx="10" fill="#23264a" stroke="#ffcb05" stroke-width="2"/>
    <text x="${xs[3]}" y="${Y + 124}" font-family="${SANS}" font-size="12" font-weight="700" fill="#ffcb05" text-anchor="middle">ASISTENTE</text>
    <text x="${xs[3]}" y="${Y + 143}" font-family="${MONO}" font-size="14" fill="#e6e9ff" text-anchor="middle">¿Quisiste decir «curar»?</text>
  </g>
</svg>
`;
  writeFileSync(join(OUT, 'recorrido.svg'), s);
}

// ─── 5. Guías: un Pokémon explica cada sección del README ──────────────────
// Cada escena tiene al Pokémon sobre su plataforma, con su propio gesto
// (Charizard escupe fuego, Snorlax ronca…), y un cuadro de diálogo como el
// de los juegos que se escribe letra por letra.
const GUIAS = [
  [
    'charizard',
    6,
    'Charizard',
    ['#ffc58a', '#e8603a'],
    ['fuego', 'espejo'],
    '¡Hola, entrenador! PokeScript es un lenguaje en español para aprender a programar. Los datos tienen tipo, las funciones son movimientos y los errores se explican en tu idioma.',
  ],
  [
    'rotom',
    479,
    'Rotom',
    ['#fff2a8', '#e0a820'],
    ['chispas', 'flota'],
    'Este es el IDE: el código, el asistente al lado y la salida abajo. Si escribes `curra` en vez de `curar`, el asistente te ofrece el arreglo.',
  ],
  [
    'pikachu',
    'pixel:25',
    'Pikachu',
    ['#fff2a8', '#e8b820'],
    ['chispas'],
    'Cada tipo de dato es un tipo de Pokémon. El mío es `electrico`: solo sé decir `verdadero` o `falso`, pero sin mí no hay `si` ni `mientras`.',
  ],
  [
    'mewtwo',
    150,
    'Mewtwo',
    ['#e2c8ff', '#6a3aa8'],
    ['aura', 'flota'],
    'Leo tu programa entero antes de que corra. Si dos tipos no se llevan, lo sé antes que tú, y te digo en qué línea y por qué.',
  ],
  [
    'charmander',
    4,
    'Charmander',
    ['#ffc58a', '#e8603a'],
    ['brasas'],
    'Todo programa empieza en `combate` y termina en su `fin`. Lo de adentro corre de arriba abajo, una instrucción por línea.',
  ],
  [
    'onix',
    95,
    'Onix',
    ['#e0d4a0', '#9a8448'],
    ['piedras'],
    'Cada dato se declara con su tipo. Una `medalla` es un valor que no cambia nunca, duro como la roca. Por eso va en mayúsculas.',
  ],
  [
    'psyduck',
    54,
    'Psyduck',
    ['#b0dcff', '#4a88d8'],
    ['duda'],
    'Aquí no hay «más o menos»: la condición de un `si` da `verdadero` o `falso`. `si vida` no compila; `si vida > 0` sí.',
  ],
  [
    'tauros',
    128,
    'Tauros',
    ['#e8dcc0', '#a08858'],
    ['embiste'],
    '`mientras` embiste una y otra vez hasta que la condición deja de cumplirse. `recorrer` pasa por un rango o por un equipo. `huir` sale del ciclo y `siguiente` salta a la otra vuelta.',
  ],
  [
    'machamp',
    68,
    'Machamp',
    ['#f4b8a0', '#c04a38'],
    ['golpes'],
    'Los movimientos son las funciones. Unos devuelven un valor con `entregar`; otros solo hacen algo, como gritar un mensaje. ¡Cuatro brazos, cero efectos colaterales!',
  ],
  [
    'kangaskhan',
    115,
    'Kangaskhan',
    ['#e8dcc0', '#a08858'],
    ['estrellas'],
    'Un `equipo` es una lista. Una `mochila` guarda cada cosa con su clave, como mi bolsa, y recuerda el orden. Se cuenta desde 1, como el primer Pokémon de tu equipo.',
  ],
  [
    'eevee',
    133,
    'Eevee',
    ['#f0dcc0', '#b08050'],
    ['estrellas'],
    'Una `especie` es una lista cerrada de valores, como mis evoluciones: no hay más que esas. Una `ficha` junta varios datos bajo un mismo nombre.',
  ],
  [
    'ditto',
    132,
    'Ditto',
    ['#ecd0f4', '#a070c0'],
    ['blandito'],
    'Cambiar de forma es lo mío, pero aquí se pide con `convertir`. Ojo: `convertir(agua) a roca` corta los decimales; para redondear está `redondear`.',
  ],
  [
    'abra',
    63,
    'Abra',
    ['#ffc0d8', '#d85890'],
    ['teleport'],
    'Con `enseñar … desde` un archivo se teletransporta lo que declara otro: movimientos, especies, fichas y medallas.',
  ],
  [
    'mew',
    151,
    'Mew',
    ['#ffd8ec', '#e070a8'],
    ['aura', 'flota'],
    'Cada animación de aquí abajo es un programa de verdad, con su salida de verdad. Están en la carpeta ejemplos y las pruebas los ejecutan.',
  ],
  [
    'machop',
    66,
    'Machop',
    ['#f4b8a0', '#c04a38'],
    ['golpes'],
    'Un equipo de niveles, una mochila de objetos y varios ciclos para entrenar. ¡A sudar!',
  ],
  [
    'gengar',
    94,
    'Gengar',
    ['#cbb4ec', '#6a4a9a'],
    ['flota', 'teleport'],
    'Una `especie` con los estados alterados, un `segun` que los cubre todos y movimientos que dicen qué le pasa a cada Pokémon. Je, je.',
  ],
  [
    'chansey',
    113,
    'Chansey',
    ['#ffd4e2', '#e0789a'],
    ['corazones'],
    '`capturar` se queda esperando lo que escribas. Si la respuesta no sirve, lo explica y vuelve a preguntar, con la paciencia de un Centro Pokémon.',
  ],
  [
    'magikarp',
    129,
    'Magikarp',
    ['#b0dcff', '#4a88d8'],
    ['salta'],
    'Olvidar un `fin` es el error más común al empezar. El mensaje dice qué bloque quedó abierto, en qué línea empezó y cuál es el que falta cerrar.',
  ],
  [
    'blastoise',
    9,
    'Blastoise',
    ['#b0dcff', '#3a70c0'],
    ['agua'],
    'El programa de la especificación: cuatro archivos que se importan entre sí. Tu Pokémon contra Bulbi, veinte turnos como máximo.',
  ],
  [
    'porygon',
    137,
    'Porygon',
    ['#b4ecf4', '#3a98b8'],
    ['pixeles', 'flota'],
    'Un programa pasa por cuatro etapas antes de mostrar algo. Si el analizador no encuentra un nombre, el asistente busca qué quisiste escribir.',
  ],
  [
    'dragonite',
    149,
    'Dragonite',
    ['#ffdca0', '#d89030'],
    ['flota', 'estrellas'],
    'Cada hito es una medalla: primero los ocho gimnasios de Kanto y después el Alto Mando. ¡Las doce están ganadas, incluido el IDE con sus colores!',
  ],
  [
    'snorlax',
    143,
    'Snorlax',
    ['#c8dce4', '#58788a'],
    ['zzz'],
    'Para despertarme hace falta Go 1.25 o más nuevo. Para el IDE, también Node.js 22, pnpm 10 y la herramienta de Wails. Con npm no me muevo.',
  ],
  [
    'pidgeot',
    18,
    'Pidgeot',
    ['#d8ecff', '#6a98c8'],
    ['flota', 'plumas'],
    'Llevo tus cambios a `dev`: commits en inglés con el formato PKS y siempre por pull request. Nada vuela directo a `main`.',
  ],
];

// envolver parte el texto en líneas de hasta max letras sin cortar el código.
function envolver(texto, max) {
  const lineas = [''];
  for (const pal of texto.match(/(?:[^\s`]*`[^`]*`)+[^\s`]*|\S+/g)) {
    const actual = lineas[lineas.length - 1];
    const junto = actual ? `${actual} ${pal}` : pal;
    if (largo(junto.replace(/`/g, '')) > max && actual) lineas.push(pal);
    else lineas[lineas.length - 1] = junto;
  }
  return lineas;
}

// Gestos: el movimiento del Pokémon (clase de su grupo) y lo que aparece a
// su alrededor. El Pokémon ocupa la caja (28, 14) – (208, 190).
const rayo = (x, y, d) =>
  `<path class="chispa" style="animation-delay:${d}s" d="M${x} ${y - 13} l-6 13 h6 l-4 13 l12 -16 h-6 l5 -10 z" fill="#ffe14a" stroke="#a87800" stroke-width="1.5"/>`;
const destello = (x, y, d, c = '#ffffff') =>
  `<path class="chispa" style="animation-delay:${d}s" d="M${x} ${y - 9} l2.5 6.5 l6.5 2.5 l-6.5 2.5 l-2.5 6.5 l-2.5 -6.5 l-6.5 -2.5 l6.5 -2.5 z" fill="${c}"/>`;
const sube = (clase, x, y, d, contenido) =>
  `<g class="${clase}" style="animation-delay:${d}s">${contenido(x, y)}</g>`;
const GESTOS = {
  fuego: {
    dentro: [0, 0.08, 0.16, 0.24, 0.32, 0.4, 0.48, 0.56, 0.64, 0.72]
      .map(
        (d, i) =>
          `<circle class="llama" style="animation-delay:${d}s" cx="146" cy="60" r="${8 + (i % 3) * 3}" fill="${['#fff27a', '#ffb020', '#f0461e'][i % 3]}"/>`,
      )
      .join(''),
  },
  brasas: {
    fuera: [60, 90, 150, 180, 120]
      .map(
        (x, i) =>
          `<circle class="burbuja" style="animation-delay:${(i * 0.8).toFixed(1)}s" cx="${x}" cy="180" r="${2 + (i % 3)}" fill="${i % 2 ? '#ffb020' : '#ff5a1a'}"/>`,
      )
      .join(''),
  },
  chispas: { fuera: rayo(40, 50, 0) + rayo(200, 40, 0.5) + rayo(28, 150, 1) + rayo(210, 140, 1.5) },
  estrellas: {
    fuera:
      destello(40, 40, 0) +
      destello(205, 60, 0.7) +
      destello(50, 160, 1.4) +
      destello(195, 150, 2.1),
  },
  aura: {
    detras: '<ellipse class="aura" cx="118" cy="104" rx="100" ry="92" fill="url(#brillo)"/>',
  },
  flota: { clase: 'flota' },
  // espejo voltea al Pokémon para que mire hacia el cuadro de diálogo.
  espejo: { espejo: true },
  blandito: { clase: 'blandito' },
  teleport: { clase: 'teleport' },
  salta: {
    clase: 'salta',
    fuera: [80, 100, 136, 156]
      .map(
        (x, i) =>
          `<circle class="salpica" style="animation-delay:${(i * 0.05).toFixed(2)}s;--dx:${(x - 118) * 0.8}px" cx="${x}" cy="186" r="4" fill="#6ac0ff"/>`,
      )
      .join(''),
  },
  embiste: {
    clase: 'embiste',
    fuera: [40, 60, 30]
      .map(
        (x, i) =>
          `<circle class="polvo" style="animation-delay:${(i * 0.08).toFixed(2)}s" cx="${x}" cy="${182 - i * 6}" r="${9 - i * 2}" fill="#d8c8a0"/>`,
      )
      .join(''),
  },
  golpes: {
    fuera: [
      [30, 70, 0],
      [206, 90, 0.9],
    ]
      .map(
        ([x, y, d]) =>
          `<g class="golpe" style="animation-delay:${d}s"><path d="M${x} ${y - 16} l5 10 l11 -3 l-6 10 l9 7 l-11 1 l1 11 l-9 -7 l-9 7 l1 -11 l-11 -1 l9 -7 l-6 -10 l11 3 z" fill="#ffe14a" stroke="#c04a38" stroke-width="2"/></g>`,
      )
      .join(''),
  },
  zzz: {
    fuera: [0, 1.1, 2.2]
      .map((d, i) =>
        sube(
          'zeta',
          176 + i * 6,
          60 - i * 4,
          d,
          (x, y) =>
            `<text x="${x}" y="${y}" font-family="${SANS}" font-size="${18 + i * 5}" font-weight="900" fill="#2d2d3a">Z</text>`,
        ),
      )
      .join(''),
  },
  duda: {
    fuera:
      `<text class="duda" x="190" y="46" font-family="${SANS}" font-size="34" font-weight="900" fill="#2d2d3a">?</text>` +
      `<text class="duda" style="animation-delay:0.6s" x="40" y="56" font-family="${SANS}" font-size="24" font-weight="900" fill="#2d2d3a">?</text>`,
  },
  agua: {
    fuera: [
      [72, 34, -1],
      [170, 44, 1],
    ]
      .map(([x, y, lado]) =>
        [0, 0.15, 0.3, 0.45]
          .map(
            (d) =>
              `<circle class="chorro" style="animation-delay:${d}s;--dx:${lado * 40}px" cx="${x}" cy="${y}" r="5" fill="#8fd8ff" stroke="#2a78c8" stroke-width="1.5"/>`,
          )
          .join(''),
      )
      .join(''),
  },
  pixeles: {
    fuera: [
      [36, 50, '#ff5a8a'],
      [200, 44, '#4ad8f0'],
      [40, 150, '#4ad8f0'],
      [202, 140, '#ff5a8a'],
      [120, 20, '#ffe14a'],
    ]
      .map(
        ([x, y, c], i) =>
          `<rect class="chispa" style="animation-delay:${(i * 0.4).toFixed(1)}s" x="${x}" y="${y}" width="10" height="10" fill="${c}"/>`,
      )
      .join(''),
  },
  corazones: {
    fuera: [70, 120, 170]
      .map((x, i) =>
        sube(
          'burbuja',
          x,
          170,
          i * 1.2,
          (cx, cy) =>
            `<path d="M${cx} ${cy + 8} c-10 -7 -10 -16 -4 -17 c3 0 4 2 4 3 c0 -1 1 -3 4 -3 c6 1 6 10 -4 17 z" fill="#ff7ab0"/>`,
        ),
      )
      .join(''),
  },
  piedras: {
    fuera: [50, 170, 110]
      .map(
        (x, i) =>
          `<path class="piedra" style="animation-delay:${(i * 0.7).toFixed(1)}s" d="M${x} 10 l8 -3 l6 5 l-2 8 l-9 2 l-5 -6 z" fill="#b8a878" stroke="#6a5a38" stroke-width="1.5"/>`,
      )
      .join(''),
  },
  plumas: {
    fuera: [60, 180]
      .map(
        (x, i) =>
          `<path class="pluma" style="animation-delay:${(i * 1.6).toFixed(1)}s" d="M${x} 10 q10 8 0 22 q-10 -8 0 -22 z" fill="#f5e0a8" stroke="#a88848" stroke-width="1.2"/>`,
      )
      .join(''),
  },
};
const CSS_GESTOS = `    .respira { animation: respira 2.6s ease-in-out 0.7s infinite; }
    @keyframes respira { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-7px); } }
    .flota { animation: flota 3.4s ease-in-out 0.7s infinite; }
    @keyframes flota { 0%, 100% { transform: translateY(-4px); } 50% { transform: translateY(-18px); } }
    .blandito { transform-origin: 118px 190px; animation: blandito 1.8s ease-in-out 0.7s infinite; }
    @keyframes blandito { 0%, 100% { transform: scale(1, 1); } 30% { transform: scale(1.1, 0.88); } 60% { transform: scale(0.92, 1.08); } }
    .teleport { animation: teleport 4.5s steps(1) 1s infinite; }
    @keyframes teleport { 0%, 78%, 100% { opacity: 1; transform: none; } 80% { opacity: 0.2; } 82% { opacity: 0; transform: translateX(14px); } 86% { opacity: 0.5; } 88% { opacity: 1; transform: none; } }
    .salta { animation: salta 1.6s ease-in-out 0.7s infinite; }
    @keyframes salta { 0%, 100% { transform: translateY(0) rotate(0); } 40% { transform: translateY(-34px) rotate(-8deg); } 55% { transform: translateY(-34px) rotate(8deg); } 80% { transform: translateY(0) rotate(0); } }
    .embiste { animation: embiste 2.4s ease-in 0.7s infinite; }
    @keyframes embiste { 0%, 60%, 100% { transform: none; } 70% { transform: translateX(-10px); } 78% { transform: translateX(22px); } 88% { transform: none; } }
    .llama { opacity: 0; transform-box: fill-box; transform-origin: center; animation: llama 2.6s ease-out 1s infinite; }
    @keyframes llama { 0% { opacity: 0; transform: translate(0, 0) scale(0.4); } 5% { opacity: 1; } 30% { opacity: 0.9; } 42% { opacity: 0; transform: translate(84px, 8px) scale(2.6); } 100% { opacity: 0; } }
    .chispa { opacity: 0; transform-box: fill-box; transform-origin: center; animation: chispa 2s ease-in-out infinite; }
    @keyframes chispa { 0%, 100% { opacity: 0; transform: scale(0.5); } 10% { opacity: 1; transform: scale(1.15); } 25% { opacity: 0; transform: scale(0.5); } }
    .aura { transform-box: fill-box; transform-origin: center; animation: aura 2.4s ease-in-out infinite; }
    @keyframes aura { 0%, 100% { opacity: 0.45; transform: scale(0.9); } 50% { opacity: 1; transform: scale(1.06); } }
    .burbuja { opacity: 0; animation: burbuja 5.4s ease-out infinite; }
    @keyframes burbuja { 0% { opacity: 0; transform: translateY(0); } 15% { opacity: 0.8; } 100% { opacity: 0; transform: translateY(-170px); } }
    .salpica { opacity: 0; animation: salpica 1.6s ease-out 0.7s infinite; }
    @keyframes salpica { 0%, 75% { opacity: 0; transform: none; } 80% { opacity: 1; } 100% { opacity: 0; transform: translate(var(--dx), -26px); } }
    .polvo { opacity: 0; transform-box: fill-box; transform-origin: center; animation: polvo 2.4s ease-out 0.7s infinite; }
    @keyframes polvo { 0%, 76% { opacity: 0; transform: scale(0.4); } 80% { opacity: 0.8; } 100% { opacity: 0; transform: translateX(-30px) scale(1.6); } }
    .golpe { opacity: 0; transform-box: fill-box; transform-origin: center; animation: golpe 1.8s ease-out infinite; }
    @keyframes golpe { 0%, 100% { opacity: 0; transform: scale(0.3); } 8% { opacity: 1; transform: scale(1.2); } 22% { opacity: 0; transform: scale(0.9); } }
    .zeta { opacity: 0; animation: zeta 3.3s ease-out infinite; }
    @keyframes zeta { 0% { opacity: 0; transform: translate(0, 0); } 20% { opacity: 0.9; } 100% { opacity: 0; transform: translate(26px, -40px); } }
    .duda { transform-box: fill-box; transform-origin: bottom; animation: duda 1.4s ease-in-out infinite; }
    @keyframes duda { 0%, 100% { transform: rotate(-10deg); } 50% { transform: rotate(10deg) translateY(-4px); } }
    .chorro { opacity: 0; animation: chorro 1.2s ease-out 0.7s infinite; }
    @keyframes chorro { 0% { opacity: 1; transform: none; } 100% { opacity: 0; transform: translate(var(--dx), -40px) scale(1.6); } }
    .piedra { opacity: 0; animation: piedra 2.2s ease-in infinite; }
    @keyframes piedra { 0% { opacity: 0; transform: translateY(0) rotate(0); } 10% { opacity: 1; } 80% { opacity: 1; transform: translateY(170px) rotate(90deg); } 100% { opacity: 0; transform: translateY(170px) rotate(90deg); } }
    .pluma { opacity: 0; animation: pluma 3.2s ease-in-out infinite; }
    @keyframes pluma { 0% { opacity: 0; transform: translate(0, 0) rotate(0); } 15% { opacity: 1; } 50% { transform: translate(14px, 80px) rotate(30deg); } 100% { opacity: 0; transform: translate(-6px, 170px) rotate(-20deg); } }
`;

{
  const W = 900;
  const H = 220;
  mkdirSync(join(OUT, 'guias'), { recursive: true });
  for (const [archivo, id, nombre, [claro, oscuro], gestos, texto] of GUIAS) {
    const lineas = envolver(texto, 58);
    if (lineas.length > 4) throw new Error(`${archivo}: el texto no cabe en el cuadro`);
    const e = escritor('l');
    let t = 0.8;
    lineas.forEach((l, i) => {
      t = e.linea(262, 82 + i * 30, l, { color: '#2d2d3a', fondo: '#ffffff', t, vel: 0.03 });
    });
    const g = gestos.map((n) => GESTOS[n]);
    // Algunos se ven mejor con su sprite de los juegos que con el dream world.
    const dibujo = String(id).startsWith('pixel:')
      ? sprite(id.slice(6), 38, 24, 160, 160)
      : pokemon(id, 28, 14, 180, 176);
    const clase = g.find((x) => x.clase)?.clase ?? 'respira';
    const parte = (k) => g.map((x) => x[k] ?? '').join('');
    const burbujas = [30, 70, 120, 160, 190, 55]
      .map(
        (x, i) =>
          `<circle class="burbuja" cx="${x}" cy="${200 - (i % 3) * 12}" r="${3 + (i % 3)}" fill="#ffffff" style="animation-delay:${(i * 0.9).toFixed(1)}s"/>`,
      )
      .join('\n    ');
    const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="${nombre}: ${esc(texto.replace(/`/g, ''))}">
  <defs>
    <linearGradient id="cielo" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${claro}"/><stop offset="1" stop-color="${oscuro}"/></linearGradient>
    <radialGradient id="brillo" cx="0.5" cy="0.5" r="0.5"><stop offset="0" stop-color="#ffffff" stop-opacity="0.9"/><stop offset="0.5" stop-color="${oscuro}" stop-opacity="0.45"/><stop offset="1" stop-color="${oscuro}" stop-opacity="0"/></radialGradient>
    <clipPath id="dialogo"><rect x="241" y="42" width="628" height="148"/></clipPath>
  </defs>
  <style>
    .entra { animation: entra 0.7s cubic-bezier(.34,1.56,.64,1) both; }
    @keyframes entra { from { opacity: 0; transform: translateX(-70px); } to { opacity: 1; transform: none; } }
    .sombra { transform-box: fill-box; transform-origin: center; animation: sombra 2.6s ease-in-out 0.7s infinite; }
    @keyframes sombra { 0%, 100% { transform: scale(1); } 50% { transform: scale(0.88); } }
    .flecha { opacity: 0; animation: flecha 0.9s steps(1) ${t.toFixed(2)}s infinite; }
    @keyframes flecha { 0% { opacity: 1; } 50% { opacity: 0; } }
${CSS_GESTOS}${e.css}  </style>
  <rect width="${W}" height="${H}" rx="16" fill="url(#cielo)"/>
  <g opacity="0.5">
    ${burbujas}
  </g>
  <ellipse cx="118" cy="192" rx="92" ry="18" fill="${oscuro}" opacity="0.55"/>
  <ellipse class="sombra" cx="118" cy="190" rx="62" ry="9" fill="#000000" opacity="0.18"/>
  ${parte('detras')}
  <g class="entra"><g class="${clase}">
    ${g.some((x) => x.espejo) ? `<g transform="translate(236 0) scale(-1 1)">${dibujo}</g>` : dibujo}
    ${parte('dentro')}
  </g></g>
  ${parte('fuera')}
  <rect x="235" y="36" width="640" height="160" rx="12" fill="#ffffff" stroke="#2d2d3a" stroke-width="5"/>
  <rect x="242" y="43" width="626" height="146" rx="8" fill="none" stroke="#c8ccd8" stroke-width="2"/>
  <rect x="256" y="18" width="${largo(nombre) * 13 + 30}" height="32" rx="8" fill="#2d2d3a"/>
  <text x="${256 + 17}" y="40" font-family="${SANS}" font-size="16" font-weight="900" letter-spacing="1" fill="#ffffff">${nombre.toUpperCase()}</text>
  <g clip-path="url(#dialogo)">
${e.svg}  </g>
  <path class="flecha" d="M842 172 h18 l-9 10 z" fill="#e8603a"/>
</svg>
`;
    writeFileSync(join(OUT, 'guias', `${archivo}.svg`), s);
  }
}

// ─── 6. Índice: el equipo del README ───────────────────────────────────────
// Cada sección es un Pokémon del equipo, como en el menú de los juegos: una
// ficha por archivo para que cada una sea un enlace aparte en el README.
{
  const INDICE = [
    [
      '01-editor',
      'El editor',
      'Código, asistente y salida',
      479,
      'Rotom',
      'FANTASMA',
      ['#b89cf0', '#5a3a9a'],
    ],
    [
      '02-tipos',
      'Tipos',
      'Cada dato es un tipo de Pokémon',
      25,
      'Pikachu',
      'ELÉCTRICO',
      ['#ffe066', '#d89a10'],
    ],
    [
      '03-basico',
      'Lo básico',
      'Datos, condiciones y ciclos',
      4,
      'Charmander',
      'FUEGO',
      ['#ffb070', '#d8502a'],
    ],
    [
      '04-programas',
      'Programas completos',
      'Ejemplos que corren de verdad',
      151,
      'Mew',
      'PSÍQUICO',
      ['#ffa0c8', '#c0407e'],
    ],
    [
      '05-por-dentro',
      'Por dentro',
      'Lexer, parser, análisis e intérprete',
      137,
      'Porygon',
      'NORMAL',
      ['#8ad8e8', '#2a88a8'],
    ],
    [
      '06-medallas',
      'Medallas',
      'La hoja de ruta, 12 de 12',
      149,
      'Dragonite',
      'DRAGÓN',
      ['#a898ff', '#4a3ab0'],
    ],
    [
      '07-probarlo',
      'Probarlo',
      'Instalar, compilar y correr',
      143,
      'Snorlax',
      'NORMAL',
      ['#d8cca0', '#7a6e48'],
    ],
    [
      '08-contribuir',
      'Contribuir',
      'Commits, ramas y pull requests',
      18,
      'Pidgeot',
      'VOLADOR',
      ['#9cd0ff', '#3a78c8'],
    ],
  ];
  const W = 440;
  const H = 110;
  mkdirSync(join(OUT, 'indice'), { recursive: true });
  INDICE.forEach(([archivo, titulo, desc, id, quien, tipo, [claro, oscuro]], i) => {
    const n = String(i + 1).padStart(2, '0');
    const anchoTipo = largo(tipo) * 8.4 + 20;
    const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="${n}. ${titulo}: ${desc}">
  <defs>
    <linearGradient id="ficha" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${claro}"/><stop offset="1" stop-color="${oscuro}"/></linearGradient>
    <radialGradient id="foco" cx="0.5" cy="0.55" r="0.5"><stop offset="0" stop-color="#ffffff" stop-opacity="0.75"/><stop offset="1" stop-color="#ffffff" stop-opacity="0"/></radialGradient>
    <clipPath id="borde"><rect x="6" y="6" width="${W - 12}" height="${H - 12}" rx="20"/></clipPath>
  </defs>
  <style>
    .salta { animation: salta 2.4s ease-in-out ${(i * 0.3).toFixed(1)}s infinite; }
    @keyframes salta { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-4px); } }
    .flecha { animation: flecha 1.2s ease-in-out infinite; }
    @keyframes flecha { 0%, 100% { transform: translateX(0); } 50% { transform: translateX(4px); } }
    .brilla { animation: brilla 5s ease-in-out ${(i * 0.6).toFixed(1)}s infinite; }
    @keyframes brilla { 0%, 70% { transform: translateX(-160px); } 100% { transform: translateX(${W + 160}px); } }
  </style>
  <rect x="2" y="2" width="${W - 4}" height="${H - 4}" rx="24" fill="#1c2340"/>
  <rect x="6" y="6" width="${W - 12}" height="${H - 12}" rx="20" fill="url(#ficha)"/>
  <g clip-path="url(#borde)">
    <rect x="6" y="6" width="${W - 12}" height="${(H - 12) / 2}" fill="#ffffff" opacity="0.14"/>
    <circle cx="${W - 58}" cy="${H - 6}" r="70" fill="none" stroke="#ffffff" stroke-width="14" opacity="0.1"/>
    <path d="M${W - 128} ${H - 6} h140" stroke="#ffffff" stroke-width="14" opacity="0.1"/>
    <circle cx="${W - 58}" cy="${H - 6}" r="22" fill="none" stroke="#ffffff" stroke-width="10" opacity="0.1"/>
    <path class="brilla" d="M0 0 h26 l-40 ${H} h-26 z" fill="#ffffff" opacity="0.18"/>
  </g>
  <circle cx="62" cy="58" r="40" fill="url(#foco)"/>
  <ellipse cx="62" cy="88" rx="30" ry="6" fill="#000000" opacity="0.22"/>
  <g class="salta">${pokemon(id, 22, 16, 80, 74)}</g>
  <rect x="116" y="18" width="46" height="20" rx="10" fill="#1c2340"/>
  <text x="139" y="32.5" text-anchor="middle" font-family="${MONO}" font-size="12" font-weight="700" fill="#ffcb05">Nº${n}</text>
  <rect x="168" y="18" width="${anchoTipo}" height="20" rx="10" fill="#ffffff" opacity="0.92"/>
  <text x="${168 + anchoTipo / 2}" y="32.5" text-anchor="middle" font-family="${SANS}" font-size="11" font-weight="900" letter-spacing="1" fill="${oscuro}">${tipo}</text>
  <text x="116" y="66" font-family="${SANS}" font-size="${largo(titulo) > 14 ? 22 : 25}" font-weight="900" fill="#ffffff" stroke="#1c2340" stroke-width="5" stroke-linejoin="round" paint-order="stroke">${titulo}</text>
  <text x="117" y="88" font-family="${SANS}" font-size="13.5" font-weight="700" fill="#1c2340" opacity="0.85">${esc(desc)}</text>
  <text x="${W - 30}" y="32" text-anchor="end" font-family="${SANS}" font-size="11" font-weight="700" fill="#1c2340" opacity="0.7">${quien.toUpperCase()}</text>
  <g class="flecha"><circle cx="${W - 38}" cy="62" r="15" fill="#1c2340"/><path d="M${W - 42} 55 l8 7 l-8 7" fill="none" stroke="#ffcb05" stroke-width="3.5" stroke-linecap="round" stroke-linejoin="round"/></g>
</svg>
`;
    writeFileSync(join(OUT, 'indice', `${archivo}.svg`), s);
  });
}

console.log('escenas listas');

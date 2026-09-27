// Genera las escenas animadas del README: la Pokédex, el combate de tipos,
// las medallas de la hoja de ruta, el recorrido de un programa por el
// compilador y los Pokémon que explican cada sección.
// Uso: pnpm docs:svg (escribe en docs/assets/).
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const OUT = process.argv[2] ?? fileURLToPath(new URL('../docs/assets/', import.meta.url));
const MONO = "Consolas, 'Courier New', monospace";
const SANS = "'Trebuchet MS', 'Segoe UI', Verdana, sans-serif";

// El arte de PokeAPI se incrusta en base64: GitHub no deja que un SVG cargue
// imágenes externas. Se descarga con scripts/descargar-sprites.mjs.
const FUENTES = fileURLToPath(new URL('../docs/assets/fuentes/', import.meta.url));
const incrustar = (archivo, tipo) =>
  `data:${tipo};base64,${readFileSync(join(FUENTES, archivo)).toString('base64')}`;
const pokemon = (id) => incrustar(`pokemon/${id}.svg`, 'image/svg+xml');
const medalla = (n) => incrustar(`medallas/${n}.png`, 'image/png');

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
{
  const W = 900;
  const H = 400;
  const e = escritor('pdx');
  const fondo = '#10231a';
  let t = 0.8;
  const lineas = [
    ['Nº 001 · POKESCRIPT', '#ffcb05', 700],
    ['TIPO: LENGUAJE / EDUCATIVO', '#7fd8a0', 700],
    ['', '', 400],
    ['Un lenguaje en español para aprender a programar.', '#d8f5e0', 400],
    ['Cada dato tiene un tipo y, como en los combates,', '#d8f5e0', 400],
    ['no todos los tipos se llevan bien.', '#d8f5e0', 400],
    ['Cuando algo falla, te dice qué pasó y cómo arreglarlo.', '#d8f5e0', 400],
  ];
  lineas.forEach(([texto, color, peso], i) => {
    if (texto) t = e.linea(104, 184 + i * 24, texto, { color, fondo, t, peso, vel: 0.028 }) + 0.2;
  });
  const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="Pokédex con la entrada de PokeScript: un lenguaje en español para aprender a programar">
  <defs>
    <clipPath id="pantalla"><rect x="84" y="158" width="${W - 168}" height="186" rx="6"/></clipPath>
    <radialGradient id="lente" cx="0.35" cy="0.35" r="0.7">
      <stop offset="0" stop-color="#b8e4ff"/><stop offset="0.45" stop-color="#3d9cf0"/><stop offset="1" stop-color="#15508f"/>
    </radialGradient>
    <linearGradient id="cuerpo" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0" stop-color="#e3223a"/><stop offset="1" stop-color="#b3101f"/>
    </linearGradient>
  </defs>
  <style>
    .brillo { animation: brillo 2.4s ease-in-out infinite; }
    @keyframes brillo { 0%, 100% { opacity: 0.25; } 50% { opacity: 0.8; } }
    .led { animation: led 1.8s steps(1) infinite; }
    .l2 { animation-delay: 0.6s; } .l3 { animation-delay: 1.2s; }
    @keyframes led { 0% { opacity: 1; } 33% { opacity: 0.25; } }
    .barrido { animation: barrido 3.5s linear infinite; }
    @keyframes barrido { from { transform: translateY(0); } to { transform: translateY(190px); } }
${e.css}  </style>
  <rect x="20" y="20" width="${W - 40}" height="${H - 40}" rx="26" fill="url(#cuerpo)" stroke="#7a0a15" stroke-width="4"/>
  <path d="M20 118 H300 L340 78 H${W - 20}" fill="none" stroke="#7a0a15" stroke-width="4"/>
  <circle cx="92" cy="72" r="40" fill="#f4f4f4" stroke="#7a0a15" stroke-width="4"/>
  <circle cx="92" cy="72" r="31" fill="url(#lente)"/>
  <circle class="brillo" cx="92" cy="72" r="31" fill="#b8e4ff"/>
  <circle cx="82" cy="61" r="8" fill="#ffffff" opacity="0.8"/>
  <circle class="led" cx="160" cy="52" r="9" fill="#ff5f57" stroke="#7a0a15" stroke-width="2"/>
  <circle class="led l2" cx="188" cy="52" r="9" fill="#ffcb05" stroke="#7a0a15" stroke-width="2"/>
  <circle class="led l3" cx="216" cy="52" r="9" fill="#5fd35f" stroke="#7a0a15" stroke-width="2"/>
  <text x="${W - 50}" y="80" font-family="${SANS}" font-size="15" font-weight="700" fill="#ffd9dd" text-anchor="end">PokeScript · UNA 2026</text>
  <rect x="60" y="140" width="${W - 120}" height="218" rx="12" fill="#dcdcdc" stroke="#7a0a15" stroke-width="3"/>
  <circle cx="78" cy="152" r="4" fill="#e3223a"/><circle cx="92" cy="152" r="4" fill="#e3223a"/>
  <rect x="84" y="158" width="${W - 168}" height="186" rx="6" fill="${fondo}"/>
  <g clip-path="url(#pantalla)">
${e.svg}  </g>
  <rect class="barrido" x="84" y="158" width="${W - 168}" height="3" fill="#7fd8a0" opacity="0.18"/>
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
    <g class="entra-rival"><g class="sacude">${pokebola(660, 118, 54, '#f8d030')}</g></g>
    <g class="chispa" transform-origin="660 118"><circle cx="660" cy="118" r="70" fill="none" stroke="#ffffff" stroke-width="6" stroke-dasharray="10 14"/></g>
    <g class="entra-propio"><g class="embiste">${pokebola(230, 244, 62, '#b8a038')}</g></g>
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
// Los doce hitos son las ocho medallas de Kanto, con su dibujo del juego, y
// los cuatro del Alto Mando con su Pokémon más conocido.
{
  const W = 900;
  const H = 380;
  const hitos = [
    ['Roca', 'Lexer', medalla(1), false],
    ['Cascada', 'Expresiones', medalla(2), true],
    ['Trueno', 'Bloques', medalla(3), true],
    ['Arcoíris', 'Primer programa', medalla(4), true],
    ['Alma', 'Tipos', medalla(5), true],
    ['Pantano', 'Movimientos', medalla(6), true],
    ['Volcán', 'Colecciones', medalla(7), true],
    ['Tierra', 'Segun', medalla(8), true],
    ['Lorelei', 'Importaciones', pokemon(131), true, '#58b8d8'],
    ['Bruno', 'Posible', pokemon(68), true, '#b04a30'],
    ['Agatha', 'Fichas', pokemon(94), true, '#6a4a9a'],
    ['Lance', 'Asistente', pokemon(149), true, '#5a5ad8'],
  ];
  const ganadas = hitos.filter((h) => h[3]).length;

  let cuerpo = '';
  hitos.forEach(([nombre, hito, imagen, ganada, color], i) => {
    const enCaja = i < 8;
    const x = enCaja ? 102 + (i % 4) * 132 : 680 + ((i - 8) % 2) * 130;
    const y = 150 + Math.floor((enCaja ? i : i - 8) / (enCaja ? 4 : 2)) * 118;
    const d = (0.25 + i * 0.12).toFixed(2);
    const lado = enCaja ? 70 : 64;
    const img = `<image x="${x - lado / 2}" y="${y - lado / 2}" width="${lado}" height="${lado}" href="${imagen}"${ganada ? '' : ' filter="url(#silueta)"'}/>`;
    const fondo = enCaja
      ? `<circle cx="${x}" cy="${y}" r="42" fill="#6e1d1d"/>`
      : `<circle cx="${x}" cy="${y}" r="42" fill="${color}" stroke="#141630" stroke-width="3"/><circle cx="${x}" cy="${y}" r="36" fill="none" stroke="#ffffff" stroke-opacity="0.5" stroke-width="2"/>`;
    const figura = ganada
      ? `<g class="pop" style="animation-delay:${d}s">${img}</g>`
      : `${img}<circle cx="${x}" cy="${y}" r="42" fill="none" stroke="#ffe9c8" stroke-opacity="0.6" stroke-width="2" stroke-dasharray="6 5"/>`;
    const claro = enCaja ? '#ffe9c8' : '#e6e9ff';
    const tenue = enCaja ? '#e6a88c' : '#8b90c0';
    cuerpo += `    ${fondo}
    ${figura}
    <text x="${x}" y="${y + 60}" font-family="${SANS}" font-size="14" font-weight="700" fill="${claro}" text-anchor="middle">${nombre}</text>
    <text x="${x}" y="${y + 76}" font-family="${SANS}" font-size="11" fill="${tenue}" text-anchor="middle">${ganada ? `${i + 1} · ${hito}` : 'falta el editor'}</text>
${ganada ? `    <path class="destello" style="animation-delay:${(1.8 + i * 0.37).toFixed(2)}s" d="M${x + 26} ${y - 36} l3 8 l8 3 l-8 3 l-3 8 l-3 -8 l-8 -3 l8 -3 z" fill="#ffffff"/>\n` : ''}`;
  });

  const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="Estuche de medallas: ${ganadas} de 12. Las ocho medallas de Kanto y el Alto Mando, una por hito; falta la Medalla Roca, el editor del hito 1">
  <defs>
    <filter id="silueta"><feColorMatrix type="matrix" values="0 0 0 0 0.12  0 0 0 0 0.05  0 0 0 0 0.05  0 0 0 0.7 0"/></filter>
  </defs>
  <style>
    .pop { opacity: 0; transform-box: fill-box; transform-origin: center; animation: pop 0.45s cubic-bezier(.34,1.56,.64,1) both; }
    @keyframes pop { from { opacity: 0; transform: scale(0.3); } to { opacity: 1; transform: scale(1); } }
    .destello { opacity: 0; transform-box: fill-box; transform-origin: center; animation: destello 4.4s ease-in-out infinite; }
    @keyframes destello { 0%, 100% { opacity: 0; transform: scale(0.4); } 8% { opacity: 1; transform: scale(1.2); } 16% { opacity: 0; transform: scale(0.4); } }
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
// Cada escena tiene al Pokémon respirando sobre su plataforma y un cuadro de
// diálogo como el de los juegos, que se escribe letra por letra.
const GUIAS = [
  [
    'charizard',
    6,
    'Charizard',
    ['#ffc58a', '#e8603a'],
    '¡Hola, entrenador! PokeScript es un lenguaje en español para aprender a programar. Los datos tienen tipo, las funciones son movimientos y los errores se explican en tu idioma.',
  ],
  [
    'rotom',
    479,
    'Rotom',
    ['#fff2a8', '#e0a820'],
    'Así va a verse el IDE: el código, el asistente al lado y la salida abajo. Si escribes `curra` en vez de `curar`, el asistente te ofrece el arreglo.',
  ],
  [
    'pikachu',
    25,
    'Pikachu',
    ['#fff2a8', '#e8b820'],
    'Cada tipo de dato es un tipo de Pokémon. El mío es `electrico`: solo sé decir `verdadero` o `falso`, pero sin mí no hay `si` ni `mientras`.',
  ],
  [
    'charmander',
    4,
    'Charmander',
    ['#ffc58a', '#e8603a'],
    'Todo programa empieza en `combate` y termina en su `fin`. Lo de adentro corre de arriba abajo, una instrucción por línea.',
  ],
  [
    'onix',
    95,
    'Onix',
    ['#e0d4a0', '#9a8448'],
    'Cada dato se declara con su tipo. Una `medalla` es un valor que no cambia nunca, duro como la roca. Por eso va en mayúsculas.',
  ],
  [
    'psyduck',
    54,
    'Psyduck',
    ['#b0dcff', '#4a88d8'],
    'Aquí no hay «más o menos»: la condición de un `si` da `verdadero` o `falso`. `si vida` no compila; `si vida > 0` sí.',
  ],
  [
    'tauros',
    128,
    'Tauros',
    ['#e8dcc0', '#a08858'],
    '`mientras` embiste una y otra vez hasta que la condición deja de cumplirse. `recorrer` pasa por un rango o por un equipo. `huir` sale del ciclo y `siguiente` salta a la otra vuelta.',
  ],
  [
    'machamp',
    68,
    'Machamp',
    ['#f4b8a0', '#c04a38'],
    'Los movimientos son las funciones. Unos devuelven un valor con `entregar`; otros solo hacen algo, como gritar un mensaje. ¡Cuatro brazos, cero efectos colaterales!',
  ],
  [
    'kangaskhan',
    115,
    'Kangaskhan',
    ['#e8dcc0', '#a08858'],
    'Un `equipo` es una lista. Una `mochila` guarda cada cosa con su clave, como mi bolsa, y recuerda el orden. Se cuenta desde 1, como el primer Pokémon de tu equipo.',
  ],
  [
    'eevee',
    133,
    'Eevee',
    ['#f0dcc0', '#b08050'],
    'Una `especie` es una lista cerrada de valores, como mis evoluciones: no hay más que esas. Una `ficha` junta varios datos bajo un mismo nombre.',
  ],
  [
    'ditto',
    132,
    'Ditto',
    ['#ecd0f4', '#a070c0'],
    'Cambiar de forma es lo mío, pero aquí se pide con `convertir`. Ojo: `convertir(agua) a roca` corta los decimales; para redondear está `redondear`.',
  ],
  [
    'abra',
    63,
    'Abra',
    ['#ffc0d8', '#d85890'],
    'Con `enseñar … desde` un archivo se teletransporta lo que declara otro: movimientos, especies, fichas y medallas.',
  ],
  [
    'machop',
    66,
    'Machop',
    ['#f4b8a0', '#c04a38'],
    'Un equipo de niveles, una mochila de objetos y varios ciclos para entrenar. ¡A sudar!',
  ],
  [
    'gengar',
    94,
    'Gengar',
    ['#cbb4ec', '#6a4a9a'],
    'Una `especie` con los estados alterados, un `segun` que los cubre todos y movimientos que dicen qué le pasa a cada Pokémon. Je, je.',
  ],
  [
    'chansey',
    113,
    'Chansey',
    ['#ffd4e2', '#e0789a'],
    '`capturar` se queda esperando lo que escribas. Si la respuesta no sirve, lo explica y vuelve a preguntar, con la paciencia de un Centro Pokémon.',
  ],
  [
    'magikarp',
    129,
    'Magikarp',
    ['#b0dcff', '#4a88d8'],
    'Olvidar un `fin` es el error más común al empezar. El mensaje dice qué bloque quedó abierto, en qué línea empezó y cuál es el que falta cerrar.',
  ],
  [
    'blastoise',
    9,
    'Blastoise',
    ['#b0dcff', '#3a70c0'],
    'El programa de la especificación: cuatro archivos que se importan entre sí. Tu Pokémon contra Bulbi, veinte turnos como máximo.',
  ],
  [
    'porygon',
    137,
    'Porygon',
    ['#b4ecf4', '#3a98b8'],
    'Un programa pasa por cuatro etapas antes de mostrar algo. Si el analizador no encuentra un nombre, el asistente busca qué quisiste escribir.',
  ],
  [
    'dragonite',
    149,
    'Dragonite',
    ['#ffdca0', '#d89030'],
    'Cada hito es una medalla: primero los ocho gimnasios de Kanto y después el Alto Mando. Falta la Medalla Roca, que llega con los colores del editor.',
  ],
  [
    'snorlax',
    143,
    'Snorlax',
    ['#c8dce4', '#58788a'],
    'Para despertarme hace falta Go 1.23 o más nuevo. Si vas a trabajar en el proyecto, también Node.js 22 y pnpm 10. Con npm no me muevo. Zzz…',
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

{
  const W = 900;
  const H = 220;
  mkdirSync(join(OUT, 'guias'), { recursive: true });
  for (const [archivo, id, nombre, [claro, oscuro], texto] of GUIAS) {
    const lineas = envolver(texto, 58);
    if (lineas.length > 4) throw new Error(`${archivo}: el texto no cabe en el cuadro`);
    const e = escritor('l');
    let t = 0.8;
    lineas.forEach((l, i) => {
      t = e.linea(262, 82 + i * 30, l, { color: '#2d2d3a', fondo: '#ffffff', t, vel: 0.03 });
    });
    const burbujas = [30, 70, 120, 160, 190, 55]
      .map(
        (x, i) =>
          `<circle class="burbuja" cx="${x}" cy="${200 - (i % 3) * 12}" r="${3 + (i % 3)}" fill="#ffffff" style="animation-delay:${(i * 0.9).toFixed(1)}s"/>`,
      )
      .join('\n    ');
    const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="${nombre}: ${esc(texto.replace(/`/g, ''))}">
  <defs>
    <linearGradient id="cielo" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${claro}"/><stop offset="1" stop-color="${oscuro}"/></linearGradient>
    <clipPath id="dialogo"><rect x="241" y="42" width="628" height="148"/></clipPath>
  </defs>
  <style>
    .entra { animation: entra 0.7s cubic-bezier(.34,1.56,.64,1) both; }
    @keyframes entra { from { opacity: 0; transform: translateX(-70px); } to { opacity: 1; transform: none; } }
    .respira { animation: respira 2.6s ease-in-out 0.7s infinite; }
    @keyframes respira { 0%, 100% { transform: translateY(0); } 50% { transform: translateY(-7px); } }
    .sombra { transform-box: fill-box; transform-origin: center; animation: sombra 2.6s ease-in-out 0.7s infinite; }
    @keyframes sombra { 0%, 100% { transform: scale(1); } 50% { transform: scale(0.88); } }
    .burbuja { opacity: 0; animation: burbuja 5.4s ease-out infinite; }
    @keyframes burbuja { 0% { opacity: 0; transform: translateY(0); } 15% { opacity: 0.7; } 100% { opacity: 0; transform: translateY(-170px); } }
    .flecha { opacity: 0; animation: flecha 0.9s steps(1) ${t.toFixed(2)}s infinite; }
    @keyframes flecha { 0% { opacity: 1; } 50% { opacity: 0; } }
${e.css}  </style>
  <rect width="${W}" height="${H}" rx="16" fill="url(#cielo)"/>
  <g opacity="0.55">
    ${burbujas}
  </g>
  <ellipse cx="118" cy="192" rx="92" ry="18" fill="${oscuro}" opacity="0.55"/>
  <ellipse class="sombra" cx="118" cy="190" rx="62" ry="9" fill="#000000" opacity="0.18"/>
  <g class="entra"><g class="respira">
    <image x="28" y="14" width="180" height="176" href="${pokemon(id)}"/>
  </g></g>
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

console.log('escenas listas');

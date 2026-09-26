// Genera las escenas animadas del README: la Pokédex, el combate de tipos,
// las medallas de la hoja de ruta y el recorrido de un programa por el
// compilador. Uso: pnpm docs:svg (escribe en docs/assets/).
import { writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const OUT = process.argv[2] ?? fileURLToPath(new URL('../docs/assets/', import.meta.url));
const MONO = "Consolas, 'Courier New', monospace";
const SANS = "'Trebuchet MS', 'Segoe UI', Verdana, sans-serif";

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
    linea(x, y, texto, { cw = 10, fs = 17, color, fondo, t, vel = 0.035, peso = 400 }) {
      const chars = largo(texto);
      const ancho = chars * cw;
      const dur = Math.max(0.2, chars * vel);
      const id = `${prefijo}${n++}`;
      css += `    @keyframes ${id} { to { transform: translateX(${ancho + 4}px); } }\n`;
      svg += `    <text x="${x}" y="${y}" font-family="${MONO}" font-size="${fs}" font-weight="${peso}" fill="${color}" textLength="${ancho}" lengthAdjust="spacingAndGlyphs">${esc(texto)}</text>\n`;
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
{
  const W = 900;
  const H = 330;
  // Formas de medalla, dibujadas alrededor de (0, 0) con radio ~34.
  const estrella = (puntas, r1, r2) => {
    let d = '';
    for (let i = 0; i < puntas * 2; i++) {
      const r = i % 2 ? r2 : r1;
      const a = (Math.PI * i) / puntas - Math.PI / 2;
      d += `${i ? 'L' : 'M'}${(r * Math.cos(a)).toFixed(1)} ${(r * Math.sin(a)).toFixed(1)} `;
    }
    return `<path d="${d}Z"/>`;
  };
  const poligono = (lados, r, giro = 0) =>
    estrella(lados, r, r * Math.cos(Math.PI / lados)).replace(
      '<path',
      `<path transform="rotate(${giro})"`,
    );
  const formas = [
    poligono(8, 34, 22.5), // roca
    '<path d="M0 -36 C18 -12 26 2 26 12 A26 26 0 0 1 -26 12 C-26 2 -18 -12 0 -36 Z"/>', // gota
    estrella(8, 36, 18), // trueno
    '<g><circle cx="0" cy="-17" r="16"/><circle cx="17" cy="0" r="16"/><circle cx="0" cy="17" r="16"/><circle cx="-17" cy="0" r="16"/></g>', // flor
    '<path d="M0 30 C-40 4 -30 -30 0 -14 C30 -30 40 4 0 30 Z"/>', // corazón
    '<g><circle r="34"/><circle r="22" fill-opacity="0.55"/><circle r="10"/></g>', // círculos
    '<path d="M0 -36 C14 -20 26 -8 22 12 A22 22 0 0 1 -22 12 C-26 -6 -12 -14 -6 -26 C-2 -14 6 -12 0 -36 Z"/>', // llama
    '<path d="M-30 30 C-30 -10 0 -34 32 -32 C34 0 10 30 -30 30 Z"/>', // hoja
    poligono(4, 36), // rombo
    poligono(6, 34, 30), // hexágono
    '<path d="M0 -34 L30 -22 L26 12 C22 26 10 32 0 36 C-10 32 -22 26 -26 12 L-30 -22 Z"/>', // escudo
    estrella(5, 36, 16), // estrella
  ];
  const medallas = [
    ['Lexer', '#b8a038', false],
    ['Expresiones', '#6890f0', true],
    ['Bloques', '#f8d030', true],
    ['Primer programa', '#f08030', true],
    ['Tipos', '#e0588a', true],
    ['Movimientos', '#a060d8', true],
    ['Colecciones', '#e05030', true],
    ['Segun', '#58b848', true],
    ['Importaciones', '#40a8c8', true],
    ['Posible', '#8878d8', true],
    ['Fichas', '#c89838', true],
    ['Asistente', '#e8b820', true],
  ];
  const ganadas = medallas.filter((m) => m[2]).length;
  let css = '';
  let cuerpo = '';
  medallas.forEach(([nombre, color, ganada], i) => {
    const x = 90 + (i % 6) * 144;
    const y = 118 + Math.floor(i / 6) * 118;
    const d = (0.25 + i * 0.12).toFixed(2);
    const forma = formas[i];
    const relleno = ganada
      ? `fill="url(#g${i})" stroke="#3a3a3a" stroke-width="2.5"`
      : 'fill="#3a3f66" stroke="#6b70a8" stroke-width="2.5" stroke-dasharray="5 4"';
    cuerpo += `    <defs><linearGradient id="g${i}" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#ffffff"/><stop offset="0.25" stop-color="${color}"/><stop offset="1" stop-color="${color}" stop-opacity="0.7"/></linearGradient></defs>
    <g class="pop" style="animation-delay:${d}s"><g transform="translate(${x} ${y})" ${relleno}>${forma}</g>
      <text x="${x}" y="${y + 56}" font-family="${SANS}" font-size="14" font-weight="700" fill="${ganada ? '#e6e9ff' : '#8b90c0'}" text-anchor="middle">${nombre}</text>
      ${ganada ? '' : `<text x="${x}" y="${y + 74}" font-family="${SANS}" font-size="11" fill="#8b90c0" text-anchor="middle">falta el editor</text>`}
      ${ganada ? `<path class="destello" style="animation-delay:${(1.8 + i * 0.37).toFixed(2)}s" d="M${x + 24} ${y - 34} l3 8 l8 3 l-8 3 l-3 8 l-3 -8 l-8 -3 l8 -3 z" fill="#ffffff"/>` : ''}
    </g>\n`;
  });
  css += `    .pop { opacity: 0; transform-box: fill-box; transform-origin: center; animation: pop 0.45s cubic-bezier(.34,1.56,.64,1) both; }
    @keyframes pop { from { opacity: 0; transform: scale(0.3); } to { opacity: 1; transform: scale(1); } }
    .destello { opacity: 0; transform-box: fill-box; transform-origin: center; animation: destello 4.4s ease-in-out infinite; }
    @keyframes destello { 0%, 100% { opacity: 0; transform: scale(0.4); } 8% { opacity: 1; transform: scale(1.2); } 16% { opacity: 0; transform: scale(0.4); } }
`;
  const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="Medallas de la hoja de ruta: ${ganadas} de 12 hitos ganados; falta el editor del hito 1">
  <style>
${css}  </style>
  <rect width="${W}" height="${H}" rx="16" fill="#1b1d36"/>
  <text x="40" y="44" font-family="${SANS}" font-size="22" font-weight="900" fill="#ffcb05">Medallas: ${ganadas} de 12</text>
  <text x="${W - 40}" y="44" font-family="${SANS}" font-size="14" fill="#8b90c0" text-anchor="end">una por cada hito de la especificación</text>
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

console.log('escenas listas');

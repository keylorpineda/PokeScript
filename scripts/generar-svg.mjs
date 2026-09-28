// Genera SVG animados de "ventana de editor" con resaltado de PokeScript.
import { writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

// Uso: pnpm docs:svg   (escribe en docs/assets/)
const OUT = process.argv[2] ?? fileURLToPath(new URL('../docs/assets/', import.meta.url));
const CW = 9; // ancho de carácter (px): cada token se posiciona en esta rejilla
const LH = 23; // alto de línea
const X0 = 62; // x donde empieza el código

const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

const KEYWORDS = new Set(
  (
    'especie ficha medalla combate movimiento entregar fin enseñar desde si sino segun entonces otro ' +
    'mientras recorrer en hasta huir siguiente gritar capturar convertir tamaño aleatorio redondear ' +
    'sumar quitar contiene resto igual diferente y o no'
  ).split(' '),
);
const TYPES = {
  roca: '#d4bc4a',
  agua: '#7fa6ff',
  fuego: '#ff9a4d',
  planta: '#8fd46a',
  electrico: '#ffd84a',
};
const TYPE_WORDS = new Set(['equipo', 'mochila', 'posible', 'de', 'a']);
const LITERALS = new Set(['verdadero', 'falso', 'fantasma']);

function tokenize(line) {
  const out = [];
  const re =
    /(\/\/.*$)|("(?:[^"\\]|\\.)*")|('(?:[^'\\]|\\.)')|(\d+(?:\.\d+)?)|([\p{L}_][\p{L}\p{Nd}_]*)|(\s+)|(.)/gu;
  let m;
  while ((m = re.exec(line))) {
    const col = [...line.slice(0, m.index)].length;
    const t = m[0];
    let cls;
    if (m[1]) cls = 'c';
    else if (m[2] || m[3]) cls = 's';
    else if (m[4]) cls = 'num';
    else if (m[5]) {
      const rest = line.slice(m.index + t.length);
      if (KEYWORDS.has(t)) cls = 'k';
      else if (TYPES[t]) cls = 'ty-' + t;
      else if (TYPE_WORDS.has(t)) cls = 'tw';
      else if (LITERALS.has(t)) cls = 'lit';
      else if (/^\s*\(/.test(rest)) cls = 'f';
      else if (/^[\p{Lu}][\p{Lu}\p{Nd}_]+$/u.test(t)) cls = 'cte';
      else if (/^\p{Lu}/u.test(t)) cls = 'tn';
      else cls = 'id';
    } else if (m[6]) continue;
    else cls = 'op';
    out.push({ col, t, cls });
  }
  return out;
}

const STYLE = `
    text { font-family: Consolas, 'JetBrains Mono', 'Fira Code', Menlo, 'Courier New', monospace; font-size: 15px; }
    .n { fill: #4b5080; text-anchor: end; }
    .id { fill: #d6deeb; } .k { fill: #c792ea; font-weight: 700; } .f { fill: #82aaff; }
    .s { fill: #c3e88d; } .num { fill: #f78c6c; } .op { fill: #89ddff; } .lit { fill: #f78c6c; font-weight: 700; }
    .c { fill: #637777; font-style: italic; } .cte { fill: #ffcb6b; } .tn { fill: #7fdbca; } .tw { fill: #ffcb6b; }
    ${Object.entries(TYPES)
      .map(([k, v]) => `.ty-${k} { fill: ${v}; font-weight: 700; }`)
      .join(' ')}
    .linea { opacity: 0; animation: entrar 0.35s ease-out forwards; }
    .aparece { opacity: 0; animation: entrar 0.45s ease-out forwards; }
    @keyframes entrar { from { opacity: 0; transform: translateX(-12px); } to { opacity: 1; transform: translateX(0); } }
    .pop { opacity: 0; transform-box: fill-box; transform-origin: center; animation: pop 0.5s cubic-bezier(.34,1.56,.64,1) forwards; }
    @keyframes pop { from { opacity: 0; transform: scale(0.85); } to { opacity: 1; transform: scale(1); } }
    .cursor { opacity: 0; animation: mostrar 0.01s forwards, parpadeo 0.9s step-end infinite; }
    @keyframes mostrar { to { opacity: 1; } }
    @keyframes parpadeo { 50% { fill: transparent; } }
    .punto { animation: latido 1.6s ease-in-out infinite; }
    @keyframes latido { 50% { opacity: 0.35; } }
    .onda { stroke-dasharray: 400; stroke-dashoffset: 400; animation: trazar 0.8s ease-out forwards; }
    @keyframes trazar { to { stroke-dashoffset: 0; } }
    .resalte { opacity: 0; animation: resaltar 0.4s forwards, latir 1.8s ease-in-out infinite; }
    @keyframes resaltar { to { opacity: 1; } }
    @keyframes latir { 50% { fill-opacity: 0.08; } }
`;

function editor({
  file,
  tabs = [],
  code,
  output = [],
  panel = null,
  error = null,
  width = 900,
  title,
}) {
  const lines = code.split('\n');
  const step = Math.min(0.09, 2.2 / lines.length);
  const codeTop = 78;
  const codeEnd = codeTop + (lines.length - 1) * LH;
  const outTop = codeEnd + 30;
  const outH = output.length ? 52 + output.length * 27 + 18 : 0;
  const height = outTop + outH;
  const tEndCode = 0.1 + lines.length * step;

  let s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${width} ${height}" width="${width}" height="${height}" role="img" aria-label="${esc(title)}">
  <defs><clipPath id="ventana"><rect width="${width}" height="${height}" rx="16"/></clipPath></defs>
  <style>${STYLE}  </style>
  <g clip-path="url(#ventana)">
    <rect width="${width}" height="${height}" fill="#1b1d36"/>
    <rect width="${width}" height="44" fill="#121328"/>
    <circle cx="24" cy="22" r="7" fill="#ff5f57"/><circle cx="46" cy="22" r="7" fill="#febc2e"/><circle cx="68" cy="22" r="7" fill="#28c840"/>
    <rect x="96" y="8" width="${file.length * 8.4 + 48}" height="36" rx="8" fill="#1b1d36"/>
    <circle cx="114" cy="26" r="6" fill="#ee1515"/><rect x="108" y="25" width="12" height="2" fill="#121328"/>
    <text x="128" y="31" style="font-size:14px" fill="#e6e9ff">${esc(file)}</text>
`;
  let tx = 96 + file.length * 8.4 + 68;
  for (const t of tabs) {
    s += `    <text x="${tx}" y="31" style="font-size:14px" fill="#6b70a8">${esc(t)}</text>\n`;
    tx += t.length * 8.4 + 28;
  }
  s += `    <circle class="punto" cx="${width - 28}" cy="22" r="5" fill="${error ? '#ff5f57' : '#78c850'}"/>\n`;

  // Resalte de la línea con error (detrás del código)
  if (error) {
    const y = codeTop + (error.line - 1) * LH - 16;
    s += `    <rect class="resalte" style="animation-delay:${error.at}s, ${error.at + 0.4}s" x="0" y="${y}" width="${width}" height="${LH}" fill="#ff5f57" fill-opacity="0.16"/>\n`;
  }

  s += '    <!-- Código -->\n';
  lines.forEach((ln, i) => {
    const y = codeTop + i * LH;
    const tokens = tokenize(ln)
      .map(
        (t) =>
          `<tspan x="${X0 + t.col * CW}" textLength="${[...t.t].length * CW}" lengthAdjust="spacingAndGlyphs" class="${t.cls}">${esc(t.t)}</tspan>`,
      )
      .join('');
    s += `    <g class="linea" style="animation-delay:${(0.1 + i * step).toFixed(2)}s"><text x="44" y="${y}" class="n">${i + 1}</text>${tokens ? `<text y="${y}">${tokens}</text>` : ''}</g>\n`;
  });

  if (error) {
    const y = codeTop + (error.line - 1) * LH + 6;
    const x1 = X0 + error.col * CW;
    const x2 = x1 + error.len * CW;
    let d = `M${x1} ${y}`;
    for (let x = x1, up = true; x < x2; x += 4, up = !up)
      d += ` L${Math.min(x + 4, x2)} ${up ? y - 3 : y}`;
    s += `    <path class="onda" style="animation-delay:${error.at}s" d="${d}" fill="none" stroke="#ff5f57" stroke-width="2"/>\n`;
  }

  if (panel) {
    const { x, y, w, h, at, accent, label, heading, body, button } = panel;
    s += `    <g class="pop" style="animation-delay:${at}s">
      <rect x="${x}" y="${y}" width="${w}" height="${h}" rx="12" fill="#23264a" stroke="#3a3f7a" stroke-width="1.5"/>
      <rect x="${x}" y="${y}" width="6" height="${h}" rx="3" fill="${accent}"/>
      <text x="${x + 24}" y="${y + 30}" style="font-size:12px; letter-spacing:2px; font-weight:700" fill="${accent}">${esc(label)}</text>
      <text x="${x + 24}" y="${y + 60}" style="font-size:16px; font-weight:700" fill="#e6e9ff">${esc(heading)}</text>
`;
    body.forEach((b, i) => {
      s += `      <text x="${x + 24}" y="${y + 92 + i * 22}" style="font-size:13.5px" fill="#b8bde8">${b}</text>\n`;
    });
    if (button) {
      const by = y + h - 50;
      s += `      <rect x="${x + 24}" y="${by}" width="${button.length * 8.5 + 32}" height="32" rx="8" fill="${accent}"/>
      <text x="${x + 24 + (button.length * 8.5 + 32) / 2}" y="${by + 21}" style="font-size:14px; font-weight:700" fill="#121328" text-anchor="middle">${esc(button)}</text>\n`;
    }
    s += '    </g>\n';
  }

  if (output.length) {
    s += `    <!-- Salida -->
    <rect x="0" y="${outTop}" width="${width}" height="${outH}" fill="#121328"/>
    <rect x="0" y="${outTop}" width="${width}" height="1" fill="#2c3060"/>
    <text x="24" y="${outTop + 28}" style="font-size:12px; letter-spacing:2px" fill="#6b70a8">SALIDA</text>
`;
    let t = (error ? error.at : tEndCode) + 0.5;
    output.forEach((o, i) => {
      const y = outTop + 58 + i * 27;
      s += `    <g class="aparece" style="animation-delay:${t.toFixed(2)}s"><text x="24" y="${y}">${o}</text></g>\n`;
      t += o.includes('›') ? 0.9 : 0.45;
    });
    s += `    <rect class="cursor" style="animation-delay:${t.toFixed(2)}s, ${t.toFixed(2)}s" x="24" y="${outTop + 58 + output.length * 27 - 14}" width="9" height="17" fill="#ffcb05"/>\n`;
  }
  s += '  </g>\n</svg>\n';
  return s;
}

const ok = (t) => `<tspan fill="#78c850" font-weight="700">${t}</tspan>`;
const dim = (t) => `<tspan fill="#8b90c0">${t}</tspan>`;
const out = (t) => `<tspan fill="#e6e9ff">${t}</tspan>`;
const inp = (t) => `<tspan fill="#ffcb05">› </tspan><tspan fill="#ffcb05">${t}</tspan>`;

// 1. Colecciones y ciclos
writeFileSync(
  join(OUT, 'ejemplo-colecciones.svg'),
  editor({
    title: 'Ejemplo de PokeScript con equipo, mochila y ciclos',
    file: 'entrenamiento.pks',
    code: `// entrenamiento.pks · equipo, mochila y ciclos
combate
    equipo de planta mi_equipo = ["Pikachu", "Charmander", "Squirtle"]
    sumar "Bulbasaur" a mi_equipo
    gritar "Tu equipo tiene ", tamaño(mi_equipo), " Pokémon"

    mochila de planta a roca bayas = {"Aranja": 3, "Zreza": 0, "Meloc": 5}
    recorrer baya, cantidad en bayas
        si cantidad igual 0
            siguiente
        fin
        gritar "  ", baya, " x", cantidad
    fin

    roca nivel = 5
    mientras nivel < 10
        nivel = nivel + 2
    fin
    gritar "Nivel final: ", nivel
fin`,
    output: [
      ok('✔ ¡Es superefectivo!') + dim(' Compilación sin errores'),
      out('Tu equipo tiene 4 Pokémon'),
      out('  Aranja x3'),
      out('  Meloc x5'),
      out('Nivel final: 11'),
    ],
  }),
);

// 2. Especies, fichas y segun
writeFileSync(
  join(OUT, 'ejemplo-especies.svg'),
  editor({
    title: 'Ejemplo de PokeScript con especie, ficha, segun y movimientos',
    file: 'estados.pks',
    panel: {
      x: 500,
      y: 92,
      w: 376,
      h: 212,
      at: 3.2,
      accent: '#ffcb05',
      label: 'SI BORRAS LA RAMA DORMIDO…',
      heading: 'No es muy efectivo…',
      body: [
        'El <tspan fill="#c792ea" font-weight="700">segun</tspan> sobre <tspan fill="#7fdbca">Estado</tspan> no cubre todos',
        'los casos. Falta: <tspan fill="#ffcb6b" font-weight="700">DORMIDO</tspan>.',
        'El compilador te nombra cada caso',
        'que olvidaste. ¡Ninguno se escapa!',
      ],
    },
    code: `// estados.pks · especie, ficha, segun y movimientos
especie Estado
    SANO, ENVENENADO, DORMIDO
fin

ficha Pokemon
    planta nombre
    roca   vida
    Estado estado
fin

movimiento roca pasar_turno(Pokemon p)
    roca dano = 0
    segun p.estado
        SANO        entonces gritar p.nombre, " está en plena forma"
        ENVENENADO  entonces dano = 10
        DORMIDO     entonces gritar p.nombre, " sigue dormido…"
    fin
    entregar p.vida - dano
fin

combate
    Pokemon bulbi = {nombre: "Bulbasaur", vida: 45, estado: ENVENENADO}
    recorrer t de 1 hasta 3
        bulbi.vida = pasar_turno(bulbi)
        gritar "Turno ", t, ": ", bulbi.nombre, " tiene ", bulbi.vida, " PS"
    fin
fin`,
    output: [
      ok('✔ ¡Es superefectivo!') + dim(' segun cubre SANO, ENVENENADO y DORMIDO'),
      out('Turno 1: Bulbasaur tiene 35 PS'),
      out('Turno 2: Bulbasaur tiene 25 PS'),
      out('Turno 3: Bulbasaur tiene 15 PS'),
    ],
  }),
);

// 3. Error: bloque sin cerrar
writeFileSync(
  join(OUT, 'ejemplo-error.svg'),
  editor({
    title: 'PokeScript detecta un bloque mientras sin cerrar y dice en qué línea se abrió',
    file: 'captura.pks',
    code: `// captura.pks · ¿dónde quedó el fin?
combate
    planta nombre
    capturar(nombre, "¿Nombre? ")
    roca intentos = 0
    mientras intentos < 3
        intentos = intentos + 1
        si intentos igual 3
            gritar "¡", nombre, " atrapado!"
        fin
fin`,
    error: { line: 6, col: 4, len: 8, at: 2.0 },
    panel: {
      x: 470,
      y: 62,
      w: 406,
      h: 236,
      at: 2.5,
      accent: '#ff5f57',
      label: 'ERROR SINTÁCTICO · LÍNEA 11',
      heading: '¡Se escapó!',
      body: [
        'El bloque <tspan fill="#c792ea" font-weight="700">combate</tspan> de la <tspan fill="#ffcb6b" font-weight="700">línea 2</tspan> nunca se cerró.',
        'Por la sangría, parece que falta el <tspan fill="#c792ea" font-weight="700">fin</tspan>',
        'del <tspan fill="#c792ea" font-weight="700">mientras</tspan> que abriste en la <tspan fill="#ffcb6b" font-weight="700">línea 6</tspan>.',
      ],
      button: 'Agregar fin',
    },
    output: [
      '<tspan fill="#ff5f57" font-weight="700">✖ ¡Se escapó!</tspan>' +
        dim(' 1 error · captura.pks:11'),
      dim('El programa no se ejecuta hasta corregirlo.'),
    ],
  }),
);

// 4. Capturar: entrada interactiva
writeFileSync(
  join(OUT, 'ejemplo-captura.svg'),
  editor({
    title: 'Ejemplo de PokeScript que pide datos con capturar y valida la entrada',
    file: 'centro.pks',
    code: `// centro.pks · capturar, contiene y sino si
movimiento electrico es_legendario(planta nombre)
    equipo de planta legendarios = ["Mewtwo", "Lugia", "Rayquaza"]
    entregar legendarios contiene nombre
fin

combate
    planta nombre
    roca nivel
    capturar(nombre, "¿Qué Pokémon atrapaste? ")
    capturar(nivel, "¿De qué nivel? ")

    si es_legendario(nombre)
        gritar "¡Increíble! ", nombre, " es legendario"
    sino si nivel >= 50
        gritar nombre, " ya es todo un veterano"
    sino
        agua progreso = nivel / 100
        gritar nombre, " va al ", redondear(progreso * 100), "% del camino"
    fin
fin`,
    output: [
      dim('¿Qué Pokémon atrapaste? ') + inp('Eevee'),
      dim('¿De qué nivel? ') + inp('doce'),
      '<tspan fill="#ffcb05" font-weight="700">¿Seguro?</tspan>' +
        dim(' "doce" no es un número roca. Intenta de nuevo.'),
      dim('¿De qué nivel? ') + inp('12'),
      out('Eevee va al 12% del camino'),
    ],
  }),
);

// 5. Tabla de efectividades animada
{
  const tipos = ['roca', 'agua', 'fuego', 'planta', 'electrico', 'especie'];
  const color = {
    roca: '#b8a038',
    agua: '#6890f0',
    fuego: '#f08030',
    planta: '#78c850',
    electrico: '#e0b000',
    especie: '#a040a0',
  };
  const tabla = [
    ['MT', 'EF', 'SE', 'RC', 'SE', 'SE'],
    ['RC', 'MT', 'SE', 'RC', 'SE', 'SE'],
    ['SE', 'SE', 'MT', 'RC', 'SE', 'SE'],
    ['RC', 'RC', 'SE', 'MT', 'SE', 'SE'],
    ['SE', 'SE', 'SE', 'RC', 'MT', 'SE'],
    ['SE', 'SE', 'SE', 'RC', 'SE', 'MT'],
  ];
  const celda = {
    MT: ['#3a3f7a', '#e6e9ff', '='],
    EF: ['#2f8f4e', '#ffffff', '×2'],
    RC: ['#b38a00', '#ffffff', '½'],
    SE: ['#9e2b2b', '#ffffff', '×0'],
  };
  const W = 900,
    H = 470,
    cx0 = 190,
    cy0 = 96,
    cw = 110,
    ch = 46;
  let s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="${W}" height="${H}" role="img" aria-label="Tabla de efectividades de PokeScript: qué conversiones entre tipos son automáticas, requieren convertir o son un error">
  <defs><clipPath id="marco"><rect width="${W}" height="${H}" rx="16"/></clipPath></defs>
  <style>
    text { font-family: 'Trebuchet MS', 'Segoe UI', Verdana, sans-serif; }
    .pop { opacity: 0; transform-box: fill-box; transform-origin: center; animation: pop 0.45s cubic-bezier(.34,1.56,.64,1) forwards; }
    @keyframes pop { from { opacity: 0; transform: scale(0.2); } to { opacity: 1; transform: scale(1); } }
    .brillo { animation: brillo 3s ease-in-out infinite; }
    @keyframes brillo { 0%, 100% { opacity: 0; } 50% { opacity: 0.18; } }
  </style>
  <g clip-path="url(#marco)">
    <rect width="${W}" height="${H}" fill="#1b1d36"/>
    <text x="30" y="46" font-size="24" font-weight="900" fill="#ffcb05">Tabla de efectividades</text>
    <text x="30" y="70" font-size="14" fill="#8b90c0">Origen (fila) → destino (columna). ¿Se puede pasar un valor de un tipo a otro?</text>
`;
  tipos.forEach((t, j) => {
    const x = cx0 + j * cw;
    s += `    <g class="pop" style="animation-delay:${(0.05 * j).toFixed(2)}s"><rect x="${x + 6}" y="${cy0}" width="${cw - 12}" height="30" rx="15" fill="${color[t]}"/><text x="${x + cw / 2}" y="${cy0 + 20}" font-size="14" font-weight="700" fill="#fff" text-anchor="middle">${t}</text></g>\n`;
  });
  tipos.forEach((t, i) => {
    const y = cy0 + 44 + i * ch;
    s += `    <g class="pop" style="animation-delay:${(0.05 * i).toFixed(2)}s"><rect x="30" y="${y + 4}" width="140" height="30" rx="15" fill="${color[t]}"/><text x="100" y="${y + 24}" font-size="14" font-weight="700" fill="#fff" text-anchor="middle">${t}</text></g>\n`;
    tabla[i].forEach((v, j) => {
      const x = cx0 + j * cw;
      const [bg, fg, sim] = celda[v];
      const d = (0.4 + (i + j) * 0.07).toFixed(2);
      s += `    <g class="pop" style="animation-delay:${d}s"><rect x="${x + 6}" y="${y}" width="${cw - 12}" height="${ch - 8}" rx="8" fill="${bg}"/><text x="${x + cw / 2}" y="${y + 25}" font-size="15" font-weight="700" fill="${fg}" text-anchor="middle">${v} ${sim}</text></g>\n`;
    });
  });
  // Resalta la diagonal (mismo tipo) con un brillo
  tipos.forEach((_, i) => {
    s += `    <rect class="brillo" style="animation-delay:${(2 + i * 0.25).toFixed(2)}s" x="${cx0 + i * cw + 6}" y="${cy0 + 44 + i * ch}" width="${cw - 12}" height="${ch - 8}" rx="8" fill="#ffffff"/>\n`;
  });
  const ley = [
    ['MT', 'mismo tipo'],
    ['EF', 'automática'],
    ['RC', 'requiere convertir'],
    ['SE', 'error de compilación'],
  ];
  ley.forEach(([k, txt], i) => {
    const x = 30 + i * 215;
    const [bg] = celda[k];
    s += `    <g class="pop" style="animation-delay:${(3 + i * 0.12).toFixed(2)}s"><rect x="${x}" y="${H - 50}" width="44" height="26" rx="6" fill="${bg}"/><text x="${x + 22}" y="${H - 32}" font-size="13" font-weight="700" fill="#fff" text-anchor="middle">${k}</text><text x="${x + 54}" y="${H - 32}" font-size="14" fill="#b8bde8">${txt}</text></g>\n`;
  });
  s += '  </g>\n</svg>\n';
  writeFileSync(join(OUT, 'efectividades.svg'), s);
}

// 6. Divisor: Pokébola que rueda
{
  const s = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 900 40" width="900" height="40" role="img" aria-label="Separador con una Pokébola rodando">
  <style>
    .rodar { animation: rodar 7s cubic-bezier(.45,.05,.55,.95) infinite alternate; }
    @keyframes rodar { from { transform: translateX(0); } to { transform: translateX(840px); } }
    .giro { transform-box: fill-box; transform-origin: center; animation: giro 7s cubic-bezier(.45,.05,.55,.95) infinite alternate; }
    @keyframes giro { from { transform: rotate(0deg); } to { transform: rotate(1080deg); } }
  </style>
  <line x1="10" y1="20" x2="890" y2="20" stroke="#8b90c0" stroke-width="2" stroke-dasharray="2 10" stroke-linecap="round" opacity="0.6"/>
  <g class="rodar">
    <g class="giro">
      <circle cx="30" cy="20" r="14" fill="#ffffff" stroke="#1b1b1b" stroke-width="2.5"/>
      <path d="M16 20 a14 14 0 0 1 28 0 z" fill="#ee1515" stroke="#1b1b1b" stroke-width="2.5"/>
      <circle cx="30" cy="20" r="5" fill="#ffffff" stroke="#1b1b1b" stroke-width="2.5"/>
    </g>
  </g>
</svg>
`;
  writeFileSync(join(OUT, 'divisor.svg'), s);
}
console.log('listo');

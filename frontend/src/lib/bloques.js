// Estructura de bloques de un archivo .pks, leída línea por línea: qué
// líneas abren un bloque (combate, movimiento, si, mientras, recorrer, segun,
// especie, ficha), cuáles lo continúan (sino, sino si) y cuáles lo cierran
// (fin). No usa CodeMirror, así que se puede probar sola. La usa el editor
// para la sangría automática, las líneas guía, el plegado y la pareja de
// cada bloque.

export const ABREN = new Set([
  'combate',
  'movimiento',
  'si',
  'mientras',
  'recorrer',
  'segun',
  'especie',
  'ficha',
]);
// Una alternativa de segun puede empezar un bloque: «SANO entonces si x».
const ABREN_TRAS_ENTONCES = new Set(['si', 'mientras', 'recorrer', 'segun']);

// palabras devuelve las palabras de una línea fuera de textos y comentarios.
export function palabras(linea) {
  const sin = linea
    .replace(/"(?:[^"\\]|\\.)*"?/g, ' ')
    .replace(/'(?:[^'\\]|\\.)*'?/g, ' ')
    .replace(/\/\/.*$/, '');
  return sin.match(/[\p{L}][\p{L}\p{Nd}_]*/gu) ?? [];
}

// columna cuenta la sangría de una línea; un tabulador vale `tab` columnas.
export function columna(linea, tab = 4) {
  let c = 0;
  for (const ch of linea) {
    if (ch === ' ') c += 1;
    else if (ch === '\t') c += tab - (c % tab);
    else break;
  }
  return c;
}

// clase dice qué hace una línea con los bloques: 'abre', 'medio' (sino y
// sino si), 'cierra' (fin) o null.
export function clase(linea) {
  const ps = palabras(linea);
  if (!ps.length) return null;
  const primera = ps[0];
  if (primera === 'fin') return 'cierra';
  if (primera === 'sino') return 'medio';
  if (ABREN.has(primera)) return 'abre';
  const i = ps.indexOf('entonces');
  if (i >= 0 && ABREN_TRAS_ENTONCES.has(ps[i + 1])) return 'abre';
  return null;
}

// analizar recorre las líneas y devuelve, para cada una, su clase y los
// bloques abiertos que la contienen, y la lista de bloques con sus partes.
// Cada bloque: { abre, medios: [], cierra (o null si no se cerró), col }.
// Un «fin» de más no cierra nada: se ignora.
export function analizar(lineas, tab = 4) {
  const bloques = [];
  const pila = [];
  const info = lineas.map((texto, i) => {
    const c = clase(texto);
    let dentro = pila.slice();
    if (c === 'cierra' || c === 'medio') {
      const b = pila[pila.length - 1];
      if (b !== undefined) {
        dentro = pila.slice(0, -1);
        if (c === 'cierra') {
          bloques[b].cierra = i;
          pila.pop();
        } else {
          bloques[b].medios.push(i);
        }
      }
    } else if (c === 'abre') {
      bloques.push({ abre: i, medios: [], cierra: null, col: columna(texto, tab) });
      pila.push(bloques.length - 1);
    }
    return { clase: c, dentro };
  });
  return { info, bloques };
}

// sangria calcula la sangría (en columnas) que le toca a una línea, dado el
// texto que tiene antes y lo que empieza en ella:
//   - dentro de un bloque: la del que lo abre más `unidad`;
//   - fin, sino y sino si: la del que abre su bloque;
//   - fuera de todo bloque: 0.
export function sangria(lineasAntes, textoLinea, unidad = 4, tab = 4) {
  const { bloques } = analizar(lineasAntes, tab);
  // Los bloques que siguen abiertos al terminar las líneas de antes, del más
  // externo al más interno.
  const abiertos = bloques.filter((b) => b.cierra === null);
  const arriba = abiertos[abiertos.length - 1];
  if (!arriba) return 0;
  const c = clase(textoLinea);
  if (c === 'cierra' || c === 'medio') return arriba.col;
  return arriba.col + unidad;
}

// nivelesGuia dice, para cada línea, en qué columnas van sus líneas guía.
// Como en VS Code, salen de la sangría escrita y no de dónde está la palabra
// que abre el bloque: hay una guía cada `unidad` columnas antes del texto
// de la línea. Así, si una línea que abre un bloque se corre con Tab, las
// guías de su cuerpo no se mueven con ella. Una línea en blanco toma el
// nivel de las que la rodean, para que las guías no se corten.
export function nivelesGuia(lineas, unidad = 4, tab = 4) {
  const nivel = lineas.map((l) => (l.trim() ? Math.floor(columna(l, tab) / unidad) : -1));
  const arriba = [];
  let previo = -1;
  for (const n of nivel) {
    arriba.push(previo);
    if (n >= 0) previo = n;
  }
  const abajo = new Array(nivel.length);
  let siguiente = -1;
  for (let i = nivel.length - 1; i >= 0; i--) {
    abajo[i] = siguiente;
    if (nivel[i] >= 0) siguiente = nivel[i];
  }
  return nivel.map((n, i) => {
    let k = n;
    if (n < 0) {
      const a = arriba[i];
      const b = abajo[i];
      if (a < 0 || b < 0) k = 0;
      else if (a < b) k = a + 1;
      else if (a === b) k = a;
      else k = b + 1;
    }
    return Array.from({ length: k }, (_, j) => j * unidad);
  });
}

// rutaEn devuelve los bloques que contienen la línea `n` (base 0), de afuera
// hacia adentro, con un nombre corto para cada uno: «combate», «si»,
// «movimiento curar», «especie Estado»…
export function rutaEn(lineas, datos, n) {
  const { info, bloques } = datos;
  if (!info[n]) return [];
  const ids = [...info[n].dentro];
  const propio = bloques.findIndex((b) => b.abre === n);
  if (propio >= 0) ids.push(propio);
  return ids.map((i) => nombreBloque(lineas[bloques[i].abre]));
}

function nombreBloque(linea) {
  const ps = palabras(linea);
  let clave = ps[0];
  if (!ABREN.has(clave)) clave = ps[ps.indexOf('entonces') + 1];
  if (clave === 'movimiento') {
    // «movimiento roca curar(…)»: el nombre es la palabra antes del paréntesis.
    const m = /([\p{L}][\p{L}\p{Nd}_]*)\s*\(/u.exec(linea.replace(/\/\/.*$/, ''));
    return m ? `movimiento ${m[1]}` : 'movimiento';
  }
  if ((clave === 'especie' || clave === 'ficha') && ps[1]) return `${clave} ${ps[1]}`;
  return clave;
}

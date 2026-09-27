// Genera el ícono de la app (build/appicon.png, 1024x1024): la Pokéball del
// juego ampliada sin suavizar, sobre un cuadro redondeado. Wails arma el .ico
// de Windows a partir de este PNG.
// Uso: node scripts/icono-app.mjs
import { readFileSync, writeFileSync, rmSync } from 'node:fs';
import { inflateSync, deflateSync } from 'node:zlib';
import { fileURLToPath } from 'node:url';

const raiz = (r) => fileURLToPath(new URL(`../${r}`, import.meta.url));

// leerPng decodifica un PNG de paleta o RGBA sin entrelazar.
function leerPng(buf) {
  let pos = 8;
  const png = { paleta: [], alfa: [], datos: [] };
  while (pos < buf.length) {
    const largo = buf.readUInt32BE(pos);
    const tipo = buf.toString('ascii', pos + 4, pos + 8);
    const d = buf.subarray(pos + 8, pos + 8 + largo);
    if (tipo === 'IHDR')
      Object.assign(png, {
        ancho: d.readUInt32BE(0),
        alto: d.readUInt32BE(4),
        prof: d[8],
        color: d[9],
      });
    if (tipo === 'PLTE')
      for (let i = 0; i < d.length; i += 3) png.paleta.push([d[i], d[i + 1], d[i + 2]]);
    if (tipo === 'tRNS') png.alfa = [...d];
    if (tipo === 'IDAT') png.datos.push(d);
    pos += 12 + largo;
  }
  const canales = { 3: 1, 6: 4, 2: 3 }[png.color];
  const bits = canales * png.prof;
  const bpp = Math.max(1, bits >> 3);
  const fila = Math.ceil((png.ancho * bits) / 8);
  const crudo = inflateSync(Buffer.concat(png.datos));
  const px = Buffer.alloc(fila * png.alto);
  for (let y = 0; y < png.alto; y++) {
    const f = crudo[y * (fila + 1)];
    for (let i = 0; i < fila; i++) {
      const v = crudo[y * (fila + 1) + 1 + i];
      const a = i >= bpp ? px[y * fila + i - bpp] : 0;
      const b = y ? px[(y - 1) * fila + i] : 0;
      const c = y && i >= bpp ? px[(y - 1) * fila + i - bpp] : 0;
      const p = a + b - c;
      const pred = [
        0,
        a,
        b,
        (a + b) >> 1,
        Math.abs(p - a) <= Math.abs(p - b) && Math.abs(p - a) <= Math.abs(p - c)
          ? a
          : Math.abs(p - b) <= Math.abs(p - c)
            ? b
            : c,
      ][f];
      px[y * fila + i] = (v + pred) & 255;
    }
  }
  return (x, y) => {
    if (png.color === 3) {
      const bit = x * png.prof;
      const idx = (px[y * fila + (bit >> 3)] >> (8 - png.prof - (bit & 7))) & ((1 << png.prof) - 1);
      return [...png.paleta[idx], png.alfa[idx] ?? 255];
    }
    const o = y * fila + x * canales;
    return [px[o], px[o + 1], px[o + 2], canales === 4 ? px[o + 3] : 255];
  };
}

// escribirPng codifica una imagen RGBA.
function escribirPng(ancho, alto, rgba) {
  const crc = (b) => {
    let c = ~0;
    for (const x of b) {
      c ^= x;
      for (let k = 0; k < 8; k++) c = c & 1 ? (c >>> 1) ^ 0xedb88320 : c >>> 1;
    }
    return ~c >>> 0;
  };
  const trozo = (tipo, d) => {
    const t = Buffer.concat([Buffer.from(tipo), d]);
    const l = Buffer.alloc(4);
    l.writeUInt32BE(d.length);
    const c = Buffer.alloc(4);
    c.writeUInt32BE(crc(t));
    return Buffer.concat([l, t, c]);
  };
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(ancho, 0);
  ihdr.writeUInt32BE(alto, 4);
  ihdr[8] = 8;
  ihdr[9] = 6;
  const filas = Buffer.alloc((ancho * 4 + 1) * alto);
  for (let y = 0; y < alto; y++)
    rgba.copy(filas, y * (ancho * 4 + 1) + 1, y * ancho * 4, (y + 1) * ancho * 4);
  return Buffer.concat([
    Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]),
    trozo('IHDR', ihdr),
    trozo('IDAT', deflateSync(filas)),
    trozo('IEND', Buffer.alloc(0)),
  ]);
}

const LADO = 1024;
const pixel = leerPng(readFileSync(raiz('frontend/public/objetos/poke-ball.png')));
const img = Buffer.alloc(LADO * LADO * 4);
const pintar = (x, y, [r, g, b, a]) => {
  const o = (y * LADO + x) * 4;
  img[o] = r;
  img[o + 1] = g;
  img[o + 2] = b;
  img[o + 3] = a;
};

// Fondo: cuadro redondeado azul noche con borde.
const radio = 180;
for (let y = 0; y < LADO; y++) {
  for (let x = 0; x < LADO; x++) {
    const dx = Math.max(radio - x, 0, x - (LADO - 1 - radio));
    const dy = Math.max(radio - y, 0, y - (LADO - 1 - radio));
    const d = Math.hypot(dx, dy);
    if (d > radio) continue;
    const borde = d > radio - 36 || x < 36 || y < 36 || x > LADO - 37 || y > LADO - 37;
    pintar(
      x,
      y,
      borde ? [20, 20, 32, 255] : y < LADO / 2 ? [42, 79, 160, 255] : [30, 58, 122, 255],
    );
  }
}

// Pokéball: recortada a lo que tiene color y ampliada sin suavizar, centrada.
let [x0, y0, x1, y1] = [30, 30, 0, 0];
for (let y = 0; y < 30; y++)
  for (let x = 0; x < 30; x++)
    if (pixel(x, y)[3] >= 128)
      [x0, y0, x1, y1] = [Math.min(x0, x), Math.min(y0, y), Math.max(x1, x), Math.max(y1, y)];
const escala = Math.floor(760 / Math.max(x1 - x0 + 1, y1 - y0 + 1));
const ox = Math.floor((LADO - (x1 - x0 + 1) * escala) / 2);
const oy = Math.floor((LADO - (y1 - y0 + 1) * escala) / 2);
for (let y = y0; y <= y1; y++) {
  for (let x = x0; x <= x1; x++) {
    const c = pixel(x, y);
    if (c[3] < 128) continue;
    for (let j = 0; j < escala; j++)
      for (let i = 0; i < escala; i++)
        pintar(ox + (x - x0) * escala + i, oy + (y - y0) * escala + j, c);
  }
}

writeFileSync(raiz('build/appicon.png'), escribirPng(LADO, LADO, img));
// Wails vuelve a generar el .ico desde el PNG en el siguiente build.
rmSync(raiz('build/windows/icon.ico'), { force: true });
console.log('build/appicon.png listo');

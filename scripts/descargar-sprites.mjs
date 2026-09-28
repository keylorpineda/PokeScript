// Descarga de PokeAPI el arte que usan las escenas del README: los Pokémon
// en vectores (dream world) y algunos sprites de los juegos. Se guardan en docs/assets/fuentes/ para que
// generar-escenas.mjs funcione sin red.
// Uso: node scripts/descargar-sprites.mjs
import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const BASE = 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites';
const FUENTES = fileURLToPath(new URL('../docs/assets/fuentes/', import.meta.url));

export const POKEMON = [
  1, 4, 6, 7, 9, 18, 25, 54, 63, 66, 68, 94, 95, 113, 115, 128, 129, 131, 132, 133, 137, 143, 149,
  150, 151, 479,
];
// Sprites de los juegos que se dibujan como pixel art.
export const PIXELES = {
  25: 'versions/generation-v/black-white/25.png',
  '25-hgss': 'versions/generation-iv/heartgold-soulsilver/25.png',
  '74-espalda': 'versions/generation-v/black-white/back/74.png',
};

async function bajar(url, destino) {
  const r = await fetch(url);
  if (!r.ok) throw new Error(`${url}: ${r.status}`);
  writeFileSync(destino, Buffer.from(await r.arrayBuffer()));
}

mkdirSync(join(FUENTES, 'pokemon'), { recursive: true });
for (const id of POKEMON) {
  await bajar(`${BASE}/pokemon/other/dream-world/${id}.svg`, join(FUENTES, 'pokemon', `${id}.svg`));
}
mkdirSync(join(FUENTES, 'pixeles'), { recursive: true });
for (const [nombre, ruta] of Object.entries(PIXELES)) {
  await bajar(`${BASE}/pokemon/${ruta}`, join(FUENTES, 'pixeles', `${nombre}.png`));
}
console.log(`listo: ${POKEMON.length} Pokémon`);

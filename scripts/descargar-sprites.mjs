// Descarga de PokeAPI el arte que usan las escenas del README: los Pokémon
// en vectores (dream world) y las ocho medallas de Kanto. Se guardan en
// docs/assets/fuentes/ para que generar-escenas.mjs funcione sin red.
// Uso: node scripts/descargar-sprites.mjs
import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const BASE = 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites';
const FUENTES = fileURLToPath(new URL('../docs/assets/fuentes/', import.meta.url));

export const POKEMON = [
  1, 4, 6, 7, 9, 25, 54, 63, 66, 68, 94, 95, 113, 115, 128, 129, 131, 132, 133, 137, 143, 149, 479,
];
export const MEDALLAS = [1, 2, 3, 4, 5, 6, 7, 8];

async function bajar(url, destino) {
  const r = await fetch(url);
  if (!r.ok) throw new Error(`${url}: ${r.status}`);
  writeFileSync(destino, Buffer.from(await r.arrayBuffer()));
}

mkdirSync(join(FUENTES, 'pokemon'), { recursive: true });
mkdirSync(join(FUENTES, 'medallas'), { recursive: true });
for (const id of POKEMON) {
  await bajar(`${BASE}/pokemon/other/dream-world/${id}.svg`, join(FUENTES, 'pokemon', `${id}.svg`));
}
for (const n of MEDALLAS) {
  await bajar(`${BASE}/badges/${n}.png`, join(FUENTES, 'medallas', `${n}.png`));
}
console.log(`listo: ${POKEMON.length} Pokémon y ${MEDALLAS.length} medallas`);

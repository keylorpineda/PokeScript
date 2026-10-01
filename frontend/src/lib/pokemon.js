// Pokémon del IDE: sprites animados de Negro y Blanco (vía Pokémon
// Showdown) y los objetos de PokeAPI.

export const POKEMON = {
  bulbasaur: { nombre: 'Bulbasaur', tipo: 'planta', color: '#78c850' },
  charmander: { nombre: 'Charmander', tipo: 'fuego', color: '#f08030' },
  squirtle: { nombre: 'Squirtle', tipo: 'agua', color: '#6890f0' },
  pikachu: { nombre: 'Pikachu', tipo: 'electrico', color: '#f8d030' },
  charizard: { nombre: 'Charizard', tipo: 'fuego', color: '#f08030' },
  eevee: { nombre: 'Eevee', tipo: 'normal', color: '#c8a070' },
  gengar: { nombre: 'Gengar', tipo: 'fantasma', color: '#705898' },
  mewtwo: { nombre: 'Mewtwo', tipo: 'psíquico', color: '#a060d8' },
  mew: { nombre: 'Mew', tipo: 'psíquico', color: '#f0a0d0' },
  psyduck: { nombre: 'Psyduck', tipo: 'agua', color: '#6890f0' },
  magikarp: { nombre: 'Magikarp', tipo: 'agua', color: '#6890f0' },
  abra: { nombre: 'Abra', tipo: 'psíquico', color: '#f85888' },
  chansey: { nombre: 'Chansey', tipo: 'normal', color: '#f0a0c0' },
  snorlax: { nombre: 'Snorlax', tipo: 'normal', color: '#5a7890' },
  porygon: { nombre: 'Porygon', tipo: 'normal', color: '#58b8d0' },
  lapras: { nombre: 'Lapras', tipo: 'agua', color: '#6890f0' },
  geodude: { nombre: 'Geodude', tipo: 'roca', color: '#b8a038' },
  onix: { nombre: 'Onix', tipo: 'roca', color: '#b8a038' },
  kangaskhan: { nombre: 'Kangaskhan', tipo: 'normal', color: '#a8a878' },
  unown: { nombre: 'Unown', tipo: 'psíquico', color: '#f85888' },
  ditto: { nombre: 'Ditto', tipo: 'normal', color: '#a890f0' },
  machamp: { nombre: 'Machamp', tipo: 'lucha', color: '#c03028' },
  tauros: { nombre: 'Tauros', tipo: 'normal', color: '#a8a878' },
  alakazam: { nombre: 'Alakazam', tipo: 'psíquico', color: '#f85888' },
  slowpoke: { nombre: 'Slowpoke', tipo: 'agua', color: '#6890f0' },
  // Evoluciones de los iniciales.
  ivysaur: { nombre: 'Ivysaur', tipo: 'planta', color: '#78c850' },
  venusaur: { nombre: 'Venusaur', tipo: 'planta', color: '#78c850' },
  charmeleon: { nombre: 'Charmeleon', tipo: 'fuego', color: '#f08030' },
  wartortle: { nombre: 'Wartortle', tipo: 'agua', color: '#6890f0' },
  blastoise: { nombre: 'Blastoise', tipo: 'agua', color: '#6890f0' },
  raichu: { nombre: 'Raichu', tipo: 'electrico', color: '#f8d030' },
};

// Los que Oak ofrece en su laboratorio.
export const INICIALES = ['bulbasaur', 'charmander', 'squirtle', 'pikachu'];

// El compañero sube un nivel cada 3 análisis o combates que compilan.
export const nivelDe = (exp) => 5 + Math.floor((exp ?? 0) / 3);

// Niveles de evolución de cada inicial: los de los juegos para las tres
// líneas de Kanto. Pikachu evoluciona con una piedra en los juegos; aquí,
// al nivel 26.
export const EVOLUCIONES = {
  bulbasaur: [
    [16, 'ivysaur'],
    [32, 'venusaur'],
  ],
  charmander: [
    [16, 'charmeleon'],
    [36, 'charizard'],
  ],
  squirtle: [
    [16, 'wartortle'],
    [36, 'blastoise'],
  ],
  pikachu: [[26, 'raichu']],
};

// forma devuelve en qué se convirtió el inicial al nivel dado.
export function forma(inicial, nivel) {
  let actual = inicial ?? 'pikachu';
  for (const [n, siguiente] of EVOLUCIONES[inicial] ?? []) {
    if (nivel >= n) actual = siguiente;
  }
  return actual;
}

export const sprite = (id) => `/sprites/ani/${id}.gif`;
export const objeto = (id) => `/objetos/${id}.png`;
export const OAK = '/sprites/oak.png';

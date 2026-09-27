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
};

// Los que Oak ofrece en su laboratorio.
export const INICIALES = ['bulbasaur', 'charmander', 'squirtle', 'pikachu'];

export const sprite = (id) => `/sprites/ani/${id}.gif`;
export const objeto = (id) => `/objetos/${id}.png`;
export const OAK = '/sprites/oak.png';

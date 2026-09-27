// Portadas de la pantalla de título: un lugar, su Pokémon y lo que flota en
// el aire. La pantalla de título las va rotando.

export const PORTADAS = [
  { fondo: 'volcanocave', mascota: 'charizard', particulas: 'brasas', vuela: true },
  { fondo: 'route', mascota: 'pikachu', particulas: 'hojas' },
  {
    fondo: 'forest',
    mascota: 'gengar',
    particulas: 'luciernagas',
    filtro: 'hue-rotate(150deg) brightness(0.5) saturate(1.4)',
  },
  {
    fondo: 'icecave',
    mascota: 'mewtwo',
    particulas: 'nieve',
    filtro: 'brightness(0.62) saturate(1.2) hue-rotate(15deg)',
  },
  { fondo: 'meadow', mascota: 'bulbasaur', particulas: 'petalos' },
  { fondo: 'river', mascota: 'squirtle', particulas: 'burbujas' },
  { fondo: 'thunderplains', mascota: 'eevee', particulas: 'chispas' },
  { fondo: 'deepsea', mascota: 'lapras', particulas: 'burbujas' },
  { fondo: 'mountain', mascota: 'snorlax', particulas: 'hojas' },
];

// Empieza por una distinta cada vez que se abre el IDE.
export function portadaInicial() {
  return Math.floor(Math.random() * PORTADAS.length);
}

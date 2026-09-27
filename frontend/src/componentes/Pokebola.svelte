<script>
  // Pokéball: el sprite del objeto de los juegos (30x30), ampliado sin
  // suavizar.
  import { objeto } from '../lib/pokemon.js';

  let { escala = 3, tipo = 'poke-ball', clase = '' } = $props();
</script>

<img
  class="pixel {clase}"
  src={objeto(tipo)}
  alt=""
  width={30 * escala}
  height={30 * escala}
  draggable="false"
/>

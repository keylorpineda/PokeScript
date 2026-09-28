<script>
  // La Pokédex roja de los primeros juegos, dibujada en una cuadrícula de
  // 16x16: lente azul, tres luces, la línea de la tapa y la pantallita.
  let { tam = 26 } = $props();

  const B = '#3a0a10'; // borde
  const R = '#e3223a'; // rojo
  const O = '#a8101f'; // rojo oscuro
  const W = '#f4f4f4'; // aro del lente
  const A = '#3d9cf0'; // azul
  const C = '#b8e4ff'; // brillo
  const S = '#1f2a24'; // marco de la pantalla
  const V = '#7fd8a0'; // pantalla

  // [x, y, ancho, alto, color]
  const PIXELES = [
    [1, 0, 14, 1, B],
    [0, 1, 1, 14, B],
    [15, 1, 1, 14, B],
    [1, 15, 14, 1, B],
    [1, 1, 14, 14, R],
    [1, 13, 14, 2, O],
    // Lente
    [4, 2, 3, 1, W],
    [3, 3, 5, 3, W],
    [4, 6, 3, 1, W],
    [4, 3, 3, 3, A],
    [4, 3, 1, 1, C],
    [3, 2, 1, 1, B],
    [7, 2, 1, 1, B],
    [3, 6, 1, 1, B],
    [7, 6, 1, 1, B],
    // Luces
    [9, 3, 1, 1, '#ff5f57'],
    [11, 3, 1, 1, '#ffcb05'],
    [13, 3, 1, 1, '#5fd35f'],
    // Línea escalonada de la tapa
    [1, 8, 8, 1, B],
    [9, 7, 1, 1, B],
    [10, 7, 5, 1, B],
    [9, 8, 1, 1, B],
    // Pantalla
    [3, 10, 10, 3, S],
    [4, 11, 8, 1, V],
  ];
</script>

<svg
  width={tam}
  height={tam}
  viewBox="0 0 16 16"
  shape-rendering="crispEdges"
  aria-hidden="true"
  class="icono-pokedex"
>
  {#each PIXELES as [x, y, w, h, c], i (i)}<rect {x} {y} width={w} height={h} fill={c} />{/each}
</svg>

<style>
  .icono-pokedex {
    display: block;
    flex: none;
    image-rendering: pixelated;
  }
</style>

<script>
  // La mochila del entrenador, como la de Rojo Fuego, dibujada en una
  // cuadrícula de 16x16. Cambia con el tema: sus colores y el emblema de la
  // solapa (llama, rayo, hoja, gota…).
  import { perfil } from '../lib/estado.svelte.js';

  let { tam = 26 } = $props();

  // borde, cuero, sombra, solapa, emblema
  const COLORES = {
    'rojo-fuego': ['#3a1010', '#d8483a', '#9a2a20', '#f07050', '#ffd84a'],
    'game-boy': ['#0f380f', '#306230', '#0f380f', '#8bac0f', '#c4d86a'],
    'torre-lavanda': ['#1a0f2a', '#7a4ab0', '#4a2a78', '#a878e0', '#e8d8ff'],
    'cueva-celeste': ['#0a1a30', '#3a78c0', '#1f4f8a', '#78b8f0', '#e8f8ff'],
    'pueblo-paleta': ['#3a2410', '#e0a040', '#b06a20', '#f4c660', '#5ab84a'],
    'mar-profundo': ['#04202a', '#1f8aa0', '#0f5a6e', '#4ac0d0', '#b8f8ff'],
    'llanura-trueno': ['#3a3010', '#f2c200', '#b08a00', '#ffe060', '#3a3010'],
    'monte-plateado': ['#26221e', '#8a8478', '#5e584e', '#b8b2a4', '#e0a040'],
  };

  // Emblema de 3x3 sobre la solapa, uno por tema.
  const EMBLEMAS = {
    'rojo-fuego': [
      [1, 0],
      [0, 1],
      [1, 1],
      [2, 1],
      [1, 2],
    ], // llama
    'game-boy': [
      [0, 0],
      [1, 0],
      [2, 0],
      [0, 1],
      [2, 1],
      [0, 2],
      [1, 2],
      [2, 2],
    ], // hebilla
    'torre-lavanda': [
      [0, 0],
      [2, 0],
      [0, 1],
      [1, 1],
      [2, 1],
      [0, 2],
      [2, 2],
    ], // fantasma
    'cueva-celeste': [
      [1, 0],
      [0, 1],
      [1, 1],
      [2, 1],
      [1, 2],
    ], // cristal
    'pueblo-paleta': [
      [2, 0],
      [1, 1],
      [2, 1],
      [0, 2],
      [1, 2],
    ], // hoja
    'mar-profundo': [
      [1, 0],
      [0, 1],
      [1, 1],
      [2, 1],
      [1, 2],
      [0, 2],
      [2, 2],
    ], // gota
    'llanura-trueno': [
      [2, 0],
      [1, 0],
      [1, 1],
      [0, 2],
      [1, 2],
    ], // rayo
    'monte-plateado': [
      [0, 0],
      [1, 0],
      [2, 0],
      [0, 1],
      [1, 1],
      [2, 1],
      [1, 2],
    ], // piedra
  };

  const c = $derived(COLORES[perfil.tema] ?? COLORES['pueblo-paleta']);
  const emblema = $derived(EMBLEMAS[perfil.tema] ?? EMBLEMAS['game-boy']);

  // [x, y, ancho, alto, índice de color]
  const FORMA = [
    // Asa
    [6, 0, 4, 1, 0],
    [5, 1, 1, 2, 0],
    [10, 1, 1, 2, 0],
    // Contorno
    [3, 3, 10, 1, 0],
    [2, 4, 1, 10, 0],
    [13, 4, 1, 10, 0],
    [3, 14, 10, 1, 0],
    // Cuerpo
    [3, 4, 10, 10, 1],
    [3, 9, 1, 5, 2],
    [12, 9, 1, 5, 2],
    [3, 13, 10, 1, 2],
    // Solapa
    [3, 4, 10, 4, 3],
    [3, 8, 10, 1, 2],
    // Costuras de los bolsillos
    [5, 10, 1, 3, 2],
    [10, 10, 1, 3, 2],
  ];
</script>

<svg
  width={tam}
  height={tam}
  viewBox="0 0 16 16"
  shape-rendering="crispEdges"
  aria-hidden="true"
  class="icono-mochila"
>
  {#each FORMA as [x, y, w, h, i], k (k)}<rect {x} {y} width={w} height={h} fill={c[i]} />{/each}
  {#each emblema as [x, y], k (k)}<rect
      x={7 + x - 1}
      y={5 + y}
      width="1"
      height="1"
      fill={c[4]}
    />{/each}
</svg>

<style>
  .icono-mochila {
    display: block;
    flex: none;
    image-rendering: pixelated;
  }
</style>

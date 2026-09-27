<script>
  // Sombra de pixel art bajo un Pokémon: una elipse de pocos píxeles con los
  // bordes escalonados (como en los juegos), en dos tonos, ampliada sin
  // suavizar. «ancho» es el tamaño en pantalla.
  let { ancho = 160, color = '#000000', clase = '' } = $props();

  const W = 32;
  const H = 8;

  // filas arma los rectángulos de una elipse de ancho w y alto h, centrada.
  function filas(w, h) {
    const rects = [];
    for (let y = 0; y < h; y++) {
      const dy = (y + 0.5 - h / 2) / (h / 2);
      const medio = Math.round((w / 2) * Math.sqrt(Math.max(0, 1 - dy * dy)));
      if (medio > 0) rects.push({ x: W / 2 - medio, y: (H - h) / 2 + y, w: medio * 2 });
    }
    return rects;
  }

  const afuera = filas(W, H);
  const adentro = filas(20, 4);
</script>

<svg
  class="sombra {clase}"
  width={ancho}
  height={(ancho * H) / W}
  viewBox="0 0 {W} {H}"
  shape-rendering="crispEdges"
  aria-hidden="true"
>
  {#each afuera as r, i (i)}<rect
      x={r.x}
      y={r.y}
      width={r.w}
      height="1"
      fill={color}
      opacity="0.2"
    />{/each}
  {#each adentro as r, i (i)}<rect
      x={r.x}
      y={r.y}
      width={r.w}
      height="1"
      fill={color}
      opacity="0.22"
    />{/each}
</svg>

<style>
  .sombra {
    display: block;
    image-rendering: pixelated;
  }
</style>

<script>
  // Sombra proyectada: la silueta del mismo sprite que tiene al lado,
  // oscura, aplastada e inclinada sobre el piso. Como es el mismo GIF, la
  // sombra se mueve con el Pokémon. Copia la imagen y el zoom del sprite
  // hermano, así sirve en cualquier pantalla y tamaño de ventana.
  import { onMount } from 'svelte';

  let el = $state();
  let src = $state('');
  let zoom = $state(1);
  let abajo = $state(0);

  function copiar() {
    const padre = el?.parentElement;
    const sprite = padre?.querySelector(':scope > img:not(.silueta)');
    if (!sprite) return;
    src = sprite.currentSrc || sprite.src;
    zoom = parseFloat(getComputedStyle(sprite).zoom) || 1;
    abajo = padre.getBoundingClientRect().bottom - sprite.getBoundingClientRect().bottom;
  }

  onMount(() => {
    copiar();
    const padre = el.parentElement;
    const tamano = new ResizeObserver(copiar);
    tamano.observe(padre);
    const cambios = new MutationObserver(copiar);
    cambios.observe(padre, { subtree: true, attributes: true, attributeFilter: ['src', 'class'] });
    return () => {
      tamano.disconnect();
      cambios.disconnect();
    };
  });
</script>

<img
  bind:this={el}
  class="silueta"
  src={src || undefined}
  alt=""
  aria-hidden="true"
  draggable="false"
  style="zoom:{zoom};bottom:{abajo / zoom}px"
/>

<style>
  .silueta {
    position: absolute;
    left: 0;
    right: 0;
    margin: 0 auto;
    transform-origin: 50% 100%;
    transform: scaleY(0.26) skewX(-42deg);
    filter: brightness(0) blur(0.7px);
    opacity: 0.32;
    image-rendering: pixelated;
    pointer-events: none;
  }
</style>

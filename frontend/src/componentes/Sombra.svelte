<script>
  // Sombra bajo un Pokémon, como en los combates de Negro y Blanco: una
  // elipse suave, más oscura en el centro y difuminada hacia los bordes.
  // «ancho» es el tamaño en pantalla; conviene que sea un poco menos que el
  // ancho del Pokémon.
  let { ancho = 120 } = $props();
</script>

<span class="sombra" style="width:{ancho}px;height:{Math.round(ancho * 0.24)}px"></span>

<style>
  .sombra {
    display: block;
    border-radius: 50%;
    background: radial-gradient(
      ellipse at center,
      rgba(0, 0, 0, 0.5) 0%,
      rgba(0, 0, 0, 0.36) 35%,
      rgba(0, 0, 0, 0.14) 58%,
      rgba(0, 0, 0, 0) 72%
    );
    filter: blur(1.5px);
    pointer-events: none;
  }
</style>

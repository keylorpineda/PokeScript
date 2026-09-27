<script>
  // Transición de inicio de combate: destello y franjas negras que cierran y
  // abren la pantalla, como cuando aparece un Pokémon salvaje.
  import { onMount } from 'svelte';

  let { alTerminar } = $props();
  onMount(() => {
    const id = setTimeout(alTerminar, 1300);
    return () => clearTimeout(id);
  });
</script>

<div class="transicion">
  <div class="destello"></div>
  {#each Array.from({ length: 10 }, (_, i) => i) as i (i)}
    <span class="franja" class:par={i % 2 === 0} style="top:{i * 10}%"></span>
  {/each}
  <p class="aviso">¡COMIENZA EL COMBATE!</p>
</div>

<style>
  .transicion {
    position: fixed;
    inset: 0;
    z-index: 50;
    pointer-events: none;
  }
  .destello {
    position: absolute;
    inset: 0;
    background: #fff;
    animation: destellar 0.4s steps(4) forwards;
  }
  @keyframes destellar {
    0%,
    50% {
      opacity: 0.9;
    }
    25%,
    75% {
      opacity: 0;
    }
    100% {
      opacity: 0;
    }
  }
  .franja {
    position: absolute;
    left: 0;
    width: 100%;
    height: 10.2%;
    background: #101018;
    animation: cerrar-der 1.3s steps(10) 0.3s both;
  }
  .franja.par {
    animation-name: cerrar-izq;
  }
  @keyframes cerrar-izq {
    0% {
      transform: translateX(-100%);
    }
    45%,
    60% {
      transform: translateX(0);
    }
    100% {
      transform: translateX(100%);
    }
  }
  @keyframes cerrar-der {
    0% {
      transform: translateX(100%);
    }
    45%,
    60% {
      transform: translateX(0);
    }
    100% {
      transform: translateX(-100%);
    }
  }
  .aviso {
    position: absolute;
    left: 0;
    right: 0;
    top: 45%;
    margin: 0;
    text-align: center;
    font-family: var(--titulo);
    font-size: 44px;
    font-weight: 700;
    letter-spacing: 3px;
    color: #ffcb05;
    text-shadow: 0 4px 0 #2a4fa0;
    opacity: 0;
    animation: mostrar 1.3s steps(1) 0.3s both;
  }
  @keyframes mostrar {
    45% {
      opacity: 1;
    }
    62% {
      opacity: 0;
    }
  }
</style>

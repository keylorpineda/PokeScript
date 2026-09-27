<script>
  // Pantalla de carga al abrir un proyecto: la Pokéball se sacude tres veces
  // y hace clic, como cuando se atrapa a un Pokémon.
  import { onMount } from 'svelte';
  import Escena from './Escena.svelte';
  import Pokebola from './Pokebola.svelte';
  import { sonar } from '../lib/sonido.js';

  let { proyecto = 'proyecto', alTerminar } = $props();
  let capturado = $state(false);

  onMount(() => {
    sonar('captura');
    const a = setTimeout(() => (capturado = true), 1700);
    const b = setTimeout(alTerminar, 2800);
    return () => {
      clearTimeout(a);
      clearTimeout(b);
    };
  });
</script>

<Escena oscurecer={0.55} paneo={false}>
  <div class="centro">
    <div class="bola" class:capturado>
      <Pokebola escala={4} />
      {#if capturado}
        {#each [0, 1, 2, 3, 4, 5, 6, 7] as i (i)}
          <span class="estrella" style="--a:{i * 45}deg"></span>
        {/each}
      {/if}
    </div>
    <div class="aviso marco">
      {#if capturado}
        ¡Ya está! ¡{proyecto} fue capturado!
      {:else}
        Abriendo {proyecto}<span class="puntos"></span>
      {/if}
    </div>
  </div>
</Escena>

<style>
  .centro {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 40px;
  }
  .bola {
    position: relative;
    transform-origin: 50% 90%;
    animation: sacudir 1.7s steps(1) both;
  }
  @keyframes sacudir {
    0%,
    26%,
    52%,
    78%,
    100% {
      transform: rotate(0);
    }
    8%,
    34%,
    60% {
      transform: rotate(-22deg);
    }
    16%,
    42%,
    68% {
      transform: rotate(22deg);
    }
  }
  .capturado {
    animation: none;
  }
  .estrella {
    position: absolute;
    left: 50%;
    top: 50%;
    width: 18px;
    height: 18px;
    margin: -9px;
    background: #ffe14a;
    clip-path: polygon(50% 0, 62% 38%, 100% 50%, 62% 62%, 50% 100%, 38% 62%, 0 50%, 38% 38%);
    animation: estallar 0.7s steps(6) forwards;
  }
  @keyframes estallar {
    from {
      transform: rotate(var(--a)) translateY(0) scale(0.4);
    }
    to {
      transform: rotate(var(--a)) translateY(-100px) scale(1.2);
      opacity: 0;
    }
  }
  .aviso {
    padding: 4px 18px;
    font-family: var(--titulo);
    font-size: 26px;
    font-weight: 600;
    min-width: 420px;
    text-align: center;
  }
  .puntos::after {
    content: '';
    animation: puntos 1.2s steps(4) infinite;
  }
  @keyframes puntos {
    0% {
      content: '';
    }
    25% {
      content: '.';
    }
    50% {
      content: '..';
    }
    75% {
      content: '...';
    }
  }
</style>

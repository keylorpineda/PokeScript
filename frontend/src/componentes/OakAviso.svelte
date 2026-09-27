<script>
  // El Profesor Oak entra desde la izquierda cuando hay algo importante:
  // logros, o su famoso «¡No es el momento de usar eso!».
  import { untrack } from 'svelte';
  import Dialogo from './Dialogo.svelte';
  import { sonar } from '../lib/sonido.js';
  import { ide } from '../lib/estado.svelte.js';
  import { OAK } from '../lib/pokemon.js';

  let i = $state(0);

  // Su campanita al aparecer.
  $effect(() => {
    if (ide.oak) untrack(() => sonar('oak'));
  });
  const lineas = $derived(ide.oak?.lineas ?? []);

  function avanzar() {
    if (i < lineas.length - 1) i += 1;
    else {
      ide.oak = null;
      i = 0;
    }
  }
</script>

{#if ide.oak}
  <div class="velo"></div>
  <div class="oak">
    <img class="pixel" src={OAK} alt="Profesor Oak" draggable="false" />
    <div class="caja">
      {#key i}
        <Dialogo texto={lineas[i]} nombre="Prof. Oak" alAvanzar={avanzar} />
      {/key}
    </div>
  </div>
{/if}

<style>
  .velo {
    position: fixed;
    inset: 0;
    z-index: 40;
    background: rgba(10, 8, 20, 0.45);
    animation: aparecer 0.3s steps(3) both;
  }
  @keyframes aparecer {
    from {
      opacity: 0;
    }
  }
  .oak {
    position: fixed;
    left: 3vw;
    right: 3vw;
    bottom: 3vh;
    z-index: 41;
    display: flex;
    align-items: flex-end;
    gap: 10px;
  }
  .oak img {
    zoom: 3;
    animation: entrar 0.5s steps(6) both;
  }
  @keyframes entrar {
    from {
      transform: translateX(-60px);
      opacity: 0;
    }
  }
  .caja {
    flex: 1;
    margin-bottom: 8px;
  }
</style>

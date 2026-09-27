<script>
  // Salida del combate: cada gritar aparece como una línea del cuadro de
  // batalla y capturar pide la respuesta con el cursor parpadeando.
  import { tick } from 'svelte';
  import { ide } from '../lib/estado.svelte.js';
  import { responder } from '../lib/acciones.js';

  let caja;
  let entrada = $state('');
  let campo = $state();

  $effect(() => {
    ide.salida.length;
    tick().then(() => caja && (caja.scrollTop = caja.scrollHeight));
  });

  $effect(() => {
    if (ide.esperandoEntrada !== null) tick().then(() => campo?.focus());
  });

  function enviar(e) {
    e.preventDefault();
    responder(entrada);
    entrada = '';
  }
</script>

<section class="salida marco oscuro">
  <header>
    <span class="titulo">COMBATE</span>
    <span class="estado" class:vivo={ide.ejecutando}>
      {ide.ejecutando
        ? ide.esperandoEntrada !== null
          ? 'ESPERANDO RESPUESTA'
          : 'EN CURSO'
        : 'LISTO'}
    </span>
  </header>
  <div class="lineas" bind:this={caja}>
    {#if !ide.salida.length}
      <p class="vacio">
        Presiona ¡COMBATE! (F5) para correr el programa. Lo que diga gritar aparece aquí.
      </p>
    {/if}
    {#each ide.salida as l, i (i)}
      {#if l.tipo === 'pregunta'}
        <!-- La pregunta de capturar y la respuesta van en la misma línea. -->
        <div class="linea pregunta" style="--n:{Math.max(1, [...l.texto].length)}">
          <span>{l.texto}</span>
          {#if l.respuesta !== undefined}
            <span class="respuesta">{l.respuesta}</span>
          {:else if ide.esperandoEntrada !== null}
            <form class="responder" onsubmit={enviar}>
              <input
                bind:this={campo}
                bind:value={entrada}
                spellcheck="false"
                autocomplete="off"
                placeholder="escribe y presiona Enter"
              />
            </form>
          {/if}
        </div>
      {:else}
        <p class="linea {l.tipo}" style="--n:{Math.max(1, [...l.texto].length)}">{l.texto}</p>
      {/if}
    {/each}
  </div>
</section>

<style>
  .salida {
    display: flex;
    flex-direction: column;
    min-height: 0;
    padding: 0 6px 4px;
  }
  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 2px 4px;
    font-family: var(--titulo);
    letter-spacing: 2px;
  }
  .titulo {
    font-size: 16px;
    color: #ffcb05;
  }
  .estado {
    font-size: 13px;
    color: #9a9ac0;
  }
  .estado.vivo {
    color: #6ee7a0;
    animation: parpadeo 1s steps(1) infinite;
  }
  .lineas {
    flex: 1;
    overflow-y: auto;
    font-family: var(--titulo);
    font-size: 19px;
    line-height: 1.45;
    user-select: text;
  }
  p {
    margin: 0;
    white-space: pre-wrap;
  }
  .vacio {
    color: #8a8ab0;
    font-size: 16px;
  }
  .linea {
    animation: escribir calc(var(--n) * 18ms) steps(var(--n)) both;
  }
  @keyframes escribir {
    from {
      clip-path: inset(0 100% 0 0);
    }
    to {
      clip-path: inset(0 0 0 0);
    }
  }
  .texto {
    color: #f4f4ff;
  }
  .pregunta {
    display: flex;
    align-items: baseline;
    color: #7fd8ff;
    white-space: pre;
  }
  .respuesta {
    color: #ffcb05;
  }
  .error {
    color: #ff7a8a;
  }
  .sistema {
    color: #b8b8e0;
    font-style: italic;
  }
  .responder {
    flex: 1;
    display: flex;
  }
  input {
    flex: 1;
    min-width: 120px;
    padding: 0;
    font-family: var(--titulo);
    font-size: 19px;
    color: #ffcb05;
    caret-color: #ffcb05;
    background: color-mix(in srgb, #ffcb05 8%, transparent);
    border: 0;
    border-bottom: 2px dashed #ffcb05;
    outline: none;
    user-select: text;
    animation: resaltar 1.2s steps(2) infinite;
  }
  input::placeholder {
    color: #6a6a8a;
  }
  @keyframes resaltar {
    50% {
      border-bottom-color: transparent;
    }
  }
</style>

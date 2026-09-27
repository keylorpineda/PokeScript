<script>
  // Presentación del Profesor Oak, como al empezar los juegos: se presenta,
  // pregunta el nombre y ofrece un compañero de su laboratorio.
  import Sombra from './Sombra.svelte';
  import Escena from './Escena.svelte';
  import Dialogo from './Dialogo.svelte';
  import Pokebola from './Pokebola.svelte';
  import { perfil, guardarPerfil } from '../lib/estado.svelte.js';
  import { POKEMON, INICIALES, OAK, sprite } from '../lib/pokemon.js';
  import { sonar } from '../lib/sonido.js';

  let { alTerminar } = $props();

  let paso = $state(0);
  let nombre = $state(perfil.nombre || '');
  let elegido = $state(0);
  let confirmando = $state(false);
  let si = $state(true);
  let saliendo = $state(false);

  const inicial = $derived(POKEMON[INICIALES[elegido]]);
  const entrenador = $derived(nombre.trim() || 'Rojo');

  const PASOS = $derived([
    { tipo: 'oak', texto: '¡Hola! ¡Bienvenido al mundo de PokeScript!' },
    { tipo: 'oak', texto: 'Me llamo Oak, pero la gente me llama el Profesor Pokémon.' },
    {
      tipo: 'pokemon',
      texto:
        'Este mundo está habitado por criaturas llamadas Pokémon… y por programas que esperan ser escritos.',
    },
    {
      tipo: 'oak',
      texto:
        'Aquí cada dato tiene un tipo: `roca`, `agua`, `fuego`, `planta`, `electrico`… Y como en los combates, no todos se llevan bien.',
    },
    { tipo: 'nombre', texto: 'Pero primero, cuéntame algo de ti. ¿Cómo te llamas?' },
    {
      tipo: 'elegir',
      texto: `¡Así que te llamas ${entrenador}! En la mesa hay unas Pokéballs. Elige al compañero que te ayudará a programar.`,
    },
    {
      tipo: 'companero',
      texto: `¡${inicial.nombre} parece muy contento contigo! Te acompañará en el editor y te avisará cuando algo no sea muy efectivo.`,
    },
    {
      tipo: 'oak',
      texto: `${entrenador}, ¡tu aventura con PokeScript está por comenzar! ¡Vamos allá!`,
    },
  ]);

  const actual = $derived(PASOS[paso]);

  function siguiente() {
    if (actual.tipo === 'nombre' && !nombre.trim()) return;
    if (actual.tipo === 'elegir') return;
    if (paso === PASOS.length - 1) {
      perfil.nombre = entrenador;
      perfil.companero = INICIALES[elegido];
      guardarPerfil();
      saliendo = true;
      sonar('combate');
      setTimeout(alTerminar, 1000);
      return;
    }
    paso += 1;
  }

  function mover(d) {
    elegido = (elegido + d + INICIALES.length) % INICIALES.length;
    sonar('mover');
  }

  function tecla(e) {
    if (actual.tipo !== 'elegir' || e.repeat) return;
    if (confirmando) {
      if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
        si = !si;
        sonar('mover');
      } else if (e.key === 'Enter') decidir(si);
      else if (e.key === 'Escape') decidir(false);
      return;
    }
    if (e.key === 'ArrowLeft') mover(-1);
    else if (e.key === 'ArrowRight') mover(1);
    else if (e.key === 'Enter') elegir(elegido);
  }

  function elegir(i) {
    elegido = i;
    confirmando = true;
    si = true;
    sonar('elegir');
  }

  function decidir(valor) {
    confirmando = false;
    if (valor) {
      sonar('captura');
      paso += 1;
    } else sonar('mover');
  }
</script>

<svelte:window onkeydown={tecla} />

<div class="intro" class:saliendo>
  <Escena oscurecer={0.25} paneo={false} lugar="pueblo-paleta">
    <div class="escenario">
      {#if actual.tipo === 'elegir'}
        <div class="laboratorio">
          <div class="vitrina">
            {#key elegido}
              <div class="aparece">
                <span class="destello"></span>
                <img
                  class="sprite"
                  src={sprite(INICIALES[elegido])}
                  alt={inicial.nombre}
                  draggable="false"
                />
              </div>
            {/key}
            <div class="ficha marco">
              <span class="nombre-poke">{inicial.nombre}</span>
              <span class="tipo" style="background:{inicial.color}">{inicial.tipo}</span>
            </div>
          </div>
          <div class="mesa">
            {#each INICIALES as id, i (id)}
              <button
                class="hueco"
                class:activo={i === elegido}
                onmouseenter={() => (elegido = i)}
                onclick={() => elegir(i)}
                aria-label={POKEMON[id].nombre}
              >
                <Pokebola escala={2.4} />
              </button>
            {/each}
          </div>
        </div>
      {:else if actual.tipo === 'pokemon'}
        <div class="desfile">
          {#each ['pikachu', 'bulbasaur', 'charmander', 'squirtle', 'eevee'] as id, i (id)}
            <div class="en-fila" style="animation-delay:{i * 0.15}s">
              <img class="sprite" src={sprite(id)} alt="" draggable="false" />
              <Sombra ancho={170} />
            </div>
          {/each}
        </div>
      {:else if actual.tipo === 'companero'}
        <div class="duo">
          <div class="figura">
            <img class="pixel oak" src={OAK} alt="Profesor Oak" draggable="false" />
            <Sombra ancho={250} />
          </div>
          <div class="figura">
            <img
              class="sprite salta"
              src={sprite(INICIALES[elegido])}
              alt={inicial.nombre}
              draggable="false"
            />
            <Sombra ancho={170} />
          </div>
        </div>
      {:else}
        <div class="figura entra">
          <img class="pixel oak" src={OAK} alt="Profesor Oak" draggable="false" />
          <Sombra ancho={250} />
        </div>
      {/if}
    </div>

    <div class="caja">
      {#key paso}
        <Dialogo
          texto={actual.texto}
          nombre="Prof. Oak"
          alAvanzar={actual.tipo === 'elegir' ? null : siguiente}
          teclado={actual.tipo !== 'elegir'}
        >
          {#if actual.tipo === 'nombre'}
            <form
              class="formulario"
              onsubmit={(e) => {
                e.preventDefault();
                siguiente();
              }}
            >
              <!-- svelte-ignore a11y_autofocus -->
              <input
                bind:value={nombre}
                maxlength="12"
                placeholder="Tu nombre"
                autofocus
                onclick={(e) => e.stopPropagation()}
              />
              <button class="boton principal" type="submit" onclick={(e) => e.stopPropagation()}
                >LISTO</button
              >
            </form>
          {/if}
        </Dialogo>
      {/key}

      {#if confirmando}
        <div class="confirmar marco">
          <div class="menu">
            <button class:sel={si} onmouseenter={() => (si = true)} onclick={() => decidir(true)}
              >SÍ</button
            >
            <button class:sel={!si} onmouseenter={() => (si = false)} onclick={() => decidir(false)}
              >NO</button
            >
          </div>
        </div>
        <div class="pregunta marco">
          ¿Eliges a {inicial.nombre}, el Pokémon de tipo {inicial.tipo}?
        </div>
      {/if}
    </div>
  </Escena>
</div>

<style>
  .intro {
    position: fixed;
    inset: 0;
    clip-path: circle(150% at 50% 45%);
  }
  .saliendo {
    animation: cerrar 1s steps(12) forwards;
  }
  @keyframes cerrar {
    to {
      clip-path: circle(0% at 50% 45%);
    }
  }
  .escenario {
    position: absolute;
    inset: 0 0 190px;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    padding-bottom: 20px;
  }
  .figura,
  .en-fila {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .figura img,
  .en-fila img {
    position: relative;
    z-index: 1;
  }
  .entra {
    animation: entrar 0.6s steps(8) both;
  }
  @keyframes entrar {
    from {
      transform: translateX(-45vw);
    }
  }
  .oak {
    zoom: 3.4;
    animation: respirar 1.6s steps(2) infinite;
  }
  @keyframes respirar {
    50% {
      transform: translateY(-1px);
    }
  }
  .desfile,
  .duo {
    display: flex;
    gap: 34px;
    align-items: flex-end;
  }
  .duo {
    gap: 70px;
  }
  .en-fila {
    display: flex;
    flex-direction: column;
    align-items: center;
    animation: saltar-dentro 0.6s steps(6) both;
  }
  .en-fila img,
  .duo .sprite {
    zoom: 2.4;
  }
  @keyframes saltar-dentro {
    from {
      transform: translateY(-50vh);
    }
    70% {
      transform: translateY(10px);
    }
  }
  .salta {
    animation: saltar 0.9s steps(4) infinite;
  }
  @keyframes saltar {
    50% {
      transform: translateY(-6px);
    }
  }
  .laboratorio {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 18px;
  }
  .vitrina {
    display: flex;
    align-items: center;
    gap: 30px;
    min-height: 240px;
  }
  .aparece {
    position: relative;
    display: grid;
    place-items: center;
    min-width: 240px;
  }
  .aparece img {
    zoom: 2.8;
    animation: brotar 0.4s steps(5) both;
  }
  @keyframes brotar {
    from {
      transform: scale(0);
      filter: brightness(10);
    }
    60% {
      filter: brightness(10);
    }
  }
  .destello {
    position: absolute;
    width: 200px;
    height: 200px;
    background: radial-gradient(circle, #fff 0 30%, transparent 70%);
    animation: apagar 0.5s steps(5) forwards;
  }
  @keyframes apagar {
    to {
      opacity: 0;
      transform: scale(1.6);
    }
  }
  .ficha {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 4px 12px;
  }
  .nombre-poke {
    font-family: var(--titulo);
    font-size: 30px;
    font-weight: 700;
  }
  .tipo {
    align-self: flex-start;
    padding: 1px 12px;
    font-family: var(--titulo);
    font-weight: 700;
    font-size: 16px;
    color: #fff;
    text-transform: uppercase;
    border: 3px solid rgba(0, 0, 0, 0.45);
    text-shadow: 2px 2px 0 rgba(0, 0, 0, 0.35);
  }
  .mesa {
    display: flex;
    gap: 22px;
    padding: 18px 38px 26px;
    background: #b77b45;
    border: 4px solid #3e220e;
    box-shadow:
      inset 0 8px 0 #d99c62,
      inset 0 -8px 0 #8a5528,
      0 10px 0 #3e220e;
  }
  .hueco {
    padding: 4px;
    background: none;
    border: 0;
    cursor: pointer;
  }
  .hueco.activo {
    animation: botar 0.5s steps(3) infinite;
  }
  @keyframes botar {
    50% {
      transform: translateY(-10px);
    }
  }
  .caja {
    position: absolute;
    left: 5vw;
    right: 5vw;
    bottom: 3vh;
  }
  .formulario {
    display: flex;
    gap: 12px;
    align-items: center;
  }
  input {
    flex: 0 1 300px;
    padding: 6px 12px;
    font-family: var(--titulo);
    font-size: 24px;
    color: var(--texto);
    background: var(--panel-2);
    border: 3px solid var(--borde);
    outline: none;
    user-select: text;
  }
  .confirmar {
    position: absolute;
    right: 0;
    bottom: calc(100% + 8px);
    padding: 2px 6px;
  }
  .pregunta {
    position: absolute;
    left: 0;
    right: 150px;
    bottom: calc(100% + 8px);
    padding: 4px 10px;
    font-family: var(--titulo);
    font-size: 20px;
  }
  .menu {
    display: flex;
    flex-direction: column;
  }
  .menu button {
    position: relative;
    padding: 2px 10px 2px 26px;
    text-align: left;
    font-family: var(--titulo);
    font-size: 22px;
    font-weight: 700;
    background: none;
    border: 0;
    cursor: pointer;
  }
  .menu button.sel::before {
    content: '';
    position: absolute;
    left: 6px;
    top: 50%;
    border-top: 7px solid transparent;
    border-bottom: 7px solid transparent;
    border-left: 10px solid var(--texto);
    transform: translateY(-50%);
  }
  .figura :global(.sombra),
  .en-fila :global(.sombra) {
    margin-top: -22px;
  }
  @media (max-height: 740px) {
    .escenario {
      inset: 0 0 150px;
    }
    .oak {
      zoom: 2.4;
    }
    .en-fila img,
    .duo .sprite {
      zoom: 1.8;
    }
    .aparece img {
      zoom: 2;
    }
    .vitrina {
      min-height: 170px;
    }
  }
  @media (max-width: 820px) {
    .desfile {
      gap: 10px;
    }
    .en-fila img {
      zoom: 1.5;
    }
    .mesa {
      gap: 10px;
      padding: 14px 18px 20px;
    }
    .vitrina {
      flex-direction: column;
      gap: 8px;
    }
  }
</style>

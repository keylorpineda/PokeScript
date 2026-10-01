<script>
  // Panel del compañero: su pedacito de mundo (como una pantalla de combate),
  // la barra de PS que baja con cada error, lo que dice y la lista de
  // diagnósticos con su arreglo.
  import Sombra from './Sombra.svelte';
  import Dialogo from './Dialogo.svelte';
  import Diagnosticos from './Diagnosticos.svelte';
  import Ataque from './Ataque.svelte';
  import { ide, perfil, miPokemon } from '../lib/estado.svelte.js';
  import { temaActual } from '../lib/temas.js';
  import { POKEMON, sprite, nivelDe } from '../lib/pokemon.js';
  import { buscar, describir, colorTipo } from '../lib/simbolos.js';

  const yo = $derived(miPokemon());
  const datos = $derived(POKEMON[yo] ?? POKEMON.pikachu);
  const fondo = $derived(temaActual(perfil.tema).escena);
  const errores = $derived(ide.diagnosticos.filter((d) => d.severity === 'error'));
  const avisos = $derived(ide.diagnosticos.length - errores.length);
  // Cada error quita 24 PS y cada advertencia 6; debilitado, no le queda nada.
  const ps = $derived(
    ide.debilitado ? 0 : ide.resultado ? Math.max(4, 100 - errores.length * 24 - avisos * 6) : 100,
  );
  const nivel = $derived(nivelDe(perfil.exp));
  const exp = $derived((((perfil.exp ?? 0) % 3) / 3) * 100);
  const colorPs = $derived(ps > 50 ? '#48c050' : ps > 20 ? '#f8b010' : '#f04030');

  // Estado alterado, como en los juegos, según el primer error.
  const ESTADOS = {
    '¡Se escapó!': ['CNF', 'Confundido', 'confundido'],
    'No se encontró la ruta': ['CNF', 'Confundido', 'confundido'],
    'No es muy efectivo…': ['PAR', 'Paralizado', 'paralizado'],
    '¡No pasó nada!': ['DOR', 'Dormido', 'dormido'],
  };
  const estado = $derived(
    ide.debilitado
      ? ['DEB', 'Debilitado', 'debilitado']
      : errores.length
        ? (ESTADOS[errores[0].heading] ?? null)
        : null,
  );

  // Qué es el nombre que está bajo el cursor.
  const bajoCursor = $derived(
    buscar(ide.resultado?.simbolos?.[ide.archivoActivo], ide.palabra).map(describir)[0] ?? null,
  );

  // Las evoluciones son sprites más grandes: se dibujan con menos aumento
  // para que no tapen el cuadro de PS.
  const AUMENTO = {
    ivysaur: 1.5,
    charmeleon: 1.5,
    wartortle: 1.5,
    raichu: 1.5,
    venusaur: 1.25,
    charizard: 1.25,
    blastoise: 1.25,
  };
  const aumento = (id) => AUMENTO[id] ?? 2;

  let habitat = $state();
  // La evolución se ve un momento: la forma vieja y la nueva parpadean.
  let evolucionando = $state(null);
  $effect(() => {
    const e = ide.evolucion;
    if (!e) return;
    evolucionando = e;
    const t = setTimeout(() => (evolucionando = null), 2600);
    return () => clearTimeout(t);
  });
</script>

<aside class="companero">
  <div class="habitat marco" bind:this={habitat}>
    <div
      class="paisaje"
      style="background-image:url(/fondos/{fondo.fondo}.png);filter:{fondo.filtro ??
        'var(--filtro-escena)'}"
    ></div>
    {#if ide.asistente.invitado}
      {#key ide.asistente.vez}
        <div class="invitado">
          <img class="sprite" src={sprite(ide.asistente.invitado)} alt="" draggable="false" />
        </div>
      {/key}
    {/if}
    {#key ide.asistente.vez}
      <div class="yo {ide.asistente.animo}">
        <span class="cuerpo {estado?.[2] ?? ''}" class:evolucionando>
          {#if evolucionando}
            <img
              class="sprite silueta vieja"
              style="zoom:{aumento(evolucionando.de)}"
              src={sprite(evolucionando.de)}
              alt=""
            />
            <img
              class="sprite silueta nueva"
              style="zoom:{aumento(evolucionando.a)}"
              src={sprite(evolucionando.a)}
              alt={datos.nombre}
            />
          {:else}
            <img
              class="sprite"
              style="zoom:{aumento(yo)}"
              src={sprite(yo)}
              alt={datos.nombre}
              draggable="false"
            />
          {/if}
          {#if estado?.[2] === 'dormido'}
            <span class="zzz"><i>Z</i><i>Z</i><i>Z</i></span>
          {:else if estado?.[2] === 'confundido'}
            <span class="mareo"><i>★</i><i>★</i><i>★</i></span>
          {/if}
        </span>
        <Sombra />
      </div>
    {/key}
    {#key ide.celebracion}
      {#if ide.celebracion && habitat}<Ataque forma={yo} {habitat} />{/if}
    {/key}
    {#if bajoCursor}
      <div class="bajo-cursor" title="Lo que hay bajo el cursor">
        <code>{bajoCursor.texto}</code>
        {#if bajoCursor.tipo}
          <span class="chip-tipo" style="background:{colorTipo(bajoCursor.tipo)}"
            >{bajoCursor.tipo}</span
          >
        {:else}
          <span class="clase">{bajoCursor.clase}</span>
        {/if}
      </div>
    {/if}

    <div class="estado">
      <div class="fila">
        <b>{datos.nombre.toUpperCase()}</b>
        {#if estado}<span class="alterado {estado[2]}" title={estado[1]}>{estado[0]}</span>{/if}
        <span>Nv{nivel}</span>
      </div>
      <div class="barra">
        <span class="ps">PS</span>
        <span class="riel"
          ><span class="lleno" style="width:{ps}%;background:{colorPs}"></span></span
        >
      </div>
      <div class="exp"><span style="width:{exp}%"></span></div>
    </div>
  </div>

  <div class="habla">
    {#key ide.asistente.vez}
      <Dialogo
        texto={ide.asistente.texto}
        nombre={datos.nombre}
        compacto
        silencioso
        teclado={false}
        velocidad={16}
      />
    {/key}
  </div>

  <Diagnosticos />
</aside>

<style>
  .companero {
    display: flex;
    flex-direction: column;
    gap: 22px;
    min-height: 0;
    overflow-y: auto;
    padding: 2px 4px 10px;
  }
  .habitat {
    position: relative;
    height: clamp(170px, 27vh, 230px);
    flex: none;
    overflow: hidden;
    border-image-source: var(--marco-oscuro);
  }
  .paisaje {
    position: absolute;
    inset: -12px;
    background-size: cover;
    background-position: center 70%;
    image-rendering: pixelated;
  }
  .yo {
    position: absolute;
    left: 50%;
    bottom: 12px;
    display: flex;
    flex-direction: column;
    align-items: center;
    transform: translateX(-50%);
  }
  .yo img,
  .invitado img {
    position: relative;
    z-index: 1;
    zoom: 2;
  }
  .yo.feliz img {
    animation: saltar 0.35s steps(3) 3;
  }
  @keyframes saltar {
    50% {
      transform: translateY(-14px);
    }
  }
  .yo.triste img {
    animation: sacudir 0.5s steps(4) 2;
  }
  @keyframes sacudir {
    25% {
      transform: translateX(-8px);
      filter: brightness(1.8);
    }
    75% {
      transform: translateX(8px);
    }
  }
  .invitado {
    position: absolute;
    right: 8px;
    top: 36px;
    animation: llegar 0.5s steps(6) both;
  }
  .invitado img {
    zoom: 1.5;
  }
  @keyframes llegar {
    from {
      transform: translateX(160px);
    }
  }
  .estado {
    position: absolute;
    left: 6px;
    top: 6px;
    width: 176px;
    padding: 4px 8px 6px;
    font-family: var(--titulo);
    color: #202028;
    background: #f8f8e8;
    border: 3px solid #303038;
    border-radius: 6px 0 6px 0;
    box-shadow: 3px 3px 0 rgba(0, 0, 0, 0.35);
  }
  .fila {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 15px;
  }
  .fila b {
    margin-right: auto;
  }
  /* Estado alterado, con los colores de los juegos. */
  .alterado {
    padding: 0 4px;
    font-size: 11px;
    font-weight: 700;
    line-height: 15px;
    color: #fff;
    border-radius: 3px;
  }
  .alterado.paralizado {
    background: #c8a000;
  }
  .alterado.dormido {
    background: #8c8c8c;
  }
  .alterado.confundido {
    background: #b060c0;
  }
  .alterado.debilitado {
    background: #c03028;
  }
  .barra {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-top: 2px;
    padding: 2px 3px;
    background: #484858;
    border-radius: 4px;
  }
  .ps {
    font-size: 11px;
    font-weight: 700;
    color: #f8c030;
  }
  .riel {
    flex: 1;
    height: 7px;
    background: #f8f8f8;
    border-radius: 3px;
    overflow: hidden;
  }
  .lleno {
    display: block;
    height: 100%;
    transition:
      width 0.8s steps(12),
      background 0.3s;
  }
  .exp {
    height: 4px;
    margin-top: 4px;
    background: #d0d0c0;
  }
  .exp span {
    display: block;
    height: 100%;
    background: #40a8f8;
    transition: width 0.6s steps(8);
  }
  .habla {
    flex: none;
  }

  /* El sprite y sus estados. */
  .cuerpo {
    position: relative;
    display: grid;
    justify-items: center;
    align-items: end;
  }
  .cuerpo > img {
    grid-area: 1 / 1;
  }
  .cuerpo.paralizado {
    animation: paralizar 1.6s steps(1) infinite;
  }
  @keyframes paralizar {
    0%,
    70%,
    82% {
      filter: none;
    }
    74%,
    88% {
      filter: sepia(1) saturate(4) hue-rotate(10deg) brightness(1.25);
    }
  }
  .cuerpo.dormido {
    filter: grayscale(0.35) brightness(0.92);
  }
  .cuerpo.debilitado {
    filter: grayscale(1) brightness(0.85);
    opacity: 0.75;
    transform: translateY(8px) rotate(-10deg);
  }
  .zzz,
  .mareo {
    position: absolute;
    z-index: 2;
    pointer-events: none;
    font-family: var(--titulo);
    font-weight: 700;
  }
  /* Dormido: tres Z que suben desde la cabeza. */
  .zzz {
    top: 4px;
    left: 34%;
    width: 0;
    height: 0;
  }
  .zzz i {
    position: absolute;
    left: 0;
    top: 0;
    font-family: var(--codigo);
    font-size: 20px;
    font-weight: 800;
    font-style: normal;
    line-height: 1;
    color: #fff;
    text-shadow:
      2px 0 0 #2a3350,
      -2px 0 0 #2a3350,
      0 2px 0 #2a3350,
      0 -2px 0 #2a3350;
    opacity: 0;
    animation: dormir 2.4s linear infinite;
  }
  .zzz i:nth-child(2) {
    animation-delay: 0.8s;
  }
  .zzz i:nth-child(3) {
    animation-delay: 1.6s;
  }
  @keyframes dormir {
    from {
      transform: translate(0, 0) scale(0.6);
      opacity: 0;
    }
    20% {
      opacity: 1;
    }
    to {
      transform: translate(18px, -30px) scale(1.25);
      opacity: 0;
    }
  }
  /* Confundido: estrellas que giran sobre la cabeza, en una elipse. */
  .mareo {
    top: -4px;
    left: 40%;
    width: 0;
    height: 0;
  }
  .mareo i {
    position: absolute;
    left: -6px;
    top: -8px;
    font-size: 13px;
    font-style: normal;
    color: #ffe14a;
    text-shadow:
      1px 1px 0 #6a4a00,
      -1px -1px 0 #6a4a00;
    animation: marear 1.2s linear infinite;
  }
  .mareo i:nth-child(2) {
    animation-delay: -0.4s;
  }
  .mareo i:nth-child(3) {
    animation-delay: -0.8s;
  }
  @keyframes marear {
    0% {
      transform: translate(22px, 0) scale(1);
      z-index: 2;
    }
    12.5% {
      transform: translate(15.6px, 4.2px) scale(1.1);
    }
    25% {
      transform: translate(0, 6px) scale(1.15);
    }
    37.5% {
      transform: translate(-15.6px, 4.2px) scale(1.1);
    }
    50% {
      transform: translate(-22px, 0) scale(1);
    }
    62.5% {
      transform: translate(-15.6px, -4.2px) scale(0.85);
    }
    75% {
      transform: translate(0, -6px) scale(0.8);
    }
    87.5% {
      transform: translate(15.6px, -4.2px) scale(0.85);
    }
    100% {
      transform: translate(22px, 0) scale(1);
    }
  }
  /* Evolución: dos siluetas blancas que se turnan cada vez más rápido. */
  .silueta {
    filter: brightness(0) invert(1) drop-shadow(0 0 6px #fff);
  }
  .silueta.vieja {
    animation: turnoVieja 2.6s linear both;
  }
  .silueta.nueva {
    animation: turnoNueva 2.6s linear both;
  }
  @keyframes turnoVieja {
    0%,
    14%,
    30%,
    42%,
    52%,
    60%,
    67%,
    73% {
      opacity: 1;
    }
    8%,
    22%,
    36%,
    47%,
    56%,
    64%,
    70%,
    76%,
    100% {
      opacity: 0;
    }
  }
  @keyframes turnoNueva {
    0%,
    14%,
    30%,
    42%,
    52%,
    60%,
    67%,
    73% {
      opacity: 0;
    }
    8%,
    22%,
    36%,
    47%,
    56%,
    64%,
    70%,
    76% {
      opacity: 1;
    }
    88% {
      opacity: 1;
      filter: brightness(0) invert(1) drop-shadow(0 0 6px #fff);
    }
    100% {
      opacity: 1;
      filter: none;
    }
  }
  /* El tipo de lo que está bajo el cursor. */
  .bajo-cursor {
    position: absolute;
    right: 6px;
    bottom: 6px;
    z-index: 2;
    display: flex;
    align-items: center;
    gap: 6px;
    max-width: calc(100% - 12px);
    padding: 2px 6px;
    font-family: var(--titulo);
    font-size: 12px;
    color: #202028;
    background: #f8f8e8;
    border: 3px solid #303038;
    border-radius: 6px 0 6px 0;
    box-shadow: 3px 3px 0 rgba(0, 0, 0, 0.35);
  }
  .bajo-cursor code {
    overflow: hidden;
    font-family: var(--codigo);
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .chip-tipo {
    padding: 0 5px;
    color: #fff;
    white-space: nowrap;
    text-shadow: 1px 1px 0 rgba(0, 0, 0, 0.45);
    border: 2px solid rgba(0, 0, 0, 0.35);
    border-radius: 3px;
  }
  .bajo-cursor .clase {
    color: #606070;
    text-transform: uppercase;
  }
</style>

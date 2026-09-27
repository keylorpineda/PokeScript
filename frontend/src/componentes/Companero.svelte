<script>
  // Panel del compañero: su pedacito de mundo (como una pantalla de combate),
  // la barra de PS que baja con cada error, lo que dice y la lista de
  // diagnósticos con su arreglo.
  import Sombra from './Sombra.svelte';
  import Dialogo from './Dialogo.svelte';
  import Diagnosticos from './Diagnosticos.svelte';
  import { ide, perfil } from '../lib/estado.svelte.js';
  import { temaActual } from '../lib/temas.js';
  import { POKEMON, sprite } from '../lib/pokemon.js';

  const datos = $derived(POKEMON[perfil.companero] ?? POKEMON.pikachu);
  const fondo = $derived(temaActual(perfil.tema).escena);
  const errores = $derived(ide.diagnosticos.filter((d) => d.severity === 'error').length);
  const ps = $derived(ide.resultado ? Math.max(4, 100 - errores * 24) : 100);
  const nivel = $derived(5 + Math.floor((perfil.exp ?? 0) / 3));
  const exp = $derived((((perfil.exp ?? 0) % 3) / 3) * 100);
  const colorPs = $derived(ps > 50 ? '#48c050' : ps > 20 ? '#f8b010' : '#f04030');
</script>

<aside class="companero">
  <div class="habitat marco">
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
        <img
          class="sprite"
          src={sprite(perfil.companero ?? 'pikachu')}
          alt={datos.nombre}
          draggable="false"
        />
        <Sombra />
      </div>
    {/key}

    <div class="estado">
      <div class="fila">
        <b>{datos.nombre.toUpperCase()}</b>
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
    justify-content: space-between;
    font-size: 15px;
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
</style>

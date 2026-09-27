<script>
  // Ventana principal del IDE: mochila de archivos, editor, salida del
  // combate y el panel del compañero, sobre el lugar del tema.
  import { onMount } from 'svelte';
  import BarraSuperior from './BarraSuperior.svelte';
  import Mochila from './Mochila.svelte';
  import Editor from './Editor.svelte';
  import Escaneo from './Escaneo.svelte';
  import Salida from './Salida.svelte';
  import Companero from './Companero.svelte';
  import Transicion from './Transicion.svelte';
  import OakAviso from './OakAviso.svelte';
  import Pokedex from './Pokedex.svelte';
  import { ide, perfil, navegacion } from '../lib/estado.svelte.js';
  import { temaActual } from '../lib/temas.js';
  import { objeto, sprite } from '../lib/pokemon.js';
  import { abrirProyecto, escucharEjecucion, guardarAhora } from '../lib/acciones.js';
  import { sonar, volumenGeneral } from '../lib/sonido.js';

  let { ruta } = $props();

  let listo = $state(false);

  // Diseño adaptable: en ventanas medianas el compañero pasa a un cajón a la
  // derecha y en las chicas la mochila también, a la izquierda.
  let ancho = $state(window.innerWidth);
  let verCompanero = $state(false);
  let verMochila = $state(false);
  const cajonCompanero = $derived(ancho < 1150);
  const cajonMochila = $derived(ancho < 820);
  const errores = $derived(ide.diagnosticos.filter((d) => d.severity === 'error').length);

  // Al abrir un archivo desde el cajón, el cajón se cierra solo.
  $effect(() => {
    ide.archivoActivo;
    verMochila = false;
  });
  let combate = $state(false);
  let ultima = 0;

  const escena = $derived(temaActual(perfil.tema).escena);

  // En el editor los efectos suenan más bajito que en la portada.
  onMount(() => {
    volumenGeneral(0.5);
    return () => volumenGeneral(1);
  });

  onMount(async () => {
    escucharEjecucion();
    try {
      await abrirProyecto(ruta);
    } catch (e) {
      // El proyecto ya no está (se borró o se movió): de vuelta al menú.
      navegacion.aviso = `No se pudo abrir ${ruta}: ${e?.message ?? e}.`;
      navegacion.pantalla = 'menu';
      return;
    }
    listo = true;
  });

  $effect(() => {
    if (ide.transicion > ultima) {
      ultima = ide.transicion;
      combate = true;
    }
  });
</script>

<svelte:window bind:innerWidth={ancho} />

<div class="ide">
  <div
    class="lugar"
    style="background-image:url(/fondos/{escena.fondo}.png);filter:{escena.filtro ??
      'var(--filtro-escena)'}"
  ></div>

  <BarraSuperior />

  <main class:sin-companero={cajonCompanero} class:solo-centro={cajonMochila}>
    {#if !cajonMochila}<Mochila />{/if}

    <section class="centro">
      <div class="pestanas">
        {#each ide.proyecto?.archivos ?? [] as archivo (archivo)}
          <button
            class="pestana"
            class:activa={ide.archivoActivo === archivo}
            onclick={() => {
              if (ide.archivoActivo !== archivo) sonar('abrir');
              ide.archivoActivo = archivo;
            }}
          >
            <img
              class="pixel"
              src={objeto(archivo === ide.proyecto.principal ? 'master-ball' : 'poke-ball')}
              alt=""
            />
            {archivo}{#if ide.sucios[archivo]}<span class="punto"></span>{/if}
          </button>
        {/each}
      </div>
      <div class="hoja marco">
        <div class="ruta">
          <span class="miga">{ide.proyecto?.nombre ?? ''}</span>
          <span class="sep"></span>
          <b>{ide.archivoActivo ?? ''}</b>
          <button
            class="estado-guardado der"
            class:pendiente={Object.keys(ide.sucios).length}
            onclick={guardarAhora}
            title="Guardar ahora (Ctrl+S)"
          >
            {Object.keys(ide.sucios).length ? 'SIN GUARDAR' : 'GUARDADO'}
          </button>
          <span class="lenguaje">POKESCRIPT</span>
        </div>
        <div class="codigo">
          {#if listo}<Editor />{/if}
          {#if ide.compilando}<Escaneo />{/if}
        </div>
      </div>
      <Salida />
    </section>

    {#if !cajonCompanero}<Companero />{/if}

    {#if cajonMochila}
      <button class="pestana-cajon izq" onclick={() => (verMochila = !verMochila)} title="Mochila">
        <img class="pixel" src={objeto('exp-share')} alt="" /><span>MOCHILA</span>
      </button>
    {/if}
    {#if cajonCompanero}
      <button
        class="pestana-cajon der"
        onclick={() => (verCompanero = !verCompanero)}
        title="Compañero y errores"
      >
        <img class="sprite" src={sprite(perfil.companero ?? 'pikachu')} alt="" />
        {#if errores}<span class="cuenta">{errores}</span>{/if}
      </button>
    {/if}
  </main>

  {#if (cajonMochila && verMochila) || (cajonCompanero && verCompanero)}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="velo-cajon"
      onclick={() => {
        verMochila = false;
        verCompanero = false;
      }}
    ></div>
  {/if}
  {#if cajonMochila && verMochila}
    <div class="cajon izq"><Mochila /></div>
  {/if}
  {#if cajonCompanero && verCompanero}
    <div class="cajon der"><Companero /></div>
  {/if}

  <footer>
    <span>Línea {ide.cursor.linea}, columna {ide.cursor.col}</span>
    <span>{ide.archivoActivo ?? ''}</span>
    <span class="paseo" aria-hidden="true">
      <img class="sprite" src={sprite(perfil.companero ?? 'pikachu')} alt="" />
    </span>
    <span class="der ancho">Tema: {temaActual(perfil.tema).nombre}</span>
    <span class="ancho">Entrenador: {perfil.nombre}</span>
  </footer>
</div>

{#if combate}<Transicion alTerminar={() => (combate = false)} />{/if}
{#if ide.pokedex}<Pokedex />{/if}
<OakAviso />

<style>
  .ide {
    position: fixed;
    inset: 0;
    display: grid;
    grid-template-rows: auto 1fr auto;
    background: var(--fondo);
    animation: entrar 0.5s steps(5) both;
  }
  @keyframes entrar {
    from {
      opacity: 0;
    }
  }
  /* El lugar del tema, tenue, detrás de los paneles. */
  .lugar {
    position: absolute;
    inset: 0;
    background-size: cover;
    background-position: center bottom;
    image-rendering: pixelated;
    opacity: 0.22;
    pointer-events: none;
  }
  main {
    position: relative;
    display: grid;
    grid-template-columns: minmax(200px, 260px) minmax(0, 1fr) minmax(270px, 340px);
    gap: 14px;
    padding: 14px;
    min-height: 0;
  }
  main.sin-companero {
    grid-template-columns: minmax(190px, 240px) minmax(0, 1fr);
  }
  main.solo-centro {
    grid-template-columns: minmax(0, 1fr);
    padding: 10px 44px 10px 44px;
  }
  .centro {
    display: grid;
    grid-template-rows: auto 1fr clamp(130px, 24vh, 210px);
    min-height: 0;
    min-width: 0;
  }
  .pestanas {
    display: flex;
    gap: 4px;
    padding-left: 12px;
    overflow-x: auto;
    scrollbar-width: none;
  }
  .pestana {
    flex: none;
    white-space: nowrap;
  }
  .pestana {
    position: relative;
    top: 4px;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 3px 12px 5px 8px;
    font-family: var(--titulo);
    font-size: 16px;
    color: var(--texto-suave);
    background: var(--fondo-2);
    border: 3px solid var(--borde);
    border-bottom: 0;
    cursor: pointer;
  }
  .pestana img {
    width: 18px;
    height: 18px;
  }
  .pestana.activa {
    top: 1px;
    padding-bottom: 8px;
    color: var(--texto);
    background: var(--panel);
    box-shadow: inset 0 3px 0 var(--acento);
    z-index: 1;
  }
  .punto {
    width: 7px;
    height: 7px;
    background: var(--acento-2);
    border: 2px solid var(--borde);
  }
  .hoja {
    display: flex;
    flex-direction: column;
    min-height: 0;
    margin-bottom: 14px;
    overflow: hidden;
  }
  .ruta {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 4px 6px;
    font-family: var(--titulo);
    font-size: 14px;
    color: var(--texto-suave);
    border-bottom: 2px dashed var(--borde-suave);
  }
  .ruta b {
    color: var(--texto);
  }
  .sep {
    border-top: 4px solid transparent;
    border-bottom: 4px solid transparent;
    border-left: 6px solid var(--texto-suave);
  }
  .ruta .der {
    margin-left: auto;
  }
  .lenguaje {
    padding: 0 6px;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 1px;
    color: var(--acento-texto);
    background: var(--acento);
    border: 2px solid var(--borde);
  }
  .codigo {
    position: relative;
    flex: 1;
    min-height: 0;
    margin-top: 6px;
  }
  footer {
    position: relative;
    display: flex;
    align-items: center;
    gap: 22px;
    height: 30px;
    padding: 0 14px;
    font-family: var(--titulo);
    font-size: 14px;
    color: var(--acento-texto);
    background: var(--acento);
    border-top: 4px solid var(--borde);
  }
  .der {
    margin-left: auto;
  }
  /* El compañero pasea por la barra de estado. */
  .paseo {
    position: absolute;
    left: 38%;
    bottom: 0;
    width: 30%;
    height: 44px;
    pointer-events: none;
  }
  .paseo img {
    position: absolute;
    bottom: 0;
    height: 40px;
    animation: pasear 16s steps(64) infinite;
  }
  @keyframes pasear {
    0% {
      left: 0;
      transform: scaleX(-1);
    }
    49% {
      left: calc(100% - 40px);
      transform: scaleX(-1);
    }
    50% {
      left: calc(100% - 40px);
      transform: scaleX(1);
    }
    100% {
      left: 0;
      transform: scaleX(1);
    }
  }
  .estado-guardado {
    padding: 0 6px;
    font-family: var(--titulo);
    font-size: 11px;
    letter-spacing: 1px;
    color: var(--exito);
    background: none;
    border: 2px solid currentColor;
    cursor: pointer;
  }
  .estado-guardado.pendiente {
    color: var(--aviso);
  }
  /* Cajones laterales para ventanas angostas. */
  .pestana-cajon {
    position: absolute;
    top: 50%;
    z-index: 3;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 8px 4px;
    font-family: var(--titulo);
    font-size: 11px;
    letter-spacing: 1px;
    color: var(--acento-texto);
    background: var(--acento);
    border: 3px solid var(--borde);
    cursor: pointer;
    transform: translateY(-50%);
  }
  .pestana-cajon.izq {
    left: 0;
    border-left: 0;
    border-radius: 0 10px 10px 0;
  }
  .pestana-cajon.der {
    right: 0;
    border-right: 0;
    border-radius: 10px 0 0 10px;
  }
  .pestana-cajon span {
    writing-mode: vertical-rl;
  }
  .pestana-cajon img {
    width: 30px;
    height: 30px;
    object-fit: contain;
  }
  .pestana-cajon .cuenta {
    writing-mode: horizontal-tb;
    min-width: 20px;
    padding: 0 4px;
    font-size: 12px;
    color: #fff;
    background: var(--error);
    border: 2px solid var(--borde);
  }
  .velo-cajon {
    position: fixed;
    inset: 0;
    z-index: 30;
    background: rgba(8, 6, 16, 0.45);
  }
  .cajon {
    position: fixed;
    top: 70px;
    bottom: 40px;
    z-index: 31;
    display: flex;
    width: min(360px, 88vw);
    padding: 8px;
    background: var(--fondo);
    border: 4px solid var(--borde);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
    animation: deslizar 0.2s steps(4) both;
  }
  .cajon > :global(*) {
    flex: 1;
    min-width: 0;
  }
  .cajon.izq {
    left: 0;
    border-left: 0;
    border-radius: 0 12px 12px 0;
    --desde: -100%;
  }
  .cajon.der {
    right: 0;
    border-right: 0;
    border-radius: 12px 0 0 12px;
    --desde: 100%;
  }
  @keyframes deslizar {
    from {
      transform: translateX(var(--desde));
    }
  }
  @media (max-width: 1000px) {
    .ancho {
      display: none;
    }
    footer {
      gap: 14px;
      font-size: 12px;
    }
  }
  @media (max-height: 640px) {
    main {
      padding-top: 8px;
      padding-bottom: 8px;
    }
    .hoja {
      margin-bottom: 8px;
    }
  }
</style>

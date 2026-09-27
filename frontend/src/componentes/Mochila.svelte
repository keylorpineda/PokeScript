<script>
  // La mochila: el árbol de archivos del proyecto. Cada archivo es una
  // Pokéball y el principal es la Master Ball. Con clic derecho (o el botón
  // de opciones) se le cambia el nombre, se marca como principal o se suelta.
  import { tick } from 'svelte';
  import { ide } from '../lib/estado.svelte.js';
  import { objeto } from '../lib/pokemon.js';
  import { sonar } from '../lib/sonido.js';
  import {
    nuevoArchivo,
    renombrarArchivo,
    borrarArchivo,
    marcarPrincipal,
    duplicarArchivo,
  } from '../lib/acciones.js';
  import { PLANTILLAS, tipoDeArchivo } from '../lib/plantillas.js';

  let abierta = $state(true);
  let eligiendo = $state(false); // panel de plantillas abierto
  let plantilla = $state(null); // la elegida, esperando el nombre
  let nuevo = $state('');
  let renombrando = $state(null);
  let apodo = $state('');
  let menu = $state(null); // { archivo, x, y }
  let soltando = $state(null);

  const principal = $derived(ide.proyecto?.principal);
  const archivos = $derived(
    [...(ide.proyecto?.archivos ?? [])].sort((a, b) =>
      a === principal ? -1 : b === principal ? 1 : a.localeCompare(b),
    ),
  );

  const errores = (a) =>
    ide.diagnosticos.filter((d) => d.file === a && d.severity === 'error').length;
  const avisos = (a) =>
    ide.diagnosticos.filter((d) => d.file === a && d.severity !== 'error').length;

  function abrir(archivo) {
    if (ide.archivoActivo !== archivo) sonar('abrir');
    ide.archivoActivo = archivo;
  }

  function opciones(e, archivo) {
    e.preventDefault();
    e.stopPropagation();
    sonar('mover');
    const r = e.currentTarget.closest('.mochila').getBoundingClientRect();
    menu = { archivo, x: e.clientX - r.left, y: e.clientY - r.top };
  }

  function empezarNuevo(e) {
    e.stopPropagation();
    eligiendo = !eligiendo;
    plantilla = null;
    sonar(eligiendo ? 'elegir' : 'mover');
  }

  // nombreLibre sugiere el nombre de la plantilla sin chocar con otro archivo.
  function nombreLibre(base) {
    const hay = new Set(ide.proyecto?.archivos ?? []);
    let n = base;
    for (let i = 2; hay.has(`${n}.pks`); i++) n = `${base}${i}`;
    return n;
  }

  async function elegirPlantilla(p) {
    plantilla = p;
    nuevo = nombreLibre(p.nombre);
    sonar('elegir');
    await tick();
    const c = document.querySelector('.mochila .campo-nuevo');
    c?.focus();
    c?.select();
  }

  async function crear(e) {
    e.preventDefault();
    if (!nuevo.trim()) return;
    await nuevoArchivo(nuevo, plantilla.contenido);
    eligiendo = false;
    plantilla = null;
  }

  async function empezarRenombrar(archivo) {
    menu = null;
    renombrando = archivo;
    apodo = archivo.replace(/\.pks$/, '');
    await tick();
    const c = document.querySelector('.mochila .campo-apodo');
    c?.focus();
    c?.select();
  }

  async function renombrar(e) {
    e.preventDefault();
    if (apodo.trim()) await renombrarArchivo(renombrando, apodo);
    renombrando = null;
  }
</script>

<svelte:window onclick={() => (menu = null)} />

<nav class="mochila marco">
  <div class="bolsillo">
    <img class="pixel" src={objeto('exp-share')} alt="" />
    <span>MOCHILA</span>
    <button class="mas" class:abierto={eligiendo} onclick={empezarNuevo} title="Nuevo archivo"
      >NUEVO</button
    >
  </div>

  {#if eligiendo}
    <div class="nuevo marco">
      {#if !plantilla}
        <p class="pregunta">¿Con qué empieza el archivo?</p>
        <p class="nota">
          Todos son archivos <code>.pks</code>: lo que cambia es lo que traen adentro.
        </p>
        {#each PLANTILLAS as p (p.id)}
          <button class="plantilla" onclick={() => elegirPlantilla(p)}>
            <img class="pixel" src={objeto(p.objeto)} alt="" />
            <span>
              <b>{p.titulo}</b>
              <small>{p.descripcion}</small>
            </span>
          </button>
        {/each}
      {:else}
        <p class="pregunta">
          <img class="pixel" src={objeto(plantilla.objeto)} alt="" />{plantilla.titulo}
        </p>
        <form class="nombrar" onsubmit={crear}>
          <label>
            <input
              class="campo-nuevo"
              bind:value={nuevo}
              spellcheck="false"
              onkeydown={(e) => e.key === 'Escape' && (plantilla = null)}
            /><span>.pks</span>
          </label>
          <div class="si-no">
            <button class="boton principal" type="submit">CREAR</button>
            <button class="boton" type="button" onclick={() => (plantilla = null)}>ATRÁS</button>
          </div>
        </form>
      {/if}
    </div>
  {/if}

  <div class="arbol">
    <button class="carpeta" onclick={() => (abierta = !abierta)}>
      <span class="flechita" class:abierta></span>
      <svg
        class="icono-carpeta"
        viewBox="0 0 12 10"
        shape-rendering="crispEdges"
        aria-hidden="true"
      >
        <rect x="0" y="1" width="5" height="2" fill="var(--acento-2)" />
        <rect x="0" y="2" width="12" height="8" fill="var(--acento-2)" />
        <rect
          x="0"
          y="4"
          width="12"
          height="6"
          fill="color-mix(in srgb, var(--acento-2) 70%, #000)"
        />
        <rect x="1" y="5" width="10" height="1" fill="rgba(255,255,255,.35)" />
      </svg>
      <span class="nombre-proyecto">{ide.proyecto?.nombre ?? '…'}</span>
      <span class="cuantos">{archivos.length}</span>
    </button>

    {#if abierta}
      <ul>
        {#each archivos as archivo (archivo)}
          {@const tipo =
            archivo === principal
              ? {
                  objeto: 'master-ball',
                  nombre: 'Archivo principal: tiene el combate',
                  clase: 'combate',
                }
              : tipoDeArchivo(ide.contenidos[archivo])}
          <li class={tipo.clase} class:activo={ide.archivoActivo === archivo}>
            {#if renombrando === archivo}
              <form class="fila" onsubmit={renombrar}>
                <img class="pixel bola" src={objeto('poke-ball')} alt="" />
                <input
                  class="campo-apodo"
                  bind:value={apodo}
                  onblur={() => (renombrando = null)}
                  onkeydown={(e) => e.key === 'Escape' && (renombrando = null)}
                />
              </form>
            {:else}
              <button
                class="fila"
                onclick={() => abrir(archivo)}
                ondblclick={() => empezarRenombrar(archivo)}
                oncontextmenu={(e) => opciones(e, archivo)}
                title="{tipo.nombre} · doble clic: cambiar nombre · clic derecho: más opciones"
              >
                <img class="pixel bola" src={objeto(tipo.objeto)} alt={tipo.nombre} />
                <span class="nombre">
                  {archivo.replace(/\.pks$/, '')}<small>.pks</small>
                </span>
                {#if archivo === principal}<span class="chip">PRINCIPAL</span>{/if}
                {#if ide.sucios[archivo]}<span class="sucio" title="Sin guardar"></span>{/if}
                {#if errores(archivo)}
                  <span class="cuenta error">{errores(archivo)}</span>
                {:else if avisos(archivo)}
                  <span class="cuenta aviso">{avisos(archivo)}</span>
                {/if}
                <!-- svelte-ignore a11y_click_events_have_key_events -->
                <!-- svelte-ignore a11y_no_static_element_interactions -->
                <span class="tres" onclick={(e) => opciones(e, archivo)} title="Opciones"
                  ><i></i><i></i><i></i></span
                >
              </button>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  {#if soltando}
    <div class="confirmar marco">
      <p>¿Soltar <b>{soltando}</b>? Se borrará el archivo.</p>
      <div class="si-no">
        <button
          class="boton"
          onclick={() => {
            borrarArchivo(soltando);
            soltando = null;
          }}>SÍ, SOLTAR</button
        >
        <button class="boton" onclick={() => (soltando = null)}>NO</button>
      </div>
    </div>
  {/if}

  <p class="pista">
    La <b>Master Ball</b> es el archivo principal: ahí vive el <code>combate</code>.
  </p>

  {#if menu}
    <div class="menu marco" style="left:{menu.x}px;top:{menu.y}px">
      <button onclick={() => empezarRenombrar(menu.archivo)}>Cambiar nombre</button>
      <button
        onclick={() => {
          duplicarArchivo(menu.archivo);
          menu = null;
        }}>Duplicar</button
      >
      <button
        disabled={menu.archivo === principal}
        onclick={() => {
          marcarPrincipal(menu.archivo);
          menu = null;
        }}>Hacer principal</button
      >
      <button
        class="peligro"
        disabled={menu.archivo === principal}
        onclick={() => {
          soltando = menu.archivo;
          menu = null;
          sonar('pregunta');
        }}>Soltar (borrar)</button
      >
    </div>
  {/if}
</nav>

<style>
  .mochila {
    position: relative;
    display: flex;
    flex-direction: column;
    min-height: 0;
    padding: 0 2px;
  }
  .bolsillo {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 4px 4px 6px;
    font-family: var(--titulo);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: 2px;
    color: var(--acento-texto);
    background: var(--acento);
    border: 3px solid var(--borde);
    box-shadow: inset 0 -3px 0 rgba(0, 0, 0, 0.2);
  }
  .bolsillo img {
    width: 30px;
  }
  .mas {
    margin-left: auto;
    padding: 1px 6px;
    font-family: var(--titulo);
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 1px;
    color: var(--borde);
    background: var(--acento-2);
    border: 2px solid var(--borde);
    cursor: pointer;
  }
  .mas:hover {
    transform: translateY(-1px);
  }
  .arbol {
    flex: 1;
    margin-top: 10px;
    overflow-y: auto;
  }
  .carpeta {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    padding: 4px 2px;
    font-family: var(--titulo);
    font-size: 16px;
    font-weight: 700;
    text-align: left;
    background: none;
    border: 0;
    cursor: pointer;
  }
  .flechita {
    border-top: 5px solid transparent;
    border-bottom: 5px solid transparent;
    border-left: 7px solid var(--texto-suave);
    transition: transform 0.1s steps(2);
  }
  .flechita.abierta {
    transform: rotate(90deg);
  }
  .icono-carpeta {
    width: 20px;
    height: 17px;
  }
  .nombre-proyecto {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .cuantos {
    font-size: 12px;
    color: var(--texto-suave);
  }
  ul {
    margin: 2px 0 0 10px;
    padding: 0 0 0 10px;
    list-style: none;
    border-left: 2px dashed var(--borde-suave);
  }
  li {
    position: relative;
    margin: 2px 0;
  }
  li::before {
    content: '';
    position: absolute;
    left: -10px;
    top: 17px;
    width: 8px;
    border-top: 2px dashed var(--borde-suave);
  }
  .fila {
    position: relative;
    display: flex;
    align-items: center;
    gap: 7px;
    width: 100%;
    padding: 4px 4px 2px;
    font-family: var(--titulo);
    font-size: 17px;
    text-align: left;
    background: none;
    border: 2px solid transparent;
    cursor: pointer;
  }
  .fila::after {
    content: '';
    position: absolute;
    inset: 0;
    z-index: -1;
    background: linear-gradient(
      90deg,
      color-mix(in srgb, var(--acento) 26%, transparent),
      transparent 85%
    );
    box-shadow: inset 4px 0 0 var(--acento);
    transform: scaleX(0);
    transform-origin: left;
    transition: transform 0.16s steps(4);
  }
  .fila {
    z-index: 0;
    transition: transform 0.12s steps(3);
  }
  li:hover .fila {
    transform: translateX(3px);
  }
  li:hover .fila::after {
    transform: scaleX(1);
  }
  li:hover .bola {
    animation: menear 0.5s steps(2) infinite;
  }
  @keyframes menear {
    25% {
      transform: rotate(-18deg);
    }
    75% {
      transform: rotate(18deg);
    }
  }
  li.vacio .bola {
    opacity: 0.55;
  }
  li.activo .fila {
    background: var(--panel-2);
    border-color: var(--borde);
    box-shadow: 2px 2px 0 var(--borde);
  }
  li.activo .bola {
    animation: botar 0.6s steps(2) infinite;
  }
  @keyframes botar {
    50% {
      transform: translateY(-2px);
    }
  }
  .bola {
    width: 24px;
    height: 24px;
    flex: none;
  }
  .nombre {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .nombre small {
    color: var(--texto-suave);
    font-size: 13px;
  }
  .nuevo {
    margin-top: 8px;
    padding: 0 2px;
    animation: desplegar 0.2s steps(3) both;
  }
  @keyframes desplegar {
    from {
      clip-path: inset(0 0 100% 0);
    }
  }
  .pregunta {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0 0 6px;
    font-family: var(--titulo);
    font-size: 15px;
    font-weight: 700;
  }
  .pregunta img {
    width: 24px;
  }
  .plantilla {
    position: relative;
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 4px 4px 4px 18px;
    text-align: left;
    background: none;
    border: 0;
    cursor: pointer;
  }
  .plantilla {
    z-index: 0;
    transition: transform 0.12s steps(3);
  }
  .plantilla::after {
    content: '';
    position: absolute;
    inset: 0;
    z-index: -1;
    background: linear-gradient(
      90deg,
      color-mix(in srgb, var(--acento) 26%, transparent),
      transparent 85%
    );
    box-shadow: inset 4px 0 0 var(--acento);
    transform: scaleX(0);
    transform-origin: left;
    transition: transform 0.16s steps(4);
  }
  .plantilla:hover {
    transform: translateX(3px);
  }
  .plantilla:hover::after {
    transform: scaleX(1);
  }
  .plantilla:hover img {
    animation: menear 0.5s steps(2) infinite;
  }
  .nota {
    margin: -2px 0 8px;
    font-size: 12px;
    line-height: 1.3;
    color: var(--texto-suave);
  }
  .nota code {
    font-family: var(--codigo);
    color: var(--palabra);
  }
  .plantilla:hover::before {
    content: '';
    position: absolute;
    left: 4px;
    border-top: 5px solid transparent;
    border-bottom: 5px solid transparent;
    border-left: 8px solid var(--texto);
  }
  .plantilla img {
    width: 26px;
    flex: none;
  }
  .plantilla b {
    display: block;
    font-family: var(--titulo);
    font-size: 15px;
  }
  .plantilla small {
    display: block;
    font-size: 11px;
    line-height: 1.25;
    color: var(--texto-suave);
  }
  .nombrar label {
    display: flex;
    align-items: center;
    gap: 2px;
    margin-bottom: 8px;
    font-family: var(--titulo);
    color: var(--texto-suave);
  }
  .mas.abierto {
    transform: translateY(-1px);
    box-shadow: 0 2px 0 var(--borde);
  }
  .chip {
    flex: none;
    padding: 0 4px;
    font-family: var(--titulo);
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.5px;
    color: var(--acento-texto);
    background: #8a3ab9;
    border: 1px solid var(--borde);
  }
  .sucio {
    width: 8px;
    height: 8px;
    background: var(--acento-2);
    border: 2px solid var(--borde);
  }
  .cuenta {
    min-width: 20px;
    padding: 0 4px;
    font-size: 13px;
    text-align: center;
    color: #fff;
    border: 2px solid var(--borde);
  }
  .cuenta.error {
    background: var(--error);
  }
  .cuenta.aviso {
    background: var(--aviso);
  }
  .tres {
    display: none;
    gap: 2px;
    padding: 4px 2px;
  }
  .tres i {
    width: 3px;
    height: 3px;
    background: var(--texto);
  }
  li:hover .tres {
    display: flex;
  }
  input {
    flex: 1;
    min-width: 0;
    padding: 1px 4px;
    font-family: var(--titulo);
    font-size: 16px;
    color: var(--texto);
    background: var(--editor-fondo);
    border: 2px solid var(--acento);
    outline: none;
    user-select: text;
  }
  .menu {
    position: absolute;
    z-index: 30;
    display: flex;
    flex-direction: column;
    min-width: 170px;
    padding: 0;
  }
  .menu button {
    position: relative;
    padding: 5px 8px 5px 22px;
    font-family: var(--titulo);
    font-size: 15px;
    text-align: left;
    background: none;
    border: 0;
    cursor: pointer;
  }
  .menu button:hover:not(:disabled)::before {
    content: '';
    position: absolute;
    left: 6px;
    top: 50%;
    transform: translateY(-50%);
    border-top: 5px solid transparent;
    border-bottom: 5px solid transparent;
    border-left: 8px solid var(--texto);
  }
  .menu button:disabled {
    opacity: 0.4;
    cursor: default;
  }
  .peligro {
    color: var(--error);
  }
  .confirmar {
    margin: 6px 0;
    padding: 2px 4px;
    font-size: 13px;
  }
  .confirmar p {
    margin: 0 0 6px;
  }
  .si-no {
    display: flex;
    gap: 6px;
  }
  .si-no .boton {
    min-height: 30px;
    font-size: 13px;
  }
  .pista {
    margin: 8px 4px 2px;
    font-size: 12px;
    line-height: 1.35;
    color: var(--texto-suave);
  }
  .pista code {
    font-family: var(--codigo);
    color: var(--palabra);
  }
</style>

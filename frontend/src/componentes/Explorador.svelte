<script>
  // Explorador de carpetas con la forma del PC de los juegos: cada carpeta
  // es una caja, los proyectos llevan su Pokéball y los .pks son Pokéballs
  // sueltas. El fondo de la caja, el nombre del PC y el Pokémon que habla
  // cambian con el tema.
  import { onMount, tick } from 'svelte';
  import { explorador, cerrarExplorador, perfil } from '../lib/estado.svelte.js';
  import { temaActual } from '../lib/temas.js';
  import { api } from '../lib/api.js';
  import { POKEMON, sprite, objeto } from '../lib/pokemon.js';
  import { sonar } from '../lib/sonido.js';

  const PC = {
    'rojo-fuego': 'PC DE BILL',
    'game-boy': 'PC DE BILL',
    'torre-lavanda': 'PC DE LA TORRE',
    'cueva-celeste': 'PC DE LA CUEVA',
    'pueblo-paleta': 'PC DE TU CASA',
    'mar-profundo': 'PC SUBMARINO',
    'llanura-trueno': 'PC DE LA CENTRAL',
    'monte-plateado': 'PC DEL MONTE',
  };

  const tema = $derived(temaActual(perfil.tema));
  const mascota = $derived(tema.escena.mascota);

  let lugares = $state([]);
  let carpeta = $state(null);
  let elegida = $state(null); // { tipo: 'carpeta' | 'archivo', ... }
  let error = $state('');
  let cargando = $state(false);
  let caja = $state();

  const mensaje = $derived.by(() => {
    if (error) return error;
    if (explorador.modo === 'crear')
      return `Elige dónde crear «${explorador.nombre}». Se hará una carpeta nueva con ese nombre.`;
    if (elegida?.tipo === 'carpeta' && elegida.proyecto)
      return `«${elegida.nombre}» es un proyecto con ${elegida.pks} ${elegida.pks === 1 ? 'archivo' : 'archivos'} .pks. ¡Ábrelo!`;
    if (elegida?.tipo === 'archivo' && !elegida.pks)
      return `«${elegida.nombre}» no es de PokeScript: aquí solo se abren carpetas con archivos .pks.`;
    return '¿Qué carpeta quieres abrir? Las que tienen Pokéball ya son proyectos.';
  });

  onMount(async () => {
    sonar('abrir');
    lugares = await api.lugaresExplorador();
    const inicio = perfil.recientes?.[0]?.ruta;
    const padreReciente = inicio ? padreDe(inicio) : null;
    await ir(padreReciente ?? lugares[0]?.ruta);
  });

  function padreDe(ruta) {
    const partes = ruta.replace(/[\\/]+$/, '').split(/[\\/]/);
    return partes.length > 1 ? partes.slice(0, -1).join('/') || '/' : null;
  }

  async function ir(ruta) {
    if (!ruta) return;
    cargando = true;
    error = '';
    try {
      carpeta = await api.listarCarpeta(ruta);
      elegida = null;
      await tick();
      if (caja) caja.scrollTop = 0;
    } catch (e) {
      error = `No se puede entrar ahí: ${e?.message ?? e}`;
      sonar('error');
    } finally {
      cargando = false;
    }
  }

  function tocar(item) {
    if (elegida === item) return;
    elegida = item;
    sonar('mover');
  }

  function doble(item) {
    if (item.tipo === 'carpeta') {
      if (explorador.modo === 'abrir' && item.proyecto) return elegir(item.ruta);
      sonar('abrir');
      ir(item.ruta);
    } else if (item.pks && explorador.modo === 'abrir') {
      elegir(carpeta.ruta);
    }
  }

  function elegir(ruta) {
    sonar('captura');
    cerrarExplorador(ruta);
  }

  function cancelar() {
    sonar('mover');
    cerrarExplorador(null);
  }

  function tecla(e) {
    if (e.key === 'Escape') cancelar();
    else if (e.key === 'Backspace' && carpeta?.padre) ir(carpeta.padre);
    else if (e.key === 'Enter' && elegida) doble(elegida);
  }

  const items = $derived(
    carpeta
      ? [
          ...carpeta.carpetas.map((c) => ({ ...c, tipo: 'carpeta' })),
          ...carpeta.archivos.map((a) => ({ ...a, tipo: 'archivo' })),
        ]
      : [],
  );

  const migas = $derived.by(() => {
    if (!carpeta) return [];
    const partes = carpeta.ruta.split(/[\\/]/).filter(Boolean);
    const sep = carpeta.ruta.includes('\\') ? '\\' : '/';
    const inicio = carpeta.ruta.startsWith('/') ? '/' : '';
    return partes.map((p, i) => ({
      nombre: p,
      ruta: inicio + partes.slice(0, i + 1).join(sep) + (i === 0 && p.endsWith(':') ? sep : ''),
    }));
  });

  const iconoLugar = { personal: 'master-ball', carpeta: 'town-map', unidad: 'exp-share' };
</script>

<svelte:window onkeydown={tecla} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="velo" onclick={cancelar}></div>

<div class="pc" role="dialog" aria-label={PC[perfil.tema] ?? 'PC de Bill'}>
  <header>
    <span class="pantallita"><span></span></span>
    <h2>{PC[perfil.tema] ?? 'PC DE BILL'}</h2>
    <p class="modo">{explorador.modo === 'crear' ? 'GUARDAR PROYECTO' : 'ABRIR PROYECTO'}</p>
    <button class="boton" onclick={cancelar}>SALIR</button>
  </header>

  <div class="cuerpo">
    <nav class="lugares marco">
      <p class="titulo">LUGARES</p>
      {#each lugares as l (l.ruta)}
        <button class:activo={carpeta?.ruta === l.ruta} onclick={() => ir(l.ruta)} title={l.ruta}>
          <img class="pixel" src={objeto(iconoLugar[l.tipo] ?? 'poke-ball')} alt="" />
          <span>{l.nombre}</span>
        </button>
      {/each}
      {#if perfil.recientes?.length}
        <p class="titulo">RECIENTES</p>
        {#each perfil.recientes as r (r.ruta)}
          <button onclick={() => ir(r.ruta)} title={r.ruta}>
            <img class="pixel" src={objeto('poke-ball')} alt="" />
            <span>{r.nombre}</span>
          </button>
        {/each}
      {/if}
    </nav>

    <section class="almacen">
      <div class="cabeza">
        <button
          class="flecha"
          onclick={() => ir(carpeta?.padre)}
          disabled={!carpeta?.padre}
          aria-label="Subir"
          title="Subir (Retroceso)"><i></i></button
        >
        <div class="etiqueta">
          <b>{carpeta?.nombre ?? '…'}</b>
          <small>
            {#each migas as m, i (m.ruta)}
              {#if i}<span class="sep">›</span>{/if}<button onclick={() => ir(m.ruta)}
                >{m.nombre}</button
              >
            {/each}
          </small>
        </div>
        {#if carpeta?.proyecto}<span class="sello">PROYECTO</span>{/if}
      </div>

      <div
        class="caja-marco"
        style="--fondo-caja:url(/fondos/{tema.escena.fondo}.png);--filtro-caja:{tema.escena
          .filtro ??
          tema.filtroSprites ??
          'none'}"
      >
        <div class="fondo-caja"></div>
        <div class="caja" bind:this={caja} class:cargando>
          {#each items as item (item.tipo + item.nombre)}
            <button
              class="hueco {item.tipo}"
              class:proyecto={item.proyecto}
              class:pks={item.pks === true}
              class:elegido={elegida === item}
              onclick={() => tocar(item)}
              ondblclick={() => doble(item)}
              title={item.nombre}
            >
              {#if item.tipo === 'carpeta'}
                <svg
                  class="icono"
                  viewBox="0 0 16 13"
                  shape-rendering="crispEdges"
                  aria-hidden="true"
                >
                  <rect x="0" y="1" width="7" height="3" fill="var(--acento-2)" />
                  <rect x="0" y="3" width="16" height="10" fill="var(--acento-2)" />
                  <rect
                    x="0"
                    y="5"
                    width="16"
                    height="8"
                    fill="color-mix(in srgb, var(--acento-2) 72%, #000)"
                  />
                  <rect x="1" y="6" width="14" height="1" fill="rgba(255,255,255,.35)" />
                </svg>
                {#if item.proyecto}
                  <img class="pixel sobre" src={objeto('poke-ball')} alt="" />
                {/if}
              {:else if item.pks}
                <img class="pixel bola" src={objeto('poke-ball')} alt="" />
              {:else}
                <span class="hoja" aria-hidden="true"></span>
              {/if}
              <span class="nombre">{item.nombre}</span>
            </button>
          {:else}
            <p class="vacia">{cargando ? 'Encendiendo…' : 'Esta caja está vacía.'}</p>
          {/each}
          {#if carpeta?.recortada}
            <p class="vacia">Hay más cosas en esta carpeta de las que caben aquí.</p>
          {/if}
        </div>
      </div>
    </section>
  </div>

  <footer>
    <div class="guia marco">
      <img class="sprite" src={sprite(mascota)} alt="" />
      <p><b>{POKEMON[mascota]?.nombre}:</b> {mensaje}</p>
    </div>
    <div class="acciones">
      {#if explorador.modo === 'crear'}
        <p class="destino">
          Se creará en: <code>{carpeta?.ruta ?? '…'}</code>
        </p>
        <button class="boton principal" disabled={!carpeta} onclick={() => elegir(carpeta.ruta)}
          >CREAR AQUÍ</button
        >
      {:else}
        {#if elegida?.tipo === 'carpeta'}
          <button class="boton principal" onclick={() => elegir(elegida.ruta)}
            >ABRIR «{elegida.nombre}»</button
          >
        {/if}
        <button
          class="boton"
          class:principal={!elegida && carpeta?.proyecto}
          disabled={!carpeta}
          onclick={() => elegir(carpeta.ruta)}>ABRIR ESTA CARPETA</button
        >
      {/if}
    </div>
  </footer>
</div>

<style>
  .velo {
    position: fixed;
    inset: 0;
    z-index: 60;
    background: rgba(8, 6, 16, 0.6);
    animation: aparecer 0.2s steps(3) both;
  }
  @keyframes aparecer {
    from {
      opacity: 0;
    }
  }
  .pc {
    position: fixed;
    inset: 4vh 4vw;
    z-index: 61;
    display: flex;
    flex-direction: column;
    color: var(--texto);
    background: var(--fondo-2);
    border: 4px solid var(--borde);
    border-radius: 14px;
    box-shadow:
      inset 0 0 0 4px var(--borde-suave),
      0 12px 0 var(--borde),
      0 24px 50px rgba(0, 0, 0, 0.5);
    animation: encender 0.4s steps(6) both;
  }
  @keyframes encender {
    from {
      clip-path: inset(49% 0 49% 0);
    }
    50% {
      clip-path: inset(49% 0 49% 0);
    }
  }
  header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: var(--acento);
    border-bottom: 4px solid var(--borde);
    border-radius: 10px 10px 0 0;
  }
  .pantallita {
    display: grid;
    place-items: center;
    width: 40px;
    height: 32px;
    background: #10231a;
    border: 3px solid var(--borde);
  }
  .pantallita span {
    width: 18px;
    height: 4px;
    background: #7fffb0;
    animation: parpadeo 1.2s steps(1) infinite;
  }
  h2 {
    margin: 0;
    font-family: var(--titulo);
    font-size: 26px;
    letter-spacing: 1px;
    color: var(--acento-texto);
    text-shadow: 2px 2px 0 rgba(0, 0, 0, 0.3);
  }
  .modo {
    margin: 0 auto 0 8px;
    font-family: var(--titulo);
    font-size: 13px;
    letter-spacing: 2px;
    color: var(--acento-texto);
    opacity: 0.8;
  }
  .cuerpo {
    flex: 1;
    display: grid;
    grid-template-columns: minmax(170px, 230px) minmax(0, 1fr);
    gap: 14px;
    min-height: 0;
    padding: 14px;
  }
  .lugares {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-height: 0;
    overflow-y: auto;
    padding: 0 2px;
  }
  .lugares .titulo {
    margin: 6px 4px 4px;
    font-family: var(--titulo);
    font-size: 12px;
    letter-spacing: 2px;
    color: var(--texto-suave);
  }
  .lugares button {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 6px;
    font-family: var(--titulo);
    font-size: 14px;
    text-align: left;
    background: none;
    border: 2px solid transparent;
    cursor: pointer;
  }
  .lugares button:hover {
    background: var(--panel-2);
  }
  .lugares button.activo {
    background: var(--panel-2);
    border-color: var(--borde);
  }
  .lugares img {
    width: 22px;
    flex: none;
  }
  .lugares span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .almacen {
    display: flex;
    flex-direction: column;
    min-height: 0;
    min-width: 0;
  }
  .cabeza {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 8px;
  }
  .flecha {
    display: grid;
    place-items: center;
    width: 38px;
    height: 38px;
    flex: none;
    background: var(--panel);
    border: 3px solid var(--borde);
    cursor: pointer;
  }
  .flecha:disabled {
    opacity: 0.4;
    cursor: default;
  }
  .flecha i {
    border-left: 8px solid transparent;
    border-right: 8px solid transparent;
    border-bottom: 11px solid var(--texto);
  }
  /* La etiqueta con el nombre de la caja, como en el PC de los juegos. */
  .etiqueta {
    flex: 1;
    min-width: 0;
    padding: 4px 14px;
    text-align: center;
    background: var(--panel);
    border: 3px solid var(--borde);
    border-radius: 18px;
    box-shadow: inset 0 -4px 0 var(--borde-suave);
  }
  .etiqueta b {
    display: block;
    font-family: var(--titulo);
    font-size: 20px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .etiqueta small {
    display: block;
    font-size: 12px;
    color: var(--texto-suave);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .etiqueta small button {
    padding: 0;
    font: inherit;
    color: inherit;
    background: none;
    border: 0;
    cursor: pointer;
  }
  .etiqueta small button:hover {
    color: var(--texto);
    text-decoration: underline;
  }
  .sep {
    margin: 0 4px;
  }
  .sello {
    padding: 2px 8px;
    font-family: var(--titulo);
    font-size: 12px;
    font-weight: 700;
    color: #fff;
    background: #8a3ab9;
    border: 2px solid var(--borde);
  }
  /* La caja: su fondo es el lugar del tema, como los fondos de caja del PC. */
  .caja-marco {
    position: relative;
    flex: 1;
    min-height: 0;
    overflow: hidden;
    border: 4px solid var(--borde);
    border-radius: 10px;
  }
  .fondo-caja {
    position: absolute;
    inset: 0;
    background: var(--fondo-caja) center / cover;
    image-rendering: pixelated;
    filter: var(--filtro-caja);
  }
  .caja {
    position: absolute;
    inset: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(104px, 1fr));
    grid-auto-rows: 118px;
    gap: 8px;
    align-content: start;
    padding: 18px 12px 12px;
    overflow-y: auto;
  }
  .caja.cargando {
    opacity: 0.6;
  }
  .hueco {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-end;
    gap: 6px;
    padding: 8px 4px 6px;
    background: color-mix(in srgb, var(--panel) 50%, transparent);
    border: 3px solid transparent;
    border-radius: 8px;
    cursor: pointer;
    transition: transform 0.08s steps(2);
  }
  .hueco:hover {
    transform: translateY(-3px);
    border-color: var(--borde-suave);
  }
  .hueco.elegido {
    background: var(--panel);
    border-color: var(--borde);
    box-shadow: 0 4px 0 var(--borde);
  }
  .hueco.elegido::before {
    content: '';
    position: absolute;
    top: -14px;
    left: 50%;
    transform: translateX(-50%);
    border-left: 8px solid transparent;
    border-right: 8px solid transparent;
    border-top: 11px solid var(--texto);
    animation: apuntar 0.6s steps(2) infinite;
  }
  @keyframes apuntar {
    50% {
      top: -10px;
    }
  }
  .icono {
    width: 56px;
    height: 46px;
  }
  .sobre {
    position: absolute;
    top: 28px;
    right: 18px;
    width: 30px;
    animation: saltito 1.4s steps(2) infinite;
  }
  @keyframes saltito {
    50% {
      transform: translateY(-3px);
    }
  }
  .bola {
    width: 46px;
  }
  .hueco.archivo:not(.pks) {
    opacity: 0.6;
  }
  .hoja {
    width: 34px;
    height: 44px;
    background: var(--panel-2);
    border: 3px solid var(--texto-suave);
    clip-path: polygon(0 0, 65% 0, 100% 28%, 100% 100%, 0 100%);
  }
  .nombre {
    width: 100%;
    font-family: var(--titulo);
    font-size: 13px;
    line-height: 1.15;
    text-align: center;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    word-break: break-word;
  }
  .vacia {
    grid-column: 1 / -1;
    margin: 20px;
    padding: 10px;
    font-family: var(--titulo);
    text-align: center;
    background: color-mix(in srgb, var(--panel) 80%, transparent);
  }
  footer {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 14px;
    align-items: center;
    padding: 0 14px 14px;
  }
  .guia {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 72px;
    padding: 0 6px;
  }
  .guia img {
    width: 56px;
    height: 56px;
    object-fit: contain;
    flex: none;
  }
  .guia p {
    margin: 0;
    font-family: var(--cuerpo);
    font-size: 15px;
    line-height: 1.4;
  }
  .acciones {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 8px;
  }
  .destino {
    max-width: 380px;
    margin: 0;
    font-size: 13px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .destino code {
    font-family: var(--codigo);
  }
  @media (max-width: 1100px) {
    .cuerpo {
      grid-template-columns: 58px minmax(0, 1fr);
    }
    .lugares span,
    .lugares .titulo {
      display: none;
    }
  }
</style>

<script>
  // Pantalla de título: las portadas van rotando (un lugar y su Pokémon) y,
  // después de «PRESIONA CUALQUIER TECLA», el menú de los juegos para
  // continuar, crear o abrir un proyecto.
  import { onMount, tick } from 'svelte';
  import Escena from './Escena.svelte';
  import { perfil, cambiarTema, cambiarSonido, navegacion } from '../lib/estado.svelte.js';
  import { TEMAS } from '../lib/temas.js';
  import { POKEMON, sprite } from '../lib/pokemon.js';
  import { PORTADAS, portadaInicial } from '../lib/portadas.js';
  import { crearProyecto, buscarProyecto, proyectosConocidos } from '../lib/acciones.js';
  import { api } from '../lib/api.js';
  import { sonar } from '../lib/sonido.js';

  // alContinuar recibe { ruta } para abrir un proyecto o { accion: 'oak' }.
  let { alContinuar, conMenu = false } = $props();

  let indice = $state(portadaInicial());
  let cambiando = $state(false);
  const portada = $derived(PORTADAS[indice]);

  let fase = $state(conMenu ? 'menu' : 'presiona'); // presiona → menu
  let vista = $state('principal'); // principal | nuevo | abrir | opciones
  let cursor = $state(0);
  let nombre = $state('');
  let error = $state('');
  let listo = false;

  const ultimo = $derived(perfil.recientes?.[0]);
  const conocidos = $derived(proyectosConocidos());
  const temas = Object.keys(TEMAS);

  const OPCIONES = $derived.by(() => {
    if (vista === 'principal')
      return [
        ...(ultimo ? [{ id: 'continuar', texto: 'CONTINUAR' }] : []),
        { id: 'nuevo', texto: 'NUEVO PROYECTO' },
        { id: 'abrir', texto: 'ABRIR PROYECTO' },
        { id: 'opciones', texto: 'OPCIONES' },
      ];
    if (vista === 'abrir')
      return [
        ...conocidos.map((p) => ({ id: 'proyecto', texto: p.nombre, ruta: p.ruta })),
        ...(api.enWails() ? [{ id: 'buscar', texto: 'BUSCAR CARPETA…' }] : []),
        { id: 'volver', texto: 'VOLVER' },
      ];
    if (vista === 'opciones')
      return [
        { id: 'tema', texto: `TEMA: ${TEMAS[perfil.tema].nombre.toUpperCase()}` },
        { id: 'sonido', texto: `SONIDOS: ${perfil.sonido ? 'SÍ' : 'NO'}` },
        { id: 'oak', texto: 'VER A OAK OTRA VEZ' },
        { id: 'volver', texto: 'VOLVER' },
      ];
    return [];
  });

  onMount(() => {
    const listoId = setTimeout(() => (listo = true), 700);
    const rotar = setInterval(() => {
      cambiando = true;
      setTimeout(() => {
        indice = (indice + 1) % PORTADAS.length;
        cambiando = false;
      }, 450);
    }, 12000);
    return () => {
      clearTimeout(listoId);
      clearInterval(rotar);
    };
  });

  function ir(v) {
    vista = v;
    cursor = 0;
    error = '';
  }

  async function elegir(op) {
    sonar('elegir');
    navegacion.aviso = '';
    switch (op.id) {
      case 'continuar':
        return alContinuar({ ruta: ultimo.ruta });
      case 'nuevo':
        ir('nuevo');
        nombre = 'mi_proyecto';
        await tick();
        document.querySelector('.titulo .campo')?.select();
        return;
      case 'abrir':
      case 'opciones':
        return ir(op.id);
      case 'volver':
        return ir('principal');
      case 'proyecto':
        return alContinuar({ ruta: op.ruta });
      case 'buscar': {
        const ruta = await buscarProyecto();
        if (ruta) alContinuar({ ruta });
        return;
      }
      case 'tema':
        return cambiarTema(temas[(temas.indexOf(perfil.tema) + 1) % temas.length]);
      case 'sonido':
        return cambiarSonido(!perfil.sonido);
      case 'oak':
        return alContinuar({ accion: 'oak' });
    }
  }

  async function crear(e) {
    e.preventDefault();
    if (!nombre.trim()) return;
    try {
      const ruta = await crearProyecto(nombre);
      if (ruta) alContinuar({ ruta });
    } catch (err) {
      sonar('error');
      error = `No se pudo: ${err?.message ?? err}`;
    }
  }

  function tecla(e) {
    if (!listo || e.repeat) return;
    if (fase === 'presiona') {
      sonar('elegir');
      if (!perfil.companero) return alContinuar({ accion: 'oak' });
      fase = 'menu';
      return;
    }
    if (vista === 'nuevo') {
      if (e.key === 'Escape') ir('principal');
      return;
    }
    if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
      e.preventDefault();
      cursor = (cursor + (e.key === 'ArrowDown' ? 1 : -1) + OPCIONES.length) % OPCIONES.length;
      sonar('mover');
    } else if (e.key === 'Enter' || e.key === 'z') {
      elegir(OPCIONES[cursor]);
    } else if (e.key === 'Escape' && vista !== 'principal') {
      sonar('mover');
      ir('principal');
    } else if (vista === 'opciones' && OPCIONES[cursor].id === 'tema' && e.key === 'ArrowRight') {
      elegir(OPCIONES[cursor]);
    }
  }
</script>

<svelte:window onkeydown={tecla} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="titulo" onclick={() => fase === 'presiona' && tecla({ key: 'Enter' })}>
  <Escena oscurecer={fase === 'menu' ? 0.25 : 0.1} {portada}>
    {#key indice}
      <div class="mascota" class:vuela={portada.vuela}>
        <img class="sprite" src={sprite(portada.mascota)} alt="" draggable="false" />
        {#if !portada.vuela}<span class="sombra"></span>{/if}
      </div>
    {/key}

    <div class="logo" class:arriba={fase === 'menu'}>
      <h1>PokeScript</h1>
      <div class="cinta">EDICIÓN PARADIGMAS · UNA 2026</div>
    </div>

    {#if fase === 'presiona'}
      <div class="presiona marco oscuro">PRESIONA CUALQUIER TECLA</div>
    {:else}
      <div class="menu">
        {#if vista === 'nuevo'}
          <form class="caja marco" onsubmit={crear}>
            <p class="titulo-caja">NUEVO PROYECTO</p>
            <label>
              Nombre
              <input class="campo" bind:value={nombre} maxlength="30" spellcheck="false" />
            </label>
            {#if api.enWails()}
              <small>Después eliges en qué carpeta guardarlo.</small>
            {/if}
            {#if error}<small class="error">{error}</small>{/if}
            <div class="botones">
              <button class="boton principal" type="submit">CREAR</button>
              <button class="boton" type="button" onclick={() => ir('principal')}>VOLVER</button>
            </div>
          </form>
        {:else}
          <div class="caja marco">
            {#if navegacion.aviso}<p class="aviso">{navegacion.aviso}</p>{/if}
            {#if vista !== 'principal'}
              <p class="titulo-caja">{vista === 'abrir' ? 'ABRIR PROYECTO' : 'OPCIONES'}</p>
            {/if}
            {#each OPCIONES as op, i (op.id + (op.ruta ?? '') + i)}
              <button
                class="opcion"
                class:sel={i === cursor}
                onmouseenter={() => {
                  if (cursor !== i) sonar('mover');
                  cursor = i;
                }}
                onclick={() => elegir(op)}
              >
                {op.texto}
              </button>
              {#if op.id === 'continuar' && ultimo}
                <div class="ficha">
                  <img class="sprite" src={sprite(perfil.companero)} alt="" />
                  <dl>
                    <dt>ENTRENADOR</dt>
                    <dd>{perfil.nombre}</dd>
                    <dt>PROYECTO</dt>
                    <dd>{ultimo.nombre}</dd>
                    <dt>COMPAÑERO</dt>
                    <dd>
                      {POKEMON[perfil.companero]?.nombre} Nv{5 + Math.floor((perfil.exp ?? 0) / 3)}
                    </dd>
                  </dl>
                </div>
              {/if}
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </Escena>
  <div class="telon"></div>
  {#if cambiando}<div class="cambio"></div>{/if}
</div>

<style>
  .titulo {
    position: fixed;
    inset: 0;
  }
  .telon {
    position: fixed;
    inset: 0;
    background: #000;
    pointer-events: none;
    animation: abrir 0.9s steps(6) forwards;
  }
  @keyframes abrir {
    to {
      opacity: 0;
    }
  }
  /* Cambio de portada: la pantalla se funde a blanco y vuelve. */
  .cambio {
    position: fixed;
    inset: 0;
    background: #fff;
    pointer-events: none;
    animation: fundir 0.9s steps(6) both;
  }
  @keyframes fundir {
    0%,
    100% {
      opacity: 0;
    }
    50% {
      opacity: 1;
    }
  }
  .logo {
    position: absolute;
    top: 9vh;
    left: 0;
    right: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 14px;
    animation: bajar 0.8s 0.5s steps(8) both;
    transition: top 0.4s steps(6);
  }
  .logo.arriba {
    top: 4vh;
  }
  @keyframes bajar {
    from {
      transform: translateY(-40vh);
    }
    80% {
      transform: translateY(12px);
    }
  }
  h1 {
    margin: 0;
    font-family: var(--titulo);
    font-size: clamp(64px, 10vw, 132px);
    font-weight: 700;
    line-height: 1;
    letter-spacing: 3px;
    color: #ffcb05;
    text-shadow:
      5px 0 0 #2a4fa0,
      -5px 0 0 #2a4fa0,
      0 5px 0 #2a4fa0,
      0 -5px 0 #2a4fa0,
      5px 5px 0 #2a4fa0,
      -5px 5px 0 #2a4fa0,
      5px -5px 0 #2a4fa0,
      -5px -5px 0 #2a4fa0,
      0 12px 0 #14265a,
      5px 12px 0 #14265a,
      -5px 12px 0 #14265a;
  }
  .cinta {
    padding: 6px 18px;
    font-family: var(--titulo);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: 3px;
    color: #fff;
    background: var(--acento);
    border: 3px solid #14141c;
    box-shadow:
      inset 0 -4px 0 rgba(0, 0, 0, 0.25),
      0 4px 0 #14141c;
  }
  .mascota {
    position: absolute;
    right: 12vw;
    bottom: 12vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    animation: llegar 0.9s 0.2s steps(10) both;
  }
  @keyframes llegar {
    from {
      transform: translateX(40vw);
    }
  }
  .mascota img {
    zoom: 3;
  }
  .mascota.vuela {
    bottom: auto;
    top: 34vh;
    right: 8vw;
  }
  .mascota.vuela img {
    animation: planear 2.4s steps(6) infinite;
  }
  @keyframes planear {
    50% {
      transform: translateY(-8px);
    }
  }
  .sombra {
    width: 180px;
    height: 26px;
    margin-top: -18px;
    background: rgba(0, 0, 0, 0.35);
    border-radius: 50%;
  }
  .presiona {
    position: absolute;
    left: 50%;
    bottom: 7vh;
    transform: translateX(-50%);
    padding: 4px 18px;
    font-family: var(--titulo);
    font-size: 22px;
    font-weight: 600;
    letter-spacing: 3px;
    white-space: nowrap;
    cursor: pointer;
    animation: parpadeo 1.2s steps(1) infinite;
  }
  .menu {
    position: absolute;
    left: 6vw;
    top: 36vh;
    width: min(440px, 44vw);
    animation: entrar 0.3s steps(4) both;
  }
  @keyframes entrar {
    from {
      clip-path: inset(0 100% 0 0);
    }
  }
  .caja {
    display: flex;
    flex-direction: column;
    padding: 2px 4px 4px;
    max-height: 56vh;
    overflow-y: auto;
  }
  .titulo-caja {
    margin: 0 0 6px;
    font-family: var(--titulo);
    font-size: 15px;
    font-weight: 700;
    letter-spacing: 2px;
    color: var(--texto-suave);
  }
  .opcion {
    position: relative;
    padding: 6px 10px 6px 30px;
    font-family: var(--titulo);
    font-size: 26px;
    font-weight: 700;
    text-align: left;
    letter-spacing: 1px;
    background: none;
    border: 0;
    cursor: pointer;
  }
  .opcion.sel::before {
    content: '';
    position: absolute;
    left: 8px;
    top: 50%;
    transform: translateY(-50%);
    border-top: 9px solid transparent;
    border-bottom: 9px solid transparent;
    border-left: 13px solid var(--texto);
    animation: apuntar 0.6s steps(2) infinite;
  }
  @keyframes apuntar {
    50% {
      transform: translate(3px, -50%);
    }
  }
  .ficha {
    display: flex;
    gap: 12px;
    align-items: center;
    margin: 0 0 8px 30px;
    padding: 8px 10px;
    background: var(--panel-2);
    border: 2px solid var(--borde-suave);
  }
  .ficha img {
    zoom: 1.4;
  }
  dl {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 2px 12px;
    margin: 0;
    font-family: var(--titulo);
    font-size: 15px;
  }
  dt {
    color: var(--texto-suave);
  }
  dd {
    margin: 0;
    font-weight: 700;
  }
  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin: 4px 6px 8px;
    font-family: var(--titulo);
    font-size: 16px;
    color: var(--texto-suave);
  }
  input {
    padding: 6px 10px;
    font-family: var(--titulo);
    font-size: 22px;
    color: var(--texto);
    background: var(--panel-2);
    border: 3px solid var(--borde);
    outline: none;
    user-select: text;
  }
  small {
    margin: 0 6px 8px;
    font-size: 13px;
    color: var(--texto-suave);
  }
  .error {
    color: var(--error);
  }
  .aviso {
    margin: 0 0 6px;
    padding: 4px 8px;
    font-size: 14px;
    color: #fff;
    background: var(--error);
    border: 2px solid var(--borde);
  }
  .botones {
    display: flex;
    gap: 10px;
    margin: 4px 6px;
  }
</style>

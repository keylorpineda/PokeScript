<script>
  // Pokédex de consulta (sección 9: el menú de consulta del IDE). Todo el
  // lenguaje por secciones, con un Pokémon que explica cada entrada, la tabla
  // de efectividades para explorar y las 49 palabras reservadas.
  import Sombra from './Sombra.svelte';
  import { onMount, tick } from 'svelte';
  import Codigo from './Codigo.svelte';
  import Pokebola from './Pokebola.svelte';
  import { ide } from '../lib/estado.svelte.js';
  import { api } from '../lib/api.js';
  import { POKEMON, sprite, objeto } from '../lib/pokemon.js';
  import { SECCIONES, ENTRADAS, POKEMON_GRUPO, PRECEDENCIA, EFECTIVIDAD } from '../lib/pokedex.js';
  import { sonar } from '../lib/sonido.js';

  let palabras = $state([]);
  let tabla = $state([]);
  let seccion = $state('inicio');
  let indice = $state(0);
  let busqueda = $state('');
  let casilla = $state(null); // { origen, destino, valor } bajo el mouse
  let copiado = $state(false);
  let lista = $state();

  onMount(async () => {
    [palabras, tabla] = await Promise.all([api.palabrasReservadas(), api.tablaEfectividades()]);
  });

  // Las palabras reservadas se vuelven entradas, con el Pokémon de su grupo.
  const todas = $derived([
    ...ENTRADAS,
    ...palabras.map((p) => ({
      seccion: 'palabras',
      titulo: p.palabra,
      grupo: p.grupo,
      pokemon: POKEMON_GRUPO[p.grupo] ?? 'porygon',
      texto: p.descripcion,
      ejemplo: p.ejemplo,
    })),
  ]);

  const normal = (t) => (t ?? '').toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '');

  const visibles = $derived.by(() => {
    const q = normal(busqueda.trim());
    if (!q) return todas.filter((e) => e.seccion === seccion);
    return todas.filter((e) =>
      normal([e.titulo, e.texto, e.ejemplo, ...(e.notas ?? [])].join(' ')).includes(q),
    );
  });
  const entrada = $derived(visibles[Math.min(indice, visibles.length - 1)]);
  const numero = $derived(entrada ? todas.indexOf(entrada) + 1 : 0);
  const cuenta = (id) => todas.filter((e) => e.seccion === id).length;

  function elegirSeccion(id) {
    busqueda = '';
    seccion = id;
    indice = 0;
    sonar('abrir');
  }

  async function elegir(i) {
    if (i < 0 || i >= visibles.length) return;
    indice = i;
    sonar('mover');
    await tick();
    lista?.querySelector('.activa')?.scrollIntoView({ block: 'nearest' });
  }

  function cerrar() {
    sonar('mover');
    ide.pokedex = false;
  }

  function tecla(e) {
    if (e.key === 'Escape') return cerrar();
    if (e.target instanceof HTMLInputElement && !['ArrowUp', 'ArrowDown'].includes(e.key)) return;
    const s = SECCIONES.findIndex((x) => x.id === seccion);
    if (e.key === 'ArrowDown') elegir(indice + 1);
    else if (e.key === 'ArrowUp') elegir(indice - 1);
    else if (e.key === 'ArrowRight') elegirSeccion(SECCIONES[(s + 1) % SECCIONES.length].id);
    else if (e.key === 'ArrowLeft')
      elegirSeccion(SECCIONES[(s - 1 + SECCIONES.length) % SECCIONES.length].id);
    else return;
    e.preventDefault();
  }

  async function copiar() {
    try {
      await navigator.clipboard.writeText(entrada.ejemplo);
      copiado = true;
      sonar('elegir');
      setTimeout(() => (copiado = false), 1500);
    } catch {
      // Sin portapapeles: no pasa nada.
    }
  }

  // Texto con `código` entre comillas invertidas.
  const trozos = (t) => (t ?? '').split('`').map((p, i) => ({ codigo: i % 2 === 1, texto: p }));

  const COLOR_TIPO = {
    roca: 'var(--t-roca)',
    agua: 'var(--t-agua)',
    fuego: 'var(--t-fuego)',
    planta: 'var(--t-planta)',
    electrico: 'var(--t-electrico)',
  };
</script>

<svelte:window onkeydown={tecla} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="velo" onclick={cerrar}></div>

<div class="aparato" role="dialog" aria-label="Pokédex de PokeScript">
  <header>
    <span class="lente"><span class="brillo"></span></span>
    <span class="led rojo"></span><span class="led amarillo"></span><span class="led verde"></span>
    <h2>POKÉDEX <small>· todo sobre PokeScript</small></h2>
    <label class="buscar">
      <img class="pixel" src={objeto('town-map')} alt="" />
      <input
        bind:value={busqueda}
        oninput={() => (indice = 0)}
        placeholder="Buscar: mochila, segun, convertir…"
        spellcheck="false"
      />
    </label>
    <button class="cerrar" onclick={cerrar} aria-label="Cerrar">CERRAR</button>
  </header>

  <div class="cuerpo">
    <!-- Secciones -->
    <nav class="secciones">
      {#each SECCIONES as s (s.id)}
        <button
          class:activa={!busqueda && s.id === seccion}
          onclick={() => elegirSeccion(s.id)}
          title={s.nombre}
        >
          <img class="pixel" src={objeto(s.objeto)} alt="" />
          <span>{s.nombre}</span>
          <small>{cuenta(s.id)}</small>
        </button>
      {/each}
    </nav>

    <!-- Lista de entradas -->
    <ul class="lista" bind:this={lista}>
      {#if busqueda}
        <li class="grupo">{visibles.length} resultado{visibles.length === 1 ? '' : 's'}</li>
      {/if}
      {#each visibles as e, i (e.seccion + e.titulo)}
        {#if !busqueda && e.grupo && e.grupo !== visibles[i - 1]?.grupo}
          <li class="grupo">{e.grupo}</li>
        {/if}
        <li>
          <button class:activa={i === indice} onclick={() => elegir(i)}>
            <span class="n">{String(todas.indexOf(e) + 1).padStart(3, '0')}</span>
            <img class="sprite" src={sprite(e.pokemon)} alt="" />
            <span class="t">{e.titulo}</span>
          </button>
        </li>
      {:else}
        <li class="vacio">Ningún Pokémon sabe de eso. Prueba con otra palabra.</li>
      {/each}
    </ul>

    <!-- Pantalla de la entrada -->
    <section class="pantalla">
      {#if entrada}
        {#key entrada.seccion + entrada.titulo}
          <div class="entrada">
            <div class="retrato">
              <div class="escenario">
                <img
                  class="sprite"
                  src={sprite(entrada.pokemon)}
                  alt={POKEMON[entrada.pokemon]?.nombre}
                />
                <Sombra ancho={110} />
              </div>
              <p class="quien">{POKEMON[entrada.pokemon]?.nombre ?? ''} explica</p>
            </div>

            <div class="datos">
              <p class="numero">Nº {String(numero).padStart(3, '0')}</p>
              <h3>{entrada.titulo}</h3>
              <div class="chips">
                {#if entrada.tipo}
                  <span
                    class="chip"
                    style="background:{COLOR_TIPO[entrada.tipo] ?? 'var(--acento)'}"
                    >TIPO {entrada.tipo.toUpperCase()}</span
                  >
                {/if}
                {#if entrada.grupo}<span class="chip gris">{entrada.grupo}</span>{/if}
                <span class="chip gris"
                  >{SECCIONES.find((s) => s.id === entrada.seccion)?.nombre}</span
                >
              </div>
              <p class="texto">
                {#each trozos(entrada.texto) as t, j (j)}{#if t.codigo}<code>{t.texto}</code
                    >{:else}{t.texto}{/if}{/each}
              </p>
            </div>
          </div>

          {#if entrada.especial === 'tabla' && tabla.length}
            <div class="tabla">
              <table>
                <thead>
                  <tr>
                    {#each tabla[0] as c, j (j)}<th>{j ? c : 'de ↓ a →'}</th>{/each}
                  </tr>
                </thead>
                <tbody>
                  {#each tabla.slice(1) as fila (fila[0])}
                    <tr>
                      <th>{fila[0]}</th>
                      {#each fila.slice(1) as valor, j (j)}
                        <td
                          class={EFECTIVIDAD[valor]?.clase}
                          onmouseenter={() =>
                            (casilla = { origen: fila[0], destino: tabla[0][j + 1], valor })}
                        >
                          {valor}
                        </td>
                      {/each}
                    </tr>
                  {/each}
                </tbody>
              </table>
              <div class="lectura marco">
                {#if casilla}
                  <b>{casilla.origen} → {casilla.destino}</b>
                  <span class="chip {EFECTIVIDAD[casilla.valor].clase}"
                    >{EFECTIVIDAD[casilla.valor].nombre}</span
                  >
                  <p>
                    {#each trozos(EFECTIVIDAD[casilla.valor].texto) as t, j (j)}{#if t.codigo}<code
                          >{t.texto}</code
                        >{:else}{t.texto}{/if}{/each}
                  </p>
                {:else}
                  <p>Pasa el mouse sobre una casilla: fila de origen, columna de destino.</p>
                {/if}
              </div>
            </div>
          {/if}

          {#if entrada.especial === 'precedencia'}
            <ol class="escalera">
              {#each PRECEDENCIA as [op, que], j (op)}
                <li style="--n:{j}">
                  <code>{op}</code><span>{que}</span>
                </li>
              {/each}
            </ol>
          {/if}

          {#if entrada.ejemplo}
            <div class="ejemplo">
              <div class="cabeza">
                <span>EJEMPLO</span>
                <button class="copiar" onclick={copiar}>{copiado ? '¡COPIADO!' : 'COPIAR'}</button>
              </div>
              <Codigo codigo={entrada.ejemplo} />
            </div>
          {/if}

          {#if entrada.notas?.length}
            <ul class="notas">
              {#each entrada.notas as nota, j (j)}
                <li>
                  <Pokebola escala={0.55} />
                  <span
                    >{#each trozos(nota) as t, k (k)}{#if t.codigo}<code>{t.texto}</code
                        >{:else}{t.texto}{/if}{/each}</span
                  >
                </li>
              {/each}
            </ul>
          {/if}
        {/key}

        <footer>
          <button class="boton" onclick={() => elegir(indice - 1)} disabled={indice === 0}
            >ANTERIOR</button
          >
          <span>{indice + 1} de {visibles.length}</span>
          <button
            class="boton"
            onclick={() => elegir(indice + 1)}
            disabled={indice >= visibles.length - 1}>SIGUIENTE</button
          >
        </footer>
      {/if}
    </section>
  </div>
</div>

<style>
  .velo {
    position: fixed;
    inset: 0;
    z-index: 44;
    background: rgba(8, 6, 16, 0.6);
    animation: aparecer 0.2s steps(3) both;
  }
  @keyframes aparecer {
    from {
      opacity: 0;
    }
  }
  .aparato {
    position: fixed;
    inset: 3vh 3vw;
    z-index: 45;
    display: flex;
    flex-direction: column;
    background: linear-gradient(180deg, #e3223a, #b3101f);
    border: 4px solid #5a0610;
    border-radius: 20px;
    box-shadow:
      inset 0 4px 0 rgba(255, 255, 255, 0.25),
      0 10px 0 #5a0610,
      0 20px 40px rgba(0, 0, 0, 0.5);
    animation: abrirse 0.35s steps(6) both;
  }
  @keyframes abrirse {
    from {
      clip-path: inset(0 50% 0 50%);
    }
  }
  header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 18px;
    border-bottom: 4px solid #5a0610;
  }
  .lente {
    position: relative;
    width: 46px;
    height: 46px;
    flex: none;
    background: radial-gradient(circle at 35% 35%, #b8e4ff, #3d9cf0 45%, #15508f);
    border: 4px solid #f4f4f4;
    border-radius: 50%;
    box-shadow: 0 0 0 3px #5a0610;
  }
  .brillo {
    position: absolute;
    left: 8px;
    top: 7px;
    width: 10px;
    height: 10px;
    background: #fff;
    border-radius: 50%;
    opacity: 0.8;
    animation: parpadeo 2.4s steps(1) infinite;
  }
  .led {
    width: 12px;
    height: 12px;
    flex: none;
    border: 2px solid #5a0610;
    border-radius: 50%;
  }
  .rojo {
    background: #ff5f57;
  }
  .amarillo {
    background: #ffcb05;
  }
  .verde {
    background: #5fd35f;
  }
  h2 {
    margin: 0 0 0 6px;
    font-family: 'Pixelify Sans', sans-serif;
    font-size: 26px;
    color: #fff;
    text-shadow: 0 3px 0 #5a0610;
    white-space: nowrap;
  }
  h2 small {
    font-size: 15px;
    opacity: 0.85;
  }
  .buscar {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
    max-width: 440px;
    margin-left: auto;
    padding: 4px 10px;
    background: #fff8f0;
    border: 3px solid #5a0610;
    border-radius: 8px;
  }
  .buscar img {
    width: 22px;
  }
  .buscar input {
    flex: 1;
    min-width: 0;
    font-family: var(--cuerpo);
    font-size: 15px;
    color: #2a1a1a;
    background: none;
    border: 0;
    outline: none;
    user-select: text;
  }
  .cerrar {
    padding: 6px 12px;
    font-family: 'Pixelify Sans', sans-serif;
    font-size: 15px;
    font-weight: 700;
    color: #fff;
    background: #5a0610;
    border: 3px solid #2a0206;
    border-radius: 8px;
    cursor: pointer;
  }
  .cuerpo {
    flex: 1;
    display: grid;
    grid-template-columns: minmax(170px, 220px) minmax(200px, 260px) minmax(0, 1fr);
    gap: 12px;
    min-height: 0;
    padding: 14px;
  }
  .secciones,
  .lista,
  .pantalla {
    min-height: 0;
    overflow-y: auto;
    border: 4px solid #5a0610;
    border-radius: 10px;
  }
  .secciones {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 6px;
    background: #8e0c1a;
  }
  .secciones button {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    font-family: var(--titulo);
    font-size: 15px;
    text-align: left;
    color: #ffe4e4;
    background: none;
    border: 2px solid transparent;
    border-radius: 6px;
    cursor: pointer;
  }
  .secciones button:hover {
    background: rgba(255, 255, 255, 0.1);
  }
  .secciones button.activa {
    color: #2a1a1a;
    background: #ffcb05;
    border-color: #5a0610;
  }
  .secciones img {
    width: 24px;
    flex: none;
  }
  .secciones span {
    flex: 1;
  }
  .secciones small {
    font-size: 11px;
    opacity: 0.7;
  }
  .lista {
    margin: 0;
    padding: 6px;
    list-style: none;
    background: #10231a;
  }
  .lista .grupo {
    margin: 8px 4px 4px;
    font-family: var(--titulo);
    font-size: 12px;
    letter-spacing: 1px;
    color: #7fd8a0;
  }
  .lista button {
    position: relative;
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    padding: 3px 6px 3px 18px;
    font-family: var(--titulo);
    font-size: 15px;
    text-align: left;
    color: #d8f5e0;
    background: none;
    border: 0;
    cursor: pointer;
  }
  .lista button:hover {
    background: rgba(127, 216, 160, 0.12);
  }
  .lista button.activa {
    color: #10231a;
    background: #7fd8a0;
  }
  .lista button.activa::before {
    content: '';
    position: absolute;
    left: 5px;
    border-top: 5px solid transparent;
    border-bottom: 5px solid transparent;
    border-left: 8px solid #10231a;
  }
  .n {
    font-size: 11px;
    opacity: 0.7;
  }
  .lista img {
    width: 30px;
    height: 30px;
    object-fit: contain;
  }
  .t {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .vacio {
    padding: 12px 8px;
    font-size: 14px;
    color: #9fd8b0;
  }
  .pantalla {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 16px 18px;
    color: var(--texto);
    background: var(--panel);
  }
  .entrada {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 18px;
    align-items: start;
    animation: pasar 0.25s steps(4) both;
  }
  @keyframes pasar {
    from {
      clip-path: inset(0 100% 0 0);
    }
  }
  .retrato {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }
  .escenario {
    position: relative;
    display: grid;
    place-items: end center;
    width: 190px;
    height: 170px;
    padding-bottom: 14px;
    background:
      repeating-linear-gradient(0deg, transparent 0 3px, rgba(0, 0, 0, 0.05) 3px 4px),
      radial-gradient(ellipse at 50% 30%, var(--panel-2), var(--fondo-2));
    border: 3px solid var(--borde);
    border-radius: 10px;
    overflow: hidden;
  }
  .escenario img {
    position: relative;
    z-index: 1;
    max-width: 170px;
    max-height: 140px;
    zoom: 1.6;
  }
  .quien {
    margin: 0;
    font-family: var(--titulo);
    font-size: 13px;
    color: var(--texto-suave);
  }
  .numero {
    margin: 0;
    font-family: var(--titulo);
    font-size: 14px;
    color: var(--texto-suave);
  }
  h3 {
    margin: 2px 0 8px;
    font-family: var(--titulo);
    font-size: 34px;
    line-height: 1.05;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-bottom: 10px;
  }
  .chip {
    padding: 1px 8px;
    font-family: var(--titulo);
    font-size: 12px;
    font-weight: 700;
    color: #fff;
    border: 2px solid rgba(0, 0, 0, 0.4);
    text-shadow: 1px 1px 0 rgba(0, 0, 0, 0.3);
  }
  .chip.gris {
    color: var(--texto);
    background: var(--panel-2);
    border-color: var(--borde-suave);
    text-shadow: none;
  }
  .texto {
    margin: 0;
    font-size: 17px;
    line-height: 1.55;
  }
  code {
    padding: 0 3px;
    font-family: var(--codigo);
    font-size: 0.92em;
    font-weight: 700;
    color: var(--palabra);
    background: var(--panel-2);
    border-radius: 3px;
  }
  .ejemplo .cabeza {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
    font-family: var(--titulo);
    font-size: 13px;
    letter-spacing: 2px;
    color: var(--texto-suave);
  }
  .copiar {
    padding: 2px 8px;
    font-family: var(--titulo);
    font-size: 12px;
    background: var(--panel-2);
    border: 2px solid var(--borde);
    cursor: pointer;
  }
  .notas {
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .notas li {
    display: flex;
    gap: 8px;
    align-items: flex-start;
    margin-bottom: 6px;
    font-size: 15px;
    line-height: 1.45;
  }
  .notas :global(img) {
    flex: none;
    margin-top: 2px;
  }
  .tabla {
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
    align-items: flex-start;
  }
  table {
    border-collapse: collapse;
    font-family: var(--titulo);
  }
  th,
  td {
    padding: 6px 10px;
    text-align: center;
    border: 2px solid var(--borde);
  }
  th {
    font-size: 13px;
    background: var(--panel-2);
  }
  td {
    font-size: 15px;
    font-weight: 700;
    cursor: default;
    transition: transform 0.08s steps(2);
  }
  td:hover {
    transform: scale(1.12);
    outline: 3px solid var(--borde);
  }
  .mt {
    color: #333;
    background: #d8d8d8;
  }
  .ef {
    color: #fff;
    background: #3fa34d;
  }
  .rc {
    color: #3a2a00;
    background: #f5c542;
  }
  .se {
    color: #fff;
    background: #d6453a;
  }
  .lectura {
    flex: 1;
    min-width: 220px;
    padding: 2px 6px;
  }
  .lectura b {
    display: block;
    margin-bottom: 4px;
    font-family: var(--titulo);
    font-size: 20px;
  }
  .lectura p {
    margin: 6px 0 0;
    font-size: 15px;
  }
  .escalera {
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .escalera li {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 0 0 4px calc(var(--n) * 14px);
    padding: 4px 10px;
    background: color-mix(in srgb, var(--acento) calc(8% + var(--n) * 6%), var(--panel));
    border: 2px solid var(--borde-suave);
    animation: pasar 0.3s steps(4) both;
    animation-delay: calc(var(--n) * 40ms);
  }
  .escalera code {
    min-width: 130px;
  }
  .escalera span {
    font-size: 14px;
    color: var(--texto-suave);
  }
  footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: auto;
    padding-top: 10px;
    font-family: var(--titulo);
    font-size: 14px;
    color: var(--texto-suave);
    border-top: 2px dashed var(--borde-suave);
  }
  @media (max-width: 1250px) {
    h2 small {
      display: none;
    }
    .cuerpo {
      grid-template-columns: 58px minmax(180px, 230px) minmax(0, 1fr);
    }
    .secciones span,
    .secciones small {
      display: none;
    }
    .secciones button {
      justify-content: center;
      padding: 8px 4px;
    }
    .secciones img {
      width: 28px;
    }
  }
  @media (max-width: 1000px) {
    .entrada {
      grid-template-columns: 1fr;
    }
    .escenario {
      width: 100%;
    }
    .buscar {
      max-width: none;
    }
  }
  .escenario :global(.sombra) {
    position: absolute;
    bottom: 14px;
  }
</style>

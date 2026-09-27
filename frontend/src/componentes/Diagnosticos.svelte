<script>
  // Pokédex de errores: una tarjeta por diagnóstico con su Pokémon invitado,
  // el pedazo de código donde está, qué pasó y por qué, el arreglo con su
  // vista previa y la lección del Profesor Oak. Las flechas pasan de uno a
  // otro, como las entradas de la Pokédex.
  import { ide } from '../lib/estado.svelte.js';
  import { sprite, objeto } from '../lib/pokemon.js';
  import { INVITADOS, aplicarArreglo, irA, oak } from '../lib/acciones.js';
  import { sonar } from '../lib/sonido.js';

  const CATEGORIAS = {
    lexico: 'LÉXICO',
    sintactico: 'SINTÁCTICO',
    semantico: 'SEMÁNTICO',
    importacion: 'IMPORTACIÓN',
    ejecucion: 'EJECUCIÓN',
  };

  let i = $state(0);
  let leccion = $state(false);

  const lista = $derived(ide.diagnosticos);
  const errores = $derived(lista.filter((d) => d.severity === 'error').length);
  const avisos = $derived(lista.length - errores);
  const d = $derived(lista[Math.min(i, lista.length - 1)]);

  // Cuando llega una compilación nueva, vuelve al primero.
  $effect(() => {
    lista;
    i = 0;
    leccion = false;
  });

  function mover(paso) {
    if (lista.length < 2) return;
    i = (i + paso + lista.length) % lista.length;
    leccion = false;
    sonar('mover');
  }

  // Líneas alrededor del error, con la parte señalada separada.
  const fragmento = $derived.by(() => {
    if (!d) return [];
    const lineas = (ide.contenidos[d.file] ?? '').split('\n');
    const salida = [];
    for (let n = Math.max(1, d.line - 1); n <= Math.min(lineas.length, d.line + 1); n++) {
      const runas = Array.from(lineas[n - 1] ?? '');
      if (n === d.line) {
        const desde = Math.max(0, d.col - 1);
        const hasta = desde + Math.max(1, d.len);
        salida.push({
          n,
          antes: runas.slice(0, desde).join(''),
          marca: runas.slice(desde, hasta).join('') || ' ',
          despues: runas.slice(hasta).join(''),
        });
      } else salida.push({ n, antes: runas.join('') });
    }
    return salida;
  });

  // La línea tal como queda después de aplicar el arreglo.
  const arreglada = $derived.by(() => {
    if (!d?.fix) return null;
    const runas = Array.from((ide.contenidos[d.file] ?? '').split('\n')[d.fix.line - 1] ?? '');
    const desde = d.fix.col - 1;
    return {
      antes: runas.join(''),
      despues:
        runas.slice(0, desde).join('') +
        d.fix.replacement +
        runas.slice(desde + d.fix.len).join(''),
    };
  });

  function explicar() {
    sonar('elegir');
    oak(
      `Lección: ${d.leccion.titulo}.`,
      d.leccion.texto,
      d.suggest ? `Para este caso: ${d.suggest}` : '¡Inténtalo de nuevo, tú puedes!',
    );
  }
</script>

{#if d}
  <section class="pokedex marco oscuro">
    <header>
      <span class="titulo">POKÉDEX DE ERRORES</span>
      <span class="cuentas">
        {#if errores}<img class="pixel" src={objeto('poke-ball')} alt="" />{errores}{/if}
        {#if avisos}<img class="pixel" src={objeto('great-ball')} alt="" />{avisos}{/if}
      </span>
      <span class="paginas">
        <button onclick={() => mover(-1)} disabled={lista.length < 2} aria-label="Anterior"
          ><i class="izq"></i></button
        >
        {i + 1}/{lista.length}
        <button onclick={() => mover(1)} disabled={lista.length < 2} aria-label="Siguiente"
          ><i class="der"></i></button
        >
      </span>
    </header>

    {#key `${i}-${d.line}-${d.col}-${d.code}`}
      <article class="tarjeta {d.severity}">
        <div class="cabeza">
          <div>
            <p class="numero">Nº {String(i + 1).padStart(3, '0')} · {d.code ?? d.category}</p>
            <h4>{d.heading}</h4>
            <p class="donde">
              <span class="chip">{CATEGORIAS[d.category] ?? d.category}</span>
              {d.file} · línea {d.line}, columna {d.col}
            </p>
          </div>
          {#if INVITADOS[d.heading]}
            <img class="sprite invitado" src={sprite(INVITADOS[d.heading])} alt="" />
          {/if}
        </div>

        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <pre
          class="codigo"
          onclick={() => irA(d)}
          title="Ir a la línea">{#each fragmento as f (f.n)}<span
              class="linea"
              class:actual={f.n === d.line}
              ><span class="num">{f.n}</span>{f.antes}{#if f.marca}<mark>{f.marca}</mark
                >{f.despues}{/if}</span
            >{/each}</pre>

        <dl>
          <dt>QUÉ PASÓ</dt>
          <dd>{d.desc}</dd>
          {#if d.cause}
            <dt>POR QUÉ</dt>
            <dd>{d.cause}</dd>
          {/if}
          {#if d.suggest && !d.fix}
            <dt>QUÉ HACER</dt>
            <dd>{d.suggest}</dd>
          {/if}
        </dl>

        {#if arreglada}
          <div class="arreglo">
            <p class="etiqueta">ARREGLO SUGERIDO</p>
            <pre class="diff"><span class="menos">- {arreglada.antes.trim()}</span><span class="mas"
                >+ {arreglada.despues.trim()}</span
              ></pre>
            <button class="boton principal" onclick={() => aplicarArreglo(d)}
              >USAR «{d.fix.replacement}»</button
            >
          </div>
        {/if}

        <div class="acciones">
          <button class="boton" onclick={() => irA(d)}>IR A LA LÍNEA</button>
          {#if d.leccion}
            <button class="boton" onclick={() => (leccion = !leccion)}>
              {leccion ? 'OCULTAR LECCIÓN' : 'VER LECCIÓN'}
            </button>
          {/if}
        </div>

        {#if leccion && d.leccion}
          <div class="leccion">
            <p class="etiqueta">LECCIÓN DEL PROF. OAK</p>
            <b>{d.leccion.titulo}</b>
            <p>{d.leccion.texto}</p>
            {#if d.leccion.ejemplo}<pre class="ejemplo">{d.leccion.ejemplo}</pre>{/if}
            <button class="boton" onclick={explicar}>QUE OAK LO EXPLIQUE</button>
          </div>
        {/if}
      </article>
    {/key}
  </section>
{/if}

<style>
  .pokedex {
    display: flex;
    flex-direction: column;
    padding: 0 4px 4px;
  }
  header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding-bottom: 6px;
    font-family: var(--titulo);
    font-size: 13px;
    letter-spacing: 1px;
    color: #ffcb05;
  }
  .titulo {
    flex: 1;
  }
  .cuentas {
    display: flex;
    align-items: center;
    gap: 3px;
    color: #f4f4ff;
  }
  .cuentas img {
    width: 18px;
  }
  .paginas {
    display: flex;
    align-items: center;
    gap: 4px;
    color: #f4f4ff;
  }
  .paginas button {
    display: grid;
    place-items: center;
    width: 20px;
    height: 20px;
    padding: 0;
    background: #3a3a52;
    border: 2px solid #101018;
    cursor: pointer;
  }
  .paginas button:disabled {
    opacity: 0.35;
    cursor: default;
  }
  .paginas i {
    border-top: 5px solid transparent;
    border-bottom: 5px solid transparent;
  }
  .izq {
    border-right: 7px solid #ffcb05;
  }
  .der {
    border-left: 7px solid #ffcb05;
  }
  .tarjeta {
    padding: 8px;
    color: var(--texto);
    background: var(--panel);
    border: 3px solid #101018;
    animation: pasar 0.25s steps(4) both;
  }
  @keyframes pasar {
    from {
      clip-path: inset(0 0 0 100%);
    }
  }
  .cabeza {
    display: flex;
    justify-content: space-between;
    gap: 6px;
    margin: -8px -8px 8px;
    padding: 6px 8px;
    color: #fff;
    background: var(--error);
    border-bottom: 3px solid #101018;
  }
  .warning .cabeza {
    color: #2a2008;
    background: var(--aviso);
  }
  .numero {
    margin: 0;
    font-family: var(--titulo);
    font-size: 11px;
    opacity: 0.85;
  }
  h4 {
    margin: 2px 0 4px;
    font-family: var(--titulo);
    font-size: 20px;
    line-height: 1.1;
    text-shadow: 2px 2px 0 rgba(0, 0, 0, 0.3);
  }
  .donde {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0;
    font-size: 12px;
  }
  .chip {
    padding: 0 4px;
    font-family: var(--titulo);
    font-size: 10px;
    color: #101018;
    background: #fff;
    border: 1px solid #101018;
  }
  .invitado {
    align-self: flex-end;
    width: 56px;
    height: 56px;
    object-fit: contain;
    margin: -10px -4px -6px 0;
    animation: asomar 0.4s steps(4) both;
  }
  @keyframes asomar {
    from {
      transform: translateY(20px);
      opacity: 0;
    }
  }
  .codigo {
    margin: 0 0 8px;
    padding: 6px 0;
    font-family: var(--codigo);
    font-size: 12.5px;
    line-height: 1.6;
    background: var(--editor-fondo);
    border: 2px solid var(--borde-suave);
    overflow-x: auto;
    cursor: pointer;
    user-select: text;
  }
  .linea {
    display: block;
    padding-right: 8px;
    white-space: pre;
    color: var(--texto-suave);
  }
  .linea.actual {
    color: var(--editor-texto);
    background: color-mix(in srgb, var(--error) 10%, transparent);
  }
  .num {
    display: inline-block;
    width: 30px;
    margin-right: 8px;
    padding-right: 6px;
    text-align: right;
    color: var(--texto-suave);
    border-right: 2px solid var(--borde-suave);
  }
  mark {
    color: inherit;
    background: color-mix(in srgb, var(--error) 22%, transparent);
    text-decoration: underline wavy var(--error);
    text-underline-offset: 3px;
  }
  .warning mark {
    background: color-mix(in srgb, var(--aviso) 22%, transparent);
    text-decoration-color: var(--aviso);
  }
  dl {
    margin: 0 0 8px;
    font-size: 13px;
    line-height: 1.4;
  }
  dt {
    margin-top: 4px;
    font-family: var(--titulo);
    font-size: 11px;
    letter-spacing: 1px;
    color: var(--texto-suave);
  }
  dd {
    margin: 0;
  }
  .etiqueta {
    margin: 0 0 4px;
    font-family: var(--titulo);
    font-size: 11px;
    letter-spacing: 1px;
    color: var(--texto-suave);
  }
  .arreglo {
    margin-bottom: 8px;
  }
  .diff {
    display: flex;
    flex-direction: column;
    margin: 0 0 6px;
    font-family: var(--codigo);
    font-size: 12.5px;
    border: 2px solid var(--borde-suave);
    overflow-x: auto;
    user-select: text;
  }
  .menos,
  .mas {
    padding: 1px 6px;
    white-space: pre;
  }
  .menos {
    color: var(--error);
    background: color-mix(in srgb, var(--error) 12%, transparent);
    text-decoration: line-through;
  }
  .mas {
    color: var(--exito);
    background: color-mix(in srgb, var(--exito) 14%, transparent);
  }
  .acciones {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .acciones .boton,
  .arreglo .boton,
  .leccion .boton {
    min-height: 32px;
    font-size: 13px;
  }
  .leccion {
    margin-top: 8px;
    padding: 6px 8px;
    font-size: 13px;
    background: var(--panel-2);
    border: 2px dashed var(--borde-suave);
    animation: pasar 0.2s steps(3) both;
  }
  .leccion p {
    margin: 4px 0 8px;
    line-height: 1.4;
  }
  .ejemplo {
    margin: 0 0 8px;
    padding: 6px;
    font-family: var(--codigo);
    font-size: 12px;
    background: var(--editor-fondo);
    border: 2px solid var(--borde-suave);
    overflow-x: auto;
  }
</style>

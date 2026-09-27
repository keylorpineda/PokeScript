<script>
  // Bloque de código de solo lectura, coloreado igual que en el editor.
  import { colorear } from '../lib/pokescript.js';

  let { codigo = '' } = $props();
  const lineas = $derived(colorear(codigo));
</script>

<pre class="codigo">{#each lineas as linea, i (i)}<span class="linea"
      ><span class="num">{i + 1}</span>{#each linea as t, j (j)}<span class={t.clase}
          >{t.texto}</span
        >{/each}</span
    >{/each}</pre>

<style>
  .codigo {
    margin: 0;
    padding: 8px 0;
    font-family: var(--codigo);
    font-size: 14px;
    line-height: 1.6;
    color: var(--editor-texto);
    background: var(--editor-fondo);
    border: 3px solid var(--borde);
    box-shadow: inset 0 0 0 2px var(--borde-suave);
    overflow-x: auto;
    user-select: text;
  }
  .linea {
    display: block;
    padding-right: 12px;
    white-space: pre;
  }
  .num {
    display: inline-block;
    width: 34px;
    margin-right: 10px;
    padding-right: 8px;
    text-align: right;
    color: var(--texto-suave);
    border-right: 2px solid var(--borde-suave);
    user-select: none;
  }
  .t-roca {
    color: var(--t-roca);
    font-weight: 700;
  }
  .t-agua {
    color: var(--t-agua);
    font-weight: 700;
  }
  .t-fuego {
    color: var(--t-fuego);
    font-weight: 700;
  }
  .t-planta {
    color: var(--t-planta);
    font-weight: 700;
  }
  .t-electrico {
    color: var(--t-electrico);
    font-weight: 700;
  }
  .coleccion,
  .tipo-nombrado {
    color: var(--nombre_tipo);
    font-weight: 700;
  }
  .estructura {
    color: var(--palabra);
    font-weight: 700;
  }
  .control {
    color: var(--control);
    font-weight: 700;
  }
  .funcion,
  .constante {
    color: var(--constante);
  }
  .funcion {
    font-weight: 700;
  }
  .numero {
    color: var(--numero);
  }
  .texto {
    color: var(--texto_lit);
  }
  .comentario {
    color: var(--comentario);
    font-style: italic;
  }
  .suave {
    color: var(--texto-suave);
  }
</style>

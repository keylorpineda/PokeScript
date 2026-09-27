<script>
  // Cuadro de diálogo de los juegos: el texto se escribe letra por letra, un
  // clic (o Enter) completa la línea y el siguiente clic avanza. Lo que va
  // entre comillas invertidas se pinta como código.
  import { onDestroy, untrack } from 'svelte';
  import { sonar } from '../lib/sonido.js';

  let {
    texto = '',
    nombre = '',
    velocidad = 24,
    alAvanzar = null,
    alTerminar = null,
    teclado = true,
    compacto = false,
    silencioso = false, // sin el pitido de cada letra
    children,
  } = $props();

  // La tecla que abrió esta pantalla no debe avanzar el diálogo.
  const creado = Date.now();
  let visibles = $state(0);
  let terminado = false;
  const letras = $derived([...texto.replace(/`/g, '')]);
  const listo = $derived(visibles >= letras.length);

  // El texto se vuelve a escribir solo si cambia de verdad: si el padre se
  // redibuja (por ejemplo, mientras escribes en un campo del diálogo) con el
  // mismo texto, sigue como estaba.
  let previo = null;
  let reloj = null;
  $effect(() => {
    const t = texto;
    if (t === previo) return;
    previo = t;
    const total = [...t.replace(/`/g, '')].length;
    clearInterval(reloj);
    untrack(() => {
      visibles = 0;
      terminado = false;
    });
    reloj = setInterval(() => {
      visibles += 1;
      if (!silencioso && visibles % 3 === 0) sonar('letra');
      if (visibles >= total) {
        clearInterval(reloj);
        fin();
      }
    }, velocidad);
  });
  onDestroy(() => clearInterval(reloj));

  function fin() {
    if (terminado) return;
    terminado = true;
    alTerminar?.();
  }

  function avanzar() {
    if (!listo) {
      visibles = letras.length;
      fin();
    } else if (alAvanzar) {
      sonar('mover');
      alAvanzar();
    }
  }

  function tecla(e) {
    if (!teclado || e.repeat || Date.now() - creado < 300) return;
    if (e.target instanceof HTMLInputElement) return;
    if (e.key === 'Enter' || e.key === ' ' || e.key === 'z') {
      e.preventDefault();
      avanzar();
    }
  }

  // Parte el texto en trozos normales y de código, recortado a lo visible.
  const trozos = $derived.by(() => {
    const salida = [];
    let quedan = visibles;
    texto.split('`').forEach((t, i) => {
      const l = [...t];
      salida.push({ codigo: i % 2 === 1, texto: l.slice(0, Math.max(0, quedan)).join('') });
      quedan -= l.length;
    });
    return salida;
  });
</script>

<svelte:window onkeydown={tecla} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="dialogo marco" class:compacto onclick={avanzar}>
  {#if nombre}<div class="nombre">{nombre}</div>{/if}
  <p class="texto">
    <span class="fantasma">{texto.replace(/`/g, '')}</span>
    <span class="visible"
      >{#each trozos as t, i (i)}{#if t.codigo}<code>{t.texto}</code
          >{:else}{t.texto}{/if}{/each}</span
    >
  </p>
  {#if listo && children}
    <div class="extra">{@render children()}</div>
  {/if}
  {#if listo && alAvanzar}<span class="flecha" aria-hidden="true"></span>{/if}
</div>

<style>
  .dialogo {
    position: relative;
    padding: 6px 14px 8px;
    cursor: pointer;
    min-height: 110px;
  }
  .compacto {
    padding: 2px 6px 4px;
    min-height: 0;
  }
  .nombre {
    position: absolute;
    top: -30px;
    left: 4px;
    padding: 1px 12px 2px;
    font-family: var(--titulo);
    font-weight: 700;
    font-size: 16px;
    letter-spacing: 1px;
    color: var(--acento-texto);
    background: var(--acento);
    border: 3px solid var(--borde);
    box-shadow: inset 0 -3px 0 rgba(0, 0, 0, 0.2);
  }
  .texto {
    position: relative;
    margin: 0;
    font-family: var(--titulo);
    font-size: 24px;
    font-weight: 500;
    line-height: 1.5;
    white-space: pre-wrap;
    color: var(--texto);
    text-shadow: 2px 2px 0 var(--borde-suave);
  }
  .compacto .texto {
    font-size: 17px;
    text-shadow: 1px 1px 0 var(--borde-suave);
  }
  .fantasma {
    visibility: hidden;
  }
  .visible {
    position: absolute;
    inset: 0;
  }
  code {
    font-family: var(--titulo);
    font-weight: 700;
    color: var(--palabra);
  }
  .extra {
    margin-top: 12px;
  }
  .flecha {
    position: absolute;
    right: 6px;
    bottom: 2px;
    width: 0;
    height: 0;
    border-left: 9px solid transparent;
    border-right: 9px solid transparent;
    border-top: 11px solid var(--acento);
    animation: saltito 0.8s steps(2) infinite;
  }
  @keyframes saltito {
    50% {
      transform: translateY(4px);
    }
  }
</style>

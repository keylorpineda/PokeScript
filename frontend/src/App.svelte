<script>
  // Recorrido de pantallas: título (con su menú) → presentación de Oak (la
  // primera vez, o cuando se pide) → carga del proyecto → IDE.
  import Titulo from './componentes/Titulo.svelte';
  import IntroOak from './componentes/IntroOak.svelte';
  import Carga from './componentes/Carga.svelte';
  import Principal from './componentes/Principal.svelte';
  import Explorador from './componentes/Explorador.svelte';
  import BarraVentana from './componentes/BarraVentana.svelte';
  import { navegacion, explorador } from './lib/estado.svelte.js';
  import { alternarPantallaCompleta } from './lib/ventana.js';

  // Dentro de Wails la ventana no tiene marco: la barra propia la reemplaza.
  const enVentana = Boolean(window.runtime?.WindowMinimise);

  // Lo que eligió el menú antes de pasar por Oak.
  let pendiente = null;

  function desdeTitulo(eleccion) {
    if (eleccion.accion === 'oak') {
      pendiente = eleccion.ruta ?? null;
      navegacion.pantalla = 'oak';
      return;
    }
    navegacion.ruta = eleccion.ruta;
    navegacion.pantalla = 'carga';
  }

  function trasOak() {
    if (pendiente) {
      navegacion.ruta = pendiente;
      navegacion.pantalla = 'carga';
    } else {
      // Recién llega el entrenador: vuelve al menú para crear o abrir un proyecto.
      navegacion.pantalla = 'menu';
    }
  }
</script>

<svelte:window
  onkeydown={(e) => {
    if (e.key === 'F11') {
      e.preventDefault();
      alternarPantallaCompleta();
    }
  }}
/>

{#if enVentana}<BarraVentana />{/if}

<!-- El contenedor transformado hace que las pantallas «fixed» ocupen solo el
     espacio debajo de la barra de la ventana. -->
<div class="pantallas" style="--alto-barra:{enVentana ? '32px' : '0px'}">
  {#if navegacion.pantalla === 'titulo' || navegacion.pantalla === 'menu'}
    {#key navegacion.pantalla}
      <Titulo alContinuar={desdeTitulo} conMenu={navegacion.pantalla === 'menu'} />
    {/key}
  {:else if navegacion.pantalla === 'oak'}
    <IntroOak alTerminar={trasOak} />
  {:else if navegacion.pantalla === 'carga'}
    <Carga
      proyecto={navegacion.ruta.split(/[\\/]/).pop()}
      alTerminar={() => (navegacion.pantalla = 'ide')}
    />
  {:else}
    {#key navegacion.ruta}
      <Principal ruta={navegacion.ruta} />
    {/key}
  {/if}

  {#if explorador.abierto}<Explorador />{/if}
</div>

<style>
  :global(:root) {
    --alto-barra: 32px;
  }
  .pantallas {
    position: fixed;
    top: var(--alto-barra);
    left: 0;
    right: 0;
    bottom: 0;
    overflow: hidden;
    transform: translateZ(0);
  }
</style>

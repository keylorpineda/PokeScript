<script>
  // Recorrido de pantallas: título (con su menú) → presentación de Oak (la
  // primera vez, o cuando se pide) → carga del proyecto → IDE.
  import Titulo from './componentes/Titulo.svelte';
  import IntroOak from './componentes/IntroOak.svelte';
  import Carga from './componentes/Carga.svelte';
  import Principal from './componentes/Principal.svelte';
  import Explorador from './componentes/Explorador.svelte';
  import { navegacion, explorador } from './lib/estado.svelte.js';

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

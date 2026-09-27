<script>
  // Recorrido de pantallas: título → presentación de Oak (solo la primera
  // vez) → carga del proyecto → IDE.
  import Titulo from './componentes/Titulo.svelte';
  import IntroOak from './componentes/IntroOak.svelte';
  import Carga from './componentes/Carga.svelte';
  import Principal from './componentes/Principal.svelte';
  import { perfil } from './lib/estado.svelte.js';

  let pantalla = $state('titulo');

  function trasTitulo() {
    pantalla = perfil.companero ? 'carga' : 'oak';
  }
</script>

{#if pantalla === 'titulo'}
  <Titulo alContinuar={trasTitulo} />
{:else if pantalla === 'oak'}
  <IntroOak alTerminar={() => (pantalla = 'carga')} />
{:else if pantalla === 'carga'}
  <Carga proyecto="centro_pokemon" alTerminar={() => (pantalla = 'ide')} />
{:else}
  <Principal />
{/if}

<script>
  // Barra de arriba: logo, acciones del combate, selector de tema y la
  // tarjeta del entrenador.
  import Pokebola from './Pokebola.svelte';
  import { ide, perfil, cambiarTema, cambiarSonido } from '../lib/estado.svelte.js';
  import { TEMAS } from '../lib/temas.js';
  import { sprite } from '../lib/pokemon.js';
  import { compilar, ejecutar, detener, salirAlMenu } from '../lib/acciones.js';
  import { sonar } from '../lib/sonido.js';

  let menuTemas = $state(false);

  function elegirTema(id) {
    sonar('tema');
    cambiarTema(id);
    menuTemas = false;
  }
</script>

<svelte:window onclick={(e) => menuTemas && !e.target.closest('.temas') && (menuTemas = false)} />

<header class="barra">
  <div class="logo">
    <Pokebola escala={1.3} />
    <span>PokeScript</span>
  </div>

  <div class="acciones">
    <button class="boton" onclick={compilar} disabled={ide.compilando} title="Ctrl+Enter">
      <img class="pixel icono" src="/objetos/town-map.png" alt="" />ANALIZAR
    </button>
    <button class="boton principal" onclick={ejecutar} title="F5">
      <Pokebola escala={0.9} />¡COMBATE!
    </button>
    <button class="boton" onclick={detener} disabled={!ide.ejecutando}>HUIR</button>
  </div>

  <div class="derecha">
    <button class="boton" onclick={salirAlMenu} title="Volver a la portada">MENÚ</button>
    <div class="temas">
      <button class="boton" onclick={() => (menuTemas = !menuTemas)}>
        <img class="sprite mini" src={sprite(TEMAS[perfil.tema].escena.mascota)} alt="" />
        {TEMAS[perfil.tema].nombre.toUpperCase()}
      </button>
      {#if menuTemas}
        <div class="menu marco">
          {#each Object.entries(TEMAS) as [id, t] (id)}
            <button class="opcion" class:sel={id === perfil.tema} onclick={() => elegirTema(id)}>
              <span
                class="muestra"
                style="background-image:url(/fondos/{t.escena.fondo}.png);filter:{t.escena.filtro ??
                  t.filtroSprites ??
                  'none'}"
              >
                <img class="sprite" src={sprite(t.escena.mascota)} alt="" />
              </span>
              {t.nombre}
            </button>
          {/each}
        </div>
      {/if}
    </div>
    <button class="boton" onclick={() => cambiarSonido(!perfil.sonido)} title="Sonido">
      {perfil.sonido ? 'SONIDO: SÍ' : 'SONIDO: NO'}
    </button>
    <div class="entrenador marco">
      <img class="sprite mini" src={sprite(perfil.companero ?? 'pikachu')} alt="" />
      <span>{perfil.nombre}</span>
    </div>
  </div>
</header>

<style>
  .barra {
    position: relative;
    z-index: 10;
    display: flex;
    align-items: center;
    gap: 18px;
    padding: 8px 12px;
    background: var(--fondo-2);
    border-bottom: 4px solid var(--borde);
  }
  .logo {
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: var(--titulo);
    font-size: 28px;
    font-weight: 700;
    color: #ffcb05;
    text-shadow:
      2px 0 0 #2a4fa0,
      -2px 0 0 #2a4fa0,
      0 2px 0 #2a4fa0,
      0 -2px 0 #2a4fa0,
      2px 2px 0 #2a4fa0,
      -2px 2px 0 #2a4fa0,
      2px -2px 0 #2a4fa0,
      -2px -2px 0 #2a4fa0,
      0 5px 0 #14265a;
  }
  .acciones,
  .derecha {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .acciones {
    margin: 0 auto;
  }
  .icono {
    width: 26px;
  }
  .mini {
    width: 30px;
    height: 30px;
    object-fit: contain;
  }
  .temas {
    position: relative;
  }
  .menu {
    position: absolute;
    right: 0;
    top: calc(100% + 8px);
    z-index: 20;
    display: grid;
    gap: 4px;
    width: 250px;
    padding: 2px;
  }
  .opcion {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 4px;
    font-family: var(--titulo);
    font-size: 17px;
    text-align: left;
    background: none;
    border: 3px solid transparent;
    cursor: pointer;
  }
  .opcion:hover,
  .opcion.sel {
    border-color: var(--borde);
    background: var(--panel-2);
  }
  .muestra {
    display: grid;
    place-items: center;
    width: 64px;
    height: 44px;
    background-size: cover;
    background-position: center;
    image-rendering: pixelated;
    border: 2px solid var(--borde);
    overflow: hidden;
  }
  .muestra img {
    width: 40px;
    height: 40px;
    object-fit: contain;
    filter: none;
  }
  .entrenador {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 6px;
    font-family: var(--titulo);
    font-size: 17px;
    font-weight: 700;
  }
</style>

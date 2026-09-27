<script>
  // Barra de arriba, en tres partes: el logo, las acciones del combate al
  // centro, y a la derecha la Pokédex y la tarjeta del entrenador (que abre
  // tema, sonidos y la vuelta a la portada).
  import Pokebola from './Pokebola.svelte';
  import { ide, perfil, cambiarTema, cambiarSonido, logro } from '../lib/estado.svelte.js';
  import { TEMAS } from '../lib/temas.js';
  import { POKEMON, sprite, objeto } from '../lib/pokemon.js';
  import { compilar, ejecutar, detener, salirAlMenu, oak } from '../lib/acciones.js';
  import { sonar } from '../lib/sonido.js';
  import { alternarPantallaCompleta } from '../lib/ventana.js';

  let abierto = $state(false);
  const nivel = $derived(5 + Math.floor((perfil.exp ?? 0) / 3));

  function alternar() {
    abierto = !abierto;
    sonar(abierto ? 'elegir' : 'mover');
  }

  function elegirTema(id) {
    sonar('tema');
    cambiarTema(id);
  }

  function abrirPokedex() {
    sonar('abrir');
    ide.pokedex = true;
    if (logro('primera-pokedex')) {
      oak(
        '¡Ah, encontraste la Pokédex! Aquí está todo lo que sé sobre PokeScript.',
        'Cada tema lo explica un Pokémon distinto. Usa las flechas para moverte y el buscador para encontrar cualquier palabra.',
      );
    }
  }
</script>

<svelte:window
  onclick={(e) => abierto && !e.target.closest('.entrenador') && (abierto = false)}
  onkeydown={(e) => e.key === 'F1' && (e.preventDefault(), abrirPokedex())}
/>

<header class="barra">
  <div class="logo">
    <Pokebola escala={1.2} />
    <span class="nombre-logo">PokeScript</span>
  </div>

  <nav class="acciones">
    <button class="boton" onclick={compilar} disabled={ide.compilando} title="Ctrl+Enter">
      <img class="pixel icono" src={objeto('town-map')} alt="" /><span>ANALIZAR</span>
    </button>
    <button class="boton principal" onclick={ejecutar} title="F5">
      <Pokebola escala={0.85} /><span>¡COMBATE!</span>
    </button>
    <button class="boton" onclick={detener} disabled={!ide.ejecutando} title="Detener">
      <span class="siempre">HUIR</span>
    </button>
  </nav>

  <div class="derecha">
    <button class="boton" onclick={abrirPokedex} title="Todo sobre el lenguaje">
      <img class="pixel icono" src={objeto('exp-share')} alt="" /><span>POKÉDEX</span>
    </button>

    <div class="entrenador">
      <button class="boton tarjeta" class:abierto onclick={alternar} title="Opciones">
        <img class="sprite mini" src={sprite(perfil.companero ?? 'pikachu')} alt="" />
        <span class="quien">{perfil.nombre}</span>
        <i class="flechita"></i>
      </button>

      {#if abierto}
        <div class="menu marco">
          <div class="ficha">
            <img class="sprite" src={sprite(perfil.companero ?? 'pikachu')} alt="" />
            <div>
              <b>{perfil.nombre}</b>
              <small>{POKEMON[perfil.companero]?.nombre} · Nv{nivel}</small>
            </div>
          </div>

          <p class="seccion">TEMA</p>
          <div class="temas">
            {#each Object.entries(TEMAS) as [id, t] (id)}
              <button class="tema" class:sel={id === perfil.tema} onclick={() => elegirTema(id)}>
                <span
                  class="muestra"
                  style="background-image:url(/fondos/{t.escena.fondo}.png);filter:{t.escena
                    .filtro ??
                    t.filtroSprites ??
                    'none'}"
                >
                  <img class="sprite" src={sprite(t.escena.mascota)} alt="" />
                </span>
                <span class="nombre-tema">{t.nombre}</span>
              </button>
            {/each}
          </div>

          <button class="fila" onclick={() => cambiarSonido(!perfil.sonido)}>
            <span class="luz" class:on={perfil.sonido}></span>
            Efectos de sonido: {perfil.sonido ? 'sí' : 'no'}
          </button>
          <button class="fila" onclick={alternarPantallaCompleta}>
            <span class="marco-pantalla" aria-hidden="true"></span>Pantalla completa (F11)
          </button>
          <button class="fila" onclick={salirAlMenu}>
            <img class="pixel" src={objeto('poke-ball')} alt="" />Volver a la portada
          </button>
        </div>
      {/if}
    </div>
  </div>
</header>

<style>
  .barra {
    position: relative;
    z-index: 10;
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    gap: 14px;
    padding: 8px 12px;
    background: var(--fondo-2);
    border-bottom: 4px solid var(--borde);
  }
  .logo {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    font-family: 'Pixelify Sans', sans-serif;
    font-size: 28px;
    font-weight: 700;
    white-space: nowrap;
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
    gap: 10px;
  }
  .derecha {
    justify-content: flex-end;
  }
  .boton {
    white-space: nowrap;
  }
  .icono {
    width: 24px;
  }
  .mini {
    width: 30px;
    height: 30px;
    object-fit: contain;
  }
  .entrenador {
    position: relative;
  }
  .tarjeta .quien {
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .flechita {
    border-left: 5px solid transparent;
    border-right: 5px solid transparent;
    border-top: 7px solid var(--texto);
    transition: transform 0.1s steps(2);
  }
  .tarjeta.abierto .flechita {
    transform: rotate(180deg);
  }
  .menu {
    position: absolute;
    right: 0;
    top: calc(100% + 8px);
    z-index: 20;
    width: 330px;
    padding: 2px 4px 4px;
    animation: bajar 0.18s steps(3) both;
  }
  @keyframes bajar {
    from {
      clip-path: inset(0 0 100% 0);
    }
  }
  .ficha {
    display: flex;
    align-items: center;
    gap: 10px;
    padding-bottom: 8px;
    border-bottom: 2px dashed var(--borde-suave);
  }
  .ficha img {
    width: 56px;
    height: 56px;
    object-fit: contain;
  }
  .ficha b {
    display: block;
    font-family: var(--titulo);
    font-size: 20px;
  }
  .ficha small {
    color: var(--texto-suave);
  }
  .seccion {
    margin: 8px 0 6px;
    font-family: var(--titulo);
    font-size: 13px;
    letter-spacing: 2px;
    color: var(--texto-suave);
  }
  .temas {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 6px;
    margin-bottom: 8px;
  }
  .tema {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
    padding: 3px;
    background: none;
    border: 3px solid transparent;
    cursor: pointer;
  }
  .tema:hover {
    border-color: var(--borde-suave);
  }
  .tema.sel {
    border-color: var(--borde);
    background: var(--panel-2);
  }
  .muestra {
    display: grid;
    place-items: center;
    width: 100%;
    aspect-ratio: 4 / 3;
    background-size: cover;
    background-position: center;
    image-rendering: pixelated;
    border: 2px solid var(--borde);
    overflow: hidden;
  }
  .muestra img {
    width: 34px;
    height: 34px;
    object-fit: contain;
    filter: none;
  }
  .nombre-tema {
    font-size: 11px;
    line-height: 1.1;
    text-align: center;
  }
  .fila {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 6px 4px;
    font-family: var(--cuerpo);
    font-size: 14px;
    font-weight: 600;
    text-align: left;
    background: none;
    border: 0;
    border-top: 2px dashed var(--borde-suave);
    cursor: pointer;
  }
  .fila:hover {
    background: var(--panel-2);
  }
  .fila img {
    width: 22px;
  }
  .luz {
    width: 12px;
    height: 12px;
    background: var(--error);
    border: 2px solid var(--borde);
  }
  .luz.on {
    background: var(--exito);
    box-shadow: 0 0 6px var(--exito);
  }

  /* En ventanas angostas se esconden los textos que sobran. */
  @media (max-width: 1250px) {
    .nombre-logo {
      display: none;
    }
  }
  @media (max-width: 1100px) {
    .derecha .boton span:not(.quien),
    .tarjeta .quien {
      display: none;
    }
  }
  .marco-pantalla {
    width: 18px;
    height: 13px;
    border: 3px solid var(--texto);
    border-radius: 2px;
  }
  @media (max-width: 820px) {
    .barra {
      gap: 8px;
      padding: 6px 8px;
    }
    .acciones .boton span:not(.siempre) {
      display: none;
    }
    .logo :global(img) {
      display: none;
    }
  }
</style>

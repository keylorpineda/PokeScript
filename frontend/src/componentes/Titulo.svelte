<script>
  // Pantalla de título: el lugar del tema en pixel art, con su Pokémon
  // animado, el logo y el clásico «PRESIONA CUALQUIER TECLA».
  import Escena from './Escena.svelte';
  import { perfil } from '../lib/estado.svelte.js';
  import { temaActual } from '../lib/temas.js';
  import { sprite } from '../lib/pokemon.js';
  import { sonar } from '../lib/sonido.js';

  let { alContinuar } = $props();

  const escena = $derived(temaActual(perfil.tema).escena);
  let listo = false;
  setTimeout(() => (listo = true), 700);

  function continuar(e) {
    if (!listo || e?.repeat) return;
    sonar('elegir');
    alContinuar();
  }
</script>

<svelte:window onkeydown={continuar} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="titulo" onclick={continuar}>
  <Escena oscurecer={0.12}>
    <div class="mascota" class:vuela={escena.vuela}>
      <img class="sprite" src={sprite(escena.mascota)} alt="" draggable="false" />
      {#if !escena.vuela}<span class="sombra"></span>{/if}
    </div>

    <div class="logo">
      <h1>PokeScript</h1>
      <div class="cinta">EDICIÓN PARADIGMAS · UNA 2026</div>
    </div>

    <div class="presiona marco oscuro">PRESIONA CUALQUIER TECLA</div>
  </Escena>
  <div class="telon"></div>
</div>

<style>
  .titulo {
    position: fixed;
    inset: 0;
    cursor: pointer;
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
    animation: llegar 0.9s 0.9s steps(10) both;
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
    animation: parpadeo 1.2s steps(1) infinite;
  }
</style>

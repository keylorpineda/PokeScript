<script>
  // Barra de la ventana con la temática del IDE: se arrastra para mover la
  // ventana y trae minimizar, maximizar, pantalla completa y cerrar. Solo se
  // muestra dentro de Wails (la ventana no tiene el marco de Windows).
  import { onMount } from 'svelte';
  import Pokebola from './Pokebola.svelte';
  import { sonar } from '../lib/sonido.js';

  const rt = window.runtime;
  let maximizada = $state(false);
  let completa = $state(false);

  async function actualizar() {
    maximizada = await rt.WindowIsMaximised();
    completa = await rt.WindowIsFullscreen();
  }

  onMount(() => {
    actualizar();
    window.addEventListener('resize', actualizar);
    return () => window.removeEventListener('resize', actualizar);
  });

  function minimizar() {
    sonar('mover');
    rt.WindowMinimise();
  }
  function maximizar() {
    sonar('mover');
    if (completa) rt.WindowUnfullscreen();
    else rt.WindowToggleMaximise();
  }
  function pantallaCompleta() {
    sonar('mover');
    if (completa) rt.WindowUnfullscreen();
    else rt.WindowFullscreen();
  }
  function cerrar() {
    rt.Quit();
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<header class="ventana" ondblclick={maximizar}>
  <div class="titulo">
    <Pokebola escala={0.6} />
    <span>PokeScript</span>
  </div>

  <div class="controles" ondblclick={(e) => e.stopPropagation()}>
    <button onclick={minimizar} title="Minimizar" aria-label="Minimizar">
      <i class="min"></i>
    </button>
    <button
      onclick={pantallaCompleta}
      title={completa ? 'Salir de pantalla completa (F11)' : 'Pantalla completa (F11)'}
      aria-label="Pantalla completa"
    >
      <i class="completa" class:salir={completa}></i>
    </button>
    <button
      onclick={maximizar}
      title={maximizada || completa ? 'Modo ventana' : 'Maximizar'}
      aria-label="Maximizar o restaurar"
    >
      <i class={maximizada || completa ? 'restaurar' : 'max'}></i>
    </button>
    <button class="cerrar" onclick={cerrar} title="Cerrar" aria-label="Cerrar">
      <i class="x"></i>
    </button>
  </div>
</header>

<style>
  .ventana {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 100;
    display: flex;
    align-items: center;
    height: var(--alto-barra);
    padding-left: 10px;
    background: var(--borde);
    color: var(--panel);
    --wails-draggable: drag;
    user-select: none;
  }
  .titulo {
    display: flex;
    align-items: center;
    gap: 6px;
    font-family: 'Pixelify Sans', sans-serif;
    font-size: 15px;
    font-weight: 700;
    letter-spacing: 1px;
    color: #ffcb05;
  }
  .controles {
    display: flex;
    height: 100%;
    margin-left: auto;
    --wails-draggable: no-drag;
  }
  button {
    display: grid;
    place-items: center;
    width: 46px;
    height: 100%;
    padding: 0;
    background: none;
    border: 0;
    cursor: pointer;
  }
  button:hover {
    background: color-mix(in srgb, var(--panel) 18%, transparent);
  }
  .cerrar:hover {
    background: #d6302f;
  }
  i {
    position: relative;
    display: block;
    color: var(--panel);
  }
  /* Íconos dibujados con cuadritos, sin emojis. */
  .min {
    width: 12px;
    height: 3px;
    margin-top: 8px;
    background: currentColor;
  }
  .max {
    width: 12px;
    height: 11px;
    border: 2px solid currentColor;
    border-top-width: 4px;
  }
  .restaurar {
    width: 10px;
    height: 9px;
    border: 2px solid currentColor;
    border-top-width: 3px;
    transform: translate(-2px, 2px);
    box-shadow:
      4px -4px 0 -2px var(--borde),
      4px -4px 0 0 currentColor;
  }
  .completa,
  .completa.salir {
    width: 14px;
    height: 12px;
    background:
      linear-gradient(currentColor, currentColor) 0 0 / 5px 2px,
      linear-gradient(currentColor, currentColor) 0 0 / 2px 5px,
      linear-gradient(currentColor, currentColor) 100% 0 / 5px 2px,
      linear-gradient(currentColor, currentColor) 100% 0 / 2px 5px,
      linear-gradient(currentColor, currentColor) 0 100% / 5px 2px,
      linear-gradient(currentColor, currentColor) 0 100% / 2px 5px,
      linear-gradient(currentColor, currentColor) 100% 100% / 5px 2px,
      linear-gradient(currentColor, currentColor) 100% 100% / 2px 5px;
    background-repeat: no-repeat;
  }
  .completa.salir {
    transform: scale(0.75);
    opacity: 0.8;
  }
  .x {
    width: 14px;
    height: 14px;
  }
  .x::before,
  .x::after {
    content: '';
    position: absolute;
    left: 6px;
    top: -1px;
    width: 3px;
    height: 16px;
    background: currentColor;
    transform: rotate(45deg);
  }
  .x::after {
    transform: rotate(-45deg);
  }
</style>

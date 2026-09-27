<script>
  // Fondo de pixel art de los juegos con vida encima: partículas según el
  // lugar (brasas en el volcán, hojas en la ruta, nieve en la cueva…), un
  // paneo lento y una viñeta. Lo demás se pinta encima con children.
  import { temaActual } from '../lib/temas.js';
  import { perfil } from '../lib/estado.svelte.js';

  // «lugar» fuerza la escena de un tema (Oak siempre está en Pueblo Paleta) y
  // «portada» da una escena completa (las portadas del título).
  let { oscurecer = 0, paneo = true, lugar = null, portada = null, children } = $props();

  const escena = $derived(portada ?? temaActual(lugar ?? perfil.tema).escena);

  // Posiciones fijas (sin aleatorio) para que no salten al redibujar.
  const PARTICULAS = Array.from({ length: 34 }, (_, i) => ({
    x: (i * 37.3) % 100,
    y: (i * 61.7) % 100,
    retraso: -((i * 1.37) % 9),
    dur: 6 + ((i * 7) % 6),
    tam: 1 + (i % 3),
  }));
</script>

<div class="escena">
  <div
    class="fondo"
    class:paneo
    style="background-image:url(/fondos/{escena.fondo}.png);filter:{escena.filtro ??
      'var(--filtro-escena)'}"
  ></div>

  <div class="particulas {escena.particulas}">
    {#each PARTICULAS as p, i (i)}
      <span
        style="left:{p.x}%;top:{p.y}%;--t:{p.tam *
          3}px;animation-delay:{p.retraso}s;animation-duration:{p.dur}s"
      ></span>
    {/each}
  </div>

  <div class="velo" style="background:rgba(8,6,16,{oscurecer})"></div>
  <div class="vineta"></div>
  <div class="contenido">{@render children?.()}</div>
</div>

<style>
  .escena {
    position: fixed;
    inset: 0;
    overflow: hidden;
    background: #000;
  }
  .fondo {
    position: absolute;
    inset: -4%;
    background-size: cover;
    background-position: center bottom;
    image-rendering: pixelated;
  }
  .paneo {
    animation: paneo 40s ease-in-out infinite alternate;
  }
  @keyframes paneo {
    from {
      transform: translateX(-2%) scale(1.02);
    }
    to {
      transform: translateX(2%) scale(1.06);
    }
  }
  .velo,
  .vineta,
  .particulas,
  .contenido {
    position: absolute;
    inset: 0;
  }
  .vineta {
    background: radial-gradient(ellipse at center, transparent 55%, rgba(0, 0, 0, 0.45) 100%);
    pointer-events: none;
  }
  .particulas {
    pointer-events: none;
  }
  .particulas span {
    position: absolute;
    width: var(--t);
    height: var(--t);
    animation-iteration-count: infinite;
    animation-timing-function: steps(40);
  }

  /* Brasas que suben del volcán. */
  .brasas span {
    background: #ffb030;
    box-shadow: 0 0 6px #ff6a00;
    animation-name: subir;
  }
  .brasas span:nth-child(3n) {
    background: #ff5a1a;
  }
  @keyframes subir {
    from {
      transform: translate(0, 40vh);
      opacity: 0;
    }
    20% {
      opacity: 1;
    }
    to {
      transform: translate(3vw, -60vh);
      opacity: 0;
    }
  }

  /* Hojas que caen en la ruta. */
  .hojas span {
    width: calc(var(--t) * 2);
    background: #5aa83a;
    border-radius: 0 60% 0 60%;
    animation-name: caer;
  }
  .hojas span:nth-child(2n) {
    background: #8bc84a;
  }
  @keyframes caer {
    from {
      transform: translate(0, -60vh) rotate(0);
      opacity: 0;
    }
    15% {
      opacity: 1;
    }
    50% {
      transform: translate(4vw, 0) rotate(180deg);
    }
    to {
      transform: translate(-2vw, 60vh) rotate(360deg);
      opacity: 0;
    }
  }

  /* Luciérnagas en el bosque de noche. */
  .luciernagas span {
    background: #fff27a;
    border-radius: 50%;
    box-shadow: 0 0 10px 3px rgba(255, 240, 120, 0.7);
    animation-name: revolotear;
    animation-timing-function: ease-in-out;
  }
  @keyframes revolotear {
    0%,
    100% {
      transform: translate(0, 0);
      opacity: 0.1;
    }
    30% {
      opacity: 1;
    }
    50% {
      transform: translate(3vw, -4vh);
      opacity: 0.3;
    }
    70% {
      transform: translate(-2vw, -1vh);
      opacity: 1;
    }
  }

  /* Nieve en la cueva de hielo. */
  .nieve span {
    background: #ffffff;
    animation-name: nevar;
  }
  @keyframes nevar {
    from {
      transform: translate(0, -60vh);
      opacity: 0;
    }
    10% {
      opacity: 0.9;
    }
    to {
      transform: translate(-6vw, 60vh);
      opacity: 0.2;
    }
  }

  /* Burbujas en el río y el mar. */
  .burbujas span {
    border: 2px solid rgba(255, 255, 255, 0.8);
    border-radius: 50%;
    animation-name: subir;
  }

  /* Chispas en la llanura del trueno. */
  .chispas span {
    width: calc(var(--t) * 3);
    height: 3px;
    background: #fff27a;
    box-shadow: 0 0 8px #ffd000;
    animation-name: chispear;
    animation-timing-function: steps(1);
  }
  @keyframes chispear {
    0%,
    100% {
      opacity: 0;
    }
    8% {
      opacity: 1;
      transform: rotate(35deg);
    }
    12% {
      opacity: 0;
    }
    40% {
      opacity: 1;
      transform: rotate(-40deg) translate(10px, 4px);
    }
    44% {
      opacity: 0;
    }
  }

  /* Pétalos en Pueblo Paleta. */
  .petalos span {
    background: #ffc0d8;
    border-radius: 60% 0 60% 0;
    animation-name: caer;
  }
  .petalos span:nth-child(3n) {
    background: #fff4b0;
  }
</style>

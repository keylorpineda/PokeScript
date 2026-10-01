<script>
  // El ataque del compañero cuando un combate termina bien: Charmander y sus
  // evoluciones lanzan una bola de fuego, Pikachu y Raichu un rayo, la línea
  // de Bulbasaur una explosión de hojas y la de Squirtle burbujas por la boca.
  // Se dibuja encima del hábitat, a partir de dónde está el sprite.
  import { onMount } from 'svelte';
  import { POKEMON } from '../lib/pokemon.js';

  let { forma, habitat } = $props();

  // De dónde sale el ataque, como fracción del sprite (0,0 arriba a la
  // izquierda): la boca, o el bulbo en la línea de Bulbasaur. Los sprites
  // miran a la izquierda.
  const ANCLAS = {
    charmander: [0.12, 0.38],
    charmeleon: [0.12, 0.29],
    charizard: [0.07, 0.64],
    squirtle: [0.13, 0.42],
    wartortle: [0.15, 0.42],
    blastoise: [0.11, 0.23],
    bulbasaur: [0.7, 0.26],
    ivysaur: [0.55, 0.16],
    venusaur: [0.41, 0.14],
    pikachu: [0.4, 0.4],
    raichu: [0.54, 0.4],
  };
  const TIEMPO = 2400;
  // Centro del cuerpo (sin la cola) como fracción del sprite: la cola de
  // Pikachu va a la derecha y la de Raichu a la izquierda.
  const CENTROS = { pikachu: [0.34, 0.55], raichu: [0.57, 0.55] };

  const tipo = $derived(POKEMON[forma]?.tipo);
  let lugar = $state(null); // { x, y, ancho, alto, sx, sy, sw, sh } en px del hábitat
  let visible = $state(true);

  onMount(() => {
    const cuerpo = habitat?.querySelector('.cuerpo');
    if (!cuerpo) return;
    // Lo que va en position:absolute se mide desde dentro del borde del
    // hábitat, no desde su orilla de afuera.
    const r = habitat.getBoundingClientRect();
    const h = {
      left: r.left + habitat.clientLeft,
      top: r.top + habitat.clientTop,
      width: habitat.clientWidth,
      height: habitat.clientHeight,
    };
    const s = cuerpo.getBoundingClientRect();
    const [fx, fy] = ANCLAS[forma] ?? [0.3, 0.4];
    lugar = {
      x: s.left - h.left + fx * s.width,
      y: s.top - h.top + fy * s.height,
      ancho: h.width,
      alto: h.height,
      sx: s.left - h.left,
      sy: s.top - h.top,
      sw: s.width,
      sh: s.height,
      cx: s.left - h.left + (CENTROS[forma]?.[0] ?? 0.5) * s.width,
      cy: s.top - h.top + (CENTROS[forma]?.[1] ?? 0.52) * s.height,
    };
    const t = setTimeout(() => (visible = false), TIEMPO);
    return () => clearTimeout(t);
  });

  // Hacia dónde va: a la izquierda, hasta cerca del borde.
  const alcance = $derived(lugar ? Math.max(60, Math.min(170, lugar.x - 18)) : 0);

  // Partículas con valores fijos (no al azar), para que se vea igual siempre.
  const BRASAS = [0, 1, 2, 3, 4, 5, 6];
  const CHISPAS = [0, 45, 90, 135, 180, 225, 270, 315];
  const HOJAS = Array.from({ length: 14 }, (_, i) => {
    const angulo = -170 + (i * 160) / 13; // de izquierda a derecha, hacia arriba
    const dist = 58 + ((i * 37) % 5) * 9;
    return {
      dx: Math.cos((angulo * Math.PI) / 180) * dist,
      dy: Math.sin((angulo * Math.PI) / 180) * dist,
      giro: (i % 2 ? 1 : -1) * (220 + ((i * 53) % 4) * 90),
      color: ['#5fbf3a', '#3f9a2a', '#8fdc5a', '#4aa832'][i % 4],
      tam: 11 + (i % 3) * 3,
      espera: (i % 4) * 0.04,
    };
  });
  const BURBUJAS = Array.from({ length: 10 }, (_, i) => ({
    tam: 8 + ((i * 7) % 4) * 3.5,
    sube: -10 - ((i * 13) % 5) * 9,
    ola: 6 + (i % 3) * 4,
    espera: i * 0.09,
    largo: 0.75 + ((i * 11) % 3) * 0.12,
  }));

  // ─── Bola de fuego en pixel art ──────────────────────────────────────
  // Se dibuja píxel a píxel (20 × 14) con capas de color desde el centro:
  // blanco, crema, amarillo, naranja, naranja oscuro y rojo, y lenguas de
  // fuego irregulares hacia atrás. Dos cuadros que se turnan hacen que
  // parpadee. Se calcula avanzando a la derecha y se guarda volteada.
  const COLORES_FUEGO = ['#ffffff', '#fff4c4', '#ffe14a', '#ffa22a', '#ff6a1a', '#e8452c'];
  const RADIOS_FUEGO = [1.9, 3, 3.9, 4.8, 5.6, 6.3];
  // Largo extra de la cola en cada fila (hacia la izquierda), por cuadro.
  const COLAS = [
    [0, 1, 3, 2, 5, 4, 7, 6, 3, 5, 2, 3, 1, 0],
    [0, 2, 1, 4, 3, 6, 5, 8, 4, 3, 4, 1, 2, 0],
  ];
  function bolaPixel(cuadro) {
    const pixeles = [];
    const cx = 13.2;
    const cy = 6.5;
    for (let y = 0; y < 14; y++) {
      for (let x = 0; x < 20; x++) {
        const dx = x + 0.5 - cx;
        const dy = y + 0.5 - cy;
        // Hacia atrás (izquierda) el fuego se estira según la cola de la fila.
        const estirar = dx < 0 ? 1 + COLAS[cuadro][y] / 7 : 1;
        const d = Math.sqrt((dx / estirar) ** 2 + dy ** 2);
        const capa = RADIOS_FUEGO.findIndex((r) => d < r);
        if (capa < 0) continue;
        // El centro se corre un poco hacia adelante: el blanco queda al frente.
        const c = dx < -2 && capa < 2 ? capa + 1 : capa;
        // Se guarda volteada (x → 19 − x): la bola sale hacia la izquierda.
        pixeles.push({ x: 19 - x, y, color: COLORES_FUEGO[c] });
      }
    }
    return pixeles;
  }
  const CUADROS_FUEGO = [bolaPixel(0), bolaPixel(1)];
  const COLORES_BRASA = [
    '#ffe14a',
    '#ffa22a',
    '#ff6a1a',
    '#e8452c',
    '#fff4c4',
    '#ffa22a',
    '#e8452c',
  ];

  // ─── Rayos que envuelven al sprite ───────────────────────────────────
  // Tres juegos de rayos quebrados alrededor y por encima del cuerpo, que se
  // turnan. Los puntos caen en una rejilla de 3 px para que se vean
  // pixelados. Los valores salen de una semilla fija: siempre igual.
  function azar(semilla) {
    let s = semilla;
    return () => {
      s = (s * 1103515245 + 12345) % 2147483648;
      return s / 2147483648;
    };
  }
  const rejilla = (v) => Math.round(v / 3) * 3;
  function linea(puntos) {
    return puntos.map(([x, y]) => `${rejilla(x)},${rejilla(y)}`).join(' ');
  }
  function quebrar(x0, y0, x1, y1, tramos, amplitud, r) {
    const pts = [[x0, y0]];
    const largo = Math.hypot(x1 - x0, y1 - y0) || 1;
    const nx = -(y1 - y0) / largo;
    const ny = (x1 - x0) / largo;
    for (let i = 1; i < tramos; i++) {
      const t = i / tramos;
      const m = (r() * 2 - 1) * amplitud;
      pts.push([x0 + (x1 - x0) * t + nx * m, y0 + (y1 - y0) * t + ny * m]);
    }
    pts.push([x1, y1]);
    return pts;
  }
  function juegoDeRayos(l, semilla) {
    const r = azar(semilla);
    const cx = l.cx;
    const cy = l.cy;
    const rx = l.sw * 0.42;
    const ry = l.sh * 0.5;
    const rayos = [];
    for (let k = 0; k < 7; k++) {
      const a = ((k * 51 + semilla * 17 + r() * 14) * Math.PI) / 180;
      // Empiezan en el borde del cuerpo, no encima: el sprite se sigue viendo.
      const x0 = cx + Math.cos(a) * rx * 0.8;
      const y0 = cy + Math.sin(a) * ry * 0.8;
      const lejos = 1.45 + r() * 0.55;
      const x1 = cx + Math.cos(a) * rx * lejos;
      const y1 = cy + Math.sin(a) * ry * lejos;
      const pts = quebrar(x0, y0, x1, y1, 5, 7, r);
      rayos.push(linea(pts));
      // Una rama que sale de la mitad.
      const [mx, my] = pts[2];
      const b = a + (r() < 0.5 ? -1 : 1) * 0.7;
      rayos.push(
        linea(quebrar(mx, my, mx + Math.cos(b) * rx * 0.6, my + Math.sin(b) * ry * 0.6, 3, 4, r)),
      );
    }
    // Dos arcos cortos pegados al contorno, para que se vea envuelto.
    for (let k = 0; k < 2; k++) {
      const a = r() * Math.PI * 2;
      const x0 = cx + Math.cos(a) * rx * 0.95;
      const y0 = cy + Math.sin(a) * ry * 0.95;
      const x1 = cx + Math.cos(a + 1.1) * rx * 0.95;
      const y1 = cy + Math.sin(a + 1.1) * ry * 0.95;
      rayos.push(linea(quebrar(x0, y0, x1, y1, 4, 5, r)));
    }
    return rayos;
  }

  // El rayo: una línea quebrada desde arriba hasta la cabeza del sprite.
  function rayo(x0, y1, desfase) {
    const pasos = 6;
    const puntos = [];
    for (let i = 0; i <= pasos; i++) {
      const y = (y1 * i) / pasos;
      const x =
        x0 + (i === 0 || i === pasos ? 0 : (i % 2 ? 1 : -1) * (10 + ((i * 7) % 3) * 4)) + desfase;
      puntos.push(`${x.toFixed(1)},${y.toFixed(1)}`);
    }
    return puntos.join(' ');
  }
</script>

{#if lugar && visible}
  <div class="ataque {tipo}" aria-hidden="true">
    {#if tipo === 'fuego'}
      <div class="viaje" style="left:{lugar.x}px;top:{lugar.y}px;--dx:{-alcance}px;--dy:-14px">
        <span class="carga"></span>
        {#each BRASAS as b (b)}
          <span class="brasa" style="--i:{b};background:{COLORES_BRASA[b]}"></span>
        {/each}
        <span class="bola">
          <svg
            class="bola-pixel"
            width="40"
            height="28"
            viewBox="0 0 20 14"
            shape-rendering="crispEdges"
          >
            {#each CUADROS_FUEGO as cuadro, n (n)}
              <g class="cuadro c{n}">
                {#each cuadro as px (px.x + '-' + px.y)}
                  <rect x={px.x} y={px.y} width="1" height="1" fill={px.color} />
                {/each}
              </g>
            {/each}
          </svg>
        </span>
      </div>
      <div class="impacto" style="left:{lugar.x - alcance}px;top:{lugar.y - 14}px">
        <span class="anillo"></span>
        {#each CHISPAS as a (a)}
          <span class="chispa" style="--a:{a}deg"></span>
        {/each}
      </div>
    {:else if tipo === 'electrico'}
      <div class="penumbra"></div>
      <div
        class="aura"
        style="left:{lugar.cx}px;top:{lugar.cy}px;width:{lugar.sw * 2}px;height:{lugar.sh * 2}px"
      ></div>
      <svg class="rayos" width={lugar.ancho} height={lugar.alto} shape-rendering="crispEdges">
        <g class="juego j0">
          <polyline class="halo" points={rayo(lugar.x, lugar.y, 0)} />
          <polyline class="centro" points={rayo(lugar.x, lugar.y, 0)} />
        </g>
        {#each [11, 29, 47] as semilla, n (semilla)}
          <g class="juego j{n + 1}">
            {#each juegoDeRayos(lugar, semilla) as puntos, i (i)}
              <polyline class="halo" points={puntos} />
              <polyline class="centro" points={puntos} />
            {/each}
          </g>
        {/each}
      </svg>
    {:else if tipo === 'planta'}
      <div class="brote" style="left:{lugar.x}px;top:{lugar.y}px">
        <span class="onda"></span>
        {#each HOJAS as h, i (i)}
          <svg
            class="hoja"
            width={h.tam}
            height={h.tam * 0.6}
            viewBox="0 0 20 12"
            style="--dx:{h.dx}px;--dy:{h.dy}px;--giro:{h.giro}deg;animation-delay:{h.espera}s"
          >
            <path d="M0 6 C4 -1 14 -1 20 6 C14 13 4 13 0 6 Z" fill={h.color} />
            <path d="M1 6 L18 6" stroke="rgba(255,255,255,0.55)" stroke-width="1" />
          </svg>
        {/each}
      </div>
    {:else if tipo === 'agua'}
      <div class="chorro" style="left:{lugar.x}px;top:{lugar.y}px;--dx:{-alcance}px">
        {#each BURBUJAS as b, i (i)}
          <span
            class="burbuja"
            style="--tam:{b.tam}px;--sube:{b.sube}px;--ola:{b.ola}px;animation-delay:{b.espera}s;animation-duration:{b.largo}s"
          ></span>
        {/each}
      </div>
    {/if}
  </div>
{/if}

<style>
  .ataque {
    position: absolute;
    inset: 0;
    z-index: 3;
    pointer-events: none;
    overflow: hidden;
  }

  /* ─── Fuego ─────────────────────────────────────────────────────────── */
  .viaje {
    position: absolute;
    width: 0;
    height: 0;
  }
  .carga {
    position: absolute;
    left: -10px;
    top: -10px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: radial-gradient(circle, #fff 0 20%, #ffd54a 45%, rgba(255, 140, 20, 0) 70%);
    animation: cargar 0.3s ease-out both;
  }
  @keyframes cargar {
    from {
      transform: scale(0.2);
      opacity: 0;
    }
    70% {
      opacity: 1;
    }
    to {
      transform: scale(1.4);
      opacity: 0;
    }
  }
  .bola,
  .brasa {
    position: absolute;
    border-radius: 50%;
    offset-rotate: 0deg;
  }
  .bola {
    left: -20px;
    top: -14px;
    width: 40px;
    height: 28px;
    border-radius: 0;
    filter: drop-shadow(0 0 5px rgba(255, 140, 30, 0.8));
    animation: volar 0.62s cubic-bezier(0.35, 0.05, 0.6, 1) 0.22s both;
  }
  .bola-pixel {
    display: block;
  }
  .cuadro {
    animation: turno 0.16s steps(1) infinite;
  }
  .cuadro.c1 {
    animation-delay: -0.08s;
  }
  @keyframes turno {
    0% {
      opacity: 1;
    }
    50% {
      opacity: 0;
    }
  }
  .brasa {
    left: -3px;
    top: -3px;
    width: 6px;
    height: 6px;
    border-radius: 0;
    animation: estela 0.62s cubic-bezier(0.35, 0.05, 0.6, 1) both;
    animation-delay: calc(0.26s + var(--i) * 0.035s);
  }
  @keyframes volar {
    from {
      transform: translate(0, 0) scale(0.4);
      opacity: 0;
    }
    12% {
      opacity: 1;
      transform: translate(calc(var(--dx) * 0.1), calc(var(--dy) * 0.4)) scale(1);
    }
    55% {
      transform: translate(calc(var(--dx) * 0.55), calc(var(--dy) * 1.1)) scale(1.1);
    }
    92% {
      opacity: 1;
    }
    to {
      transform: translate(var(--dx), var(--dy)) scale(1.2);
      opacity: 0;
    }
  }
  @keyframes estela {
    from {
      transform: translate(0, 0) scale(0.9);
      opacity: 0;
    }
    15% {
      opacity: 0.9;
    }
    to {
      transform: translate(calc(var(--dx) * 0.9), var(--dy)) scale(0.15);
      opacity: 0;
    }
  }
  .impacto {
    position: absolute;
    width: 0;
    height: 0;
  }
  .anillo {
    position: absolute;
    left: -18px;
    top: -18px;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    border: 4px solid #ffb02e;
    box-shadow: 0 0 12px rgba(255, 120, 20, 0.8);
    animation: estallar 0.45s ease-out 0.82s both;
  }
  @keyframes estallar {
    from {
      transform: scale(0.2);
      opacity: 0;
    }
    30% {
      opacity: 1;
    }
    to {
      transform: scale(1.8);
      opacity: 0;
    }
  }
  .chispa {
    position: absolute;
    left: -3px;
    top: -3px;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #ffe36b;
    box-shadow: 0 0 6px #ff9a1a;
    transform: rotate(var(--a));
    animation: chispear 0.5s ease-out 0.84s both;
  }
  @keyframes chispear {
    from {
      transform: rotate(var(--a)) translateX(0);
      opacity: 0;
    }
    10% {
      opacity: 1;
    }
    to {
      transform: rotate(var(--a)) translateX(34px) scale(0.3);
      opacity: 0;
    }
  }

  /* ─── Eléctrico ─────────────────────────────────────────────────────── */
  /* El hábitat se oscurece un momento para que los rayos resalten. */
  .penumbra {
    position: absolute;
    inset: 0;
    background: rgba(24, 22, 8, 0.55);
    animation: oscurecer 1.5s ease-out both;
  }
  @keyframes oscurecer {
    from {
      opacity: 0;
    }
    12%,
    75% {
      opacity: 1;
    }
    to {
      opacity: 0;
    }
  }
  /* El resplandor amarillo que envuelve al sprite. */
  .aura {
    position: absolute;
    border-radius: 50%;
    background: radial-gradient(
      ellipse,
      rgba(255, 250, 190, 0.45) 0 25%,
      rgba(255, 230, 60, 0.3) 48%,
      rgba(255, 210, 0, 0) 70%
    );
    mix-blend-mode: screen;
    transform: translate(-50%, -50%);
    animation: aura 1.5s steps(10) both;
  }
  @keyframes aura {
    from {
      opacity: 0;
      transform: translate(-50%, -50%) scale(0.6);
    }
    15% {
      opacity: 1;
    }
    30% {
      opacity: 0.7;
      transform: translate(-50%, -50%) scale(1.05);
    }
    45% {
      opacity: 1;
    }
    75% {
      opacity: 0.85;
    }
    to {
      opacity: 0;
      transform: translate(-50%, -50%) scale(1.2);
    }
  }
  .rayos {
    position: absolute;
    left: 0;
    top: 0;
    overflow: visible;
  }
  .rayos polyline {
    fill: none;
    stroke-linejoin: miter;
    stroke-linecap: square;
  }
  .halo {
    stroke: #ffe14a;
    stroke-width: 4;
    opacity: 0.5;
  }
  .centro {
    stroke: #fffde8;
    stroke-width: 2;
  }
  .rayos {
    filter: drop-shadow(0 0 4px #ffe14a);
  }
  /* Los juegos de rayos se turnan: parpadean como electricidad. */
  .juego {
    opacity: 0;
    animation: 1.4s steps(1) both;
  }
  .j0 {
    animation-name: j0;
  }
  .j1 {
    animation-name: j1;
  }
  .j2 {
    animation-name: j2;
  }
  .j3 {
    animation-name: j3;
  }
  @keyframes j0 {
    0% {
      opacity: 0;
    }
    4% {
      opacity: 1;
    }
    12% {
      opacity: 0;
    }
    18% {
      opacity: 1;
    }
    26%,
    100% {
      opacity: 0;
    }
  }
  @keyframes j1 {
    0%,
    20% {
      opacity: 0;
    }
    22%,
    30% {
      opacity: 1;
    }
    34%,
    46% {
      opacity: 0;
    }
    48%,
    56% {
      opacity: 1;
    }
    60%,
    72% {
      opacity: 0;
    }
    74%,
    80% {
      opacity: 1;
    }
    84%,
    100% {
      opacity: 0;
    }
  }
  @keyframes j2 {
    0%,
    28% {
      opacity: 0;
    }
    30%,
    38% {
      opacity: 1;
    }
    42%,
    58% {
      opacity: 0;
    }
    60%,
    68% {
      opacity: 1;
    }
    72%,
    84% {
      opacity: 0;
    }
    86%,
    90% {
      opacity: 1;
    }
    94%,
    100% {
      opacity: 0;
    }
  }
  @keyframes j3 {
    0%,
    36% {
      opacity: 0;
    }
    38%,
    46% {
      opacity: 1;
    }
    50%,
    64% {
      opacity: 0;
    }
    66%,
    74% {
      opacity: 1;
    }
    78%,
    92% {
      opacity: 0;
    }
    94%,
    98% {
      opacity: 1;
    }
    100% {
      opacity: 0;
    }
  }

  /* ─── Planta ────────────────────────────────────────────────────────── */
  .brote {
    position: absolute;
    width: 0;
    height: 0;
  }
  .onda {
    position: absolute;
    left: -22px;
    top: -22px;
    width: 44px;
    height: 44px;
    border-radius: 50%;
    border: 4px solid rgba(143, 220, 90, 0.9);
    box-shadow: 0 0 12px rgba(95, 191, 58, 0.8);
    animation: estallar 0.55s ease-out both;
  }
  .hoja {
    position: absolute;
    left: -8px;
    top: -5px;
    overflow: visible;
    filter: drop-shadow(1px 1px 0 rgba(0, 0, 0, 0.35));
    animation: volarHoja 1.25s cubic-bezier(0.2, 0.7, 0.4, 1) both;
  }
  @keyframes volarHoja {
    from {
      transform: translate(0, 0) rotate(0deg) scale(0.3);
      opacity: 0;
    }
    12% {
      opacity: 1;
    }
    55% {
      transform: translate(var(--dx), var(--dy)) rotate(calc(var(--giro) * 0.6)) scale(1);
      opacity: 1;
    }
    to {
      transform: translate(calc(var(--dx) * 1.15), calc(var(--dy) + 34px)) rotate(var(--giro))
        scale(0.9);
      opacity: 0;
    }
  }

  /* ─── Agua ──────────────────────────────────────────────────────────── */
  .chorro {
    position: absolute;
    width: 0;
    height: 0;
  }
  .burbuja {
    position: absolute;
    left: calc(var(--tam) / -2);
    top: calc(var(--tam) / -2);
    width: var(--tam);
    height: var(--tam);
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.85);
    background: radial-gradient(
      circle at 34% 30%,
      rgba(255, 255, 255, 0.95) 0 14%,
      rgba(200, 238, 255, 0.55) 24%,
      rgba(100, 180, 245, 0.3) 62%,
      rgba(60, 130, 225, 0.65) 100%
    );
    box-shadow: 0 0 4px rgba(120, 200, 255, 0.7);
    animation-name: burbujear;
    animation-timing-function: ease-out;
    animation-fill-mode: both;
  }
  @keyframes burbujear {
    from {
      transform: translate(0, 0) scale(0.2);
      opacity: 0;
    }
    10% {
      opacity: 1;
    }
    35% {
      transform: translate(calc(var(--dx) * 0.35), calc(var(--sube) * 0.3 - var(--ola))) scale(1);
    }
    65% {
      transform: translate(calc(var(--dx) * 0.68), calc(var(--sube) * 0.7 + var(--ola))) scale(1);
    }
    88% {
      transform: translate(calc(var(--dx) * 0.9), var(--sube)) scale(1.05);
      opacity: 1;
    }
    to {
      transform: translate(var(--dx), var(--sube)) scale(1.6);
      opacity: 0;
    }
  }
</style>

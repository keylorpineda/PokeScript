// Sonidos retro generados con Web Audio (ondas cuadradas, triangulares y
// ruido, como el chip de la Game Boy): no hace falta ningún archivo.

let ctx = null;
let ruido = null;
let activo = true;
let ultimaTecla = 0;

export function sonidoActivo(valor) {
  activo = valor;
}

function audio() {
  ctx ??= new AudioContext();
  if (ctx.state === 'suspended') ctx.resume();
  return ctx;
}

// nota: [frecuencia, inicio, duración, onda, volumen, frecuencia final]
function nota(freq, inicio, dur, tipo = 'square', vol = 0.05, hasta = null) {
  const a = audio();
  const t = a.currentTime + inicio;
  const osc = a.createOscillator();
  const gan = a.createGain();
  osc.type = tipo;
  osc.frequency.setValueAtTime(freq, t);
  if (hasta) osc.frequency.exponentialRampToValueAtTime(hasta, t + dur);
  gan.gain.setValueAtTime(vol, t);
  gan.gain.exponentialRampToValueAtTime(0.0001, t + dur);
  osc.connect(gan).connect(a.destination);
  osc.start(t);
  osc.stop(t + dur + 0.02);
}

// golpeRuido: ráfaga de ruido filtrado, para golpes y explosiones.
function golpeRuido(inicio, dur, vol = 0.12, corte = 1800) {
  const a = audio();
  if (!ruido) {
    ruido = a.createBuffer(1, a.sampleRate * 0.6, a.sampleRate);
    const d = ruido.getChannelData(0);
    for (let i = 0; i < d.length; i++) d[i] = Math.random() * 2 - 1;
  }
  const t = a.currentTime + inicio;
  const fuente = a.createBufferSource();
  fuente.buffer = ruido;
  const filtro = a.createBiquadFilter();
  filtro.type = 'lowpass';
  filtro.frequency.setValueAtTime(corte, t);
  filtro.frequency.exponentialRampToValueAtTime(120, t + dur);
  const gan = a.createGain();
  gan.gain.setValueAtTime(vol, t);
  gan.gain.exponentialRampToValueAtTime(0.0001, t + dur);
  fuente.connect(filtro).connect(gan).connect(a.destination);
  fuente.start(t);
  fuente.stop(t + dur);
}

const arpegio = (notas, paso, dur = paso, tipo = 'square', vol = 0.045) =>
  notas.map((f, i) => [f, i * paso, dur, tipo, vol]);

const SONIDOS = {
  letra: () => nota(1500, 0, 0.018, 'square', 0.01),
  tecla: () => nota(2200 + Math.random() * 300, 0, 0.012, 'square', 0.006),
  mover: () => nota(880, 0, 0.045, 'square', 0.03),
  elegir: () => [...arpegio([988, 1319], 0.06)].forEach((n) => nota(...n)),
  abrir: () => arpegio([660, 880, 1175], 0.04, 0.05, 'square', 0.03).forEach((n) => nota(...n)),
  // ANALIZAR: la Pokédex se enciende (dos pitidos y un barrido) y cada fase que
  // pasa suena un poco más aguda, con eco.
  escaneo: () => {
    nota(1760, 0, 0.05, 'square', 0.035);
    nota(1760, 0.09, 0.05, 'square', 0.035);
    nota(330, 0.18, 0.35, 'sine', 0.06, 1320);
    nota(165, 0.18, 0.35, 'triangle', 0.05, 660);
  },
  fase: (i = 0) => {
    const base = [523, 659, 784][i] ?? 988;
    nota(base, 0, 0.08, 'square', 0.045);
    nota(base * 2, 0.06, 0.1, 'square', 0.03);
    nota(base * 2, 0.16, 0.08, 'square', 0.012);
  },
  linea: () => nota(1320, 0, 0.03, 'square', 0.018),
  pregunta: () => arpegio([784, 1175], 0.08, 0.1, 'square', 0.04).forEach((n) => nota(...n)),
  exito: () => {
    arpegio([784, 988, 1175, 1568], 0.09, 0.1).forEach((n) => nota(...n));
    nota(392, 0.27, 0.35, 'triangle', 0.07);
    nota(1568, 0.36, 0.3, 'square', 0.035);
  },
  victoria: () => {
    arpegio([523, 523, 523, 659, 784, 698, 784, 1047], 0.11, 0.1).forEach((n) => nota(...n));
    arpegio([262, 330, 392, 523], 0.22, 0.2, 'triangle', 0.07).forEach((n) => nota(...n));
  },
  error: () => {
    golpeRuido(0, 0.12, 0.08, 900);
    nota(220, 0.02, 0.14, 'sawtooth', 0.04, 150);
    nota(147, 0.14, 0.22, 'sawtooth', 0.04, 98);
  },
  golpe: () => {
    golpeRuido(0, 0.25, 0.16, 2400);
    nota(180, 0, 0.18, 'square', 0.05, 60);
  },
  huir: () =>
    arpegio([1175, 988, 784, 587, 440], 0.05, 0.06, 'square', 0.035).forEach((n) => nota(...n)),
  combate: () => {
    // Encuentro salvaje: una corrida cromática que sube alternando octavas,
    // el bajo marcando el pulso, el barrido de las franjas y un acorde final.
    for (let i = 0; i < 16; i++) {
      const f = 392 * Math.pow(2, i / 12);
      nota(i % 2 ? f * 2 : f, i * 0.052, 0.05, 'square', 0.035);
    }
    [0, 0.21, 0.42, 0.63].forEach((t, i) => nota(i % 2 ? 147 : 98, t, 0.18, 'triangle', 0.09));
    golpeRuido(0.78, 0.38, 0.07, 4000);
    [523, 659, 784, 1047].forEach((f) => nota(f, 0.9, 0.42, 'square', 0.026));
    nota(131, 0.9, 0.5, 'triangle', 0.1);
  },
  captura: () => {
    [0, 0.45, 0.9].forEach((t) => {
      nota(330, t, 0.06, 'square', 0.04);
      golpeRuido(t, 0.05, 0.05, 1200);
    });
    arpegio([1047, 1319, 1568], 0.08, 0.12).forEach(([f, i, d, o, v]) =>
      nota(f, 1.35 + i, d, o, v),
    );
  },
  oak: () =>
    arpegio([659, 784, 988, 1319], 0.07, 0.12, 'triangle', 0.08).forEach((n) => nota(...n)),
  tema: () => nota(300, 0, 0.35, 'square', 0.03, 1800),
};

export function sonar(nombre, ...args) {
  if (!activo) return;
  if (nombre === 'tecla') {
    const ahora = performance.now();
    if (ahora - ultimaTecla < 35) return;
    ultimaTecla = ahora;
  }
  try {
    SONIDOS[nombre]?.(...args);
  } catch {
    // Sin audio disponible: el IDE sigue igual.
  }
}

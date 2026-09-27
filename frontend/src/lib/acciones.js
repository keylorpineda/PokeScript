// Acciones del IDE: abrir el proyecto, compilar, ejecutar y responder a
// capturar. Cambian el estado compartido; los componentes solo lo pintan.
import { api } from './api.js';
import { ide, perfil, guardarPerfil, logro, navegacion } from './estado.svelte.js';
import { POKEMON } from './pokemon.js';
import { sonar } from './sonido.js';

// Carpeta del proyecto abierto (con Wails, una ruta del disco; en el simulador,
// el nombre del proyecto).
const RUTA_ACTUAL = () => ide.proyecto?.ruta ?? '';

// Pokémon invitado según el encabezado del diagnóstico, como en el README.
export const INVITADOS = {
  '¡Se escapó!': 'magikarp',
  '¡No pasó nada!': 'psyduck',
  'No es muy efectivo…': 'mewtwo',
  'No se encontró la ruta': 'abra',
  '¡Falló el ataque!': 'snorlax',
  '¿Seguro que quieres hacer eso?': 'porygon',
};

const companero = () => POKEMON[perfil.companero]?.nombre ?? 'Tu compañero';

export function decir(texto, animo = 'normal', invitado = null) {
  ide.asistente = { texto, animo, invitado, vez: (ide.asistente.vez ?? 0) + 1 };
}

export function oak(...lineas) {
  ide.oak = { lineas };
}

// recordar deja el proyecto de primero en la lista de recientes.
function recordar(ruta, nombre) {
  perfil.recientes = [
    { ruta, nombre, fecha: Date.now() },
    ...(perfil.recientes ?? []).filter((p) => p.ruta !== ruta),
  ].slice(0, 6);
  guardarPerfil();
}

export async function abrirProyecto(ruta) {
  const info = await api.leerProyecto(ruta);
  Object.assign(ide, {
    proyecto: { ruta, ...info },
    contenidos: {},
    sucios: {},
    diagnosticos: [],
    resultado: null,
    salida: [],
    esperandoEntrada: null,
  });
  for (const archivo of info.archivos) {
    ide.contenidos[archivo] = await api.leerArchivo(ruta, archivo);
  }
  ide.archivoActivo = info.principal;
  recordar(ruta, info.nombre);
  decir(
    `¡Hola, ${perfil.nombre}! Estoy listo. Escribe tu programa y presiona ANALIZAR cuando quieras que lo revise.`,
    'feliz',
  );
}

export async function guardarTodo() {
  for (const archivo of Object.keys(ide.sucios)) {
    await api.guardarArchivo(RUTA_ACTUAL(), archivo, ide.contenidos[archivo]);
  }
  ide.sucios = {};
}

function reaccionar(r) {
  ide.resultado = r;
  ide.diagnosticos = r.diagnosticos ?? [];
  const errores = ide.diagnosticos.filter((d) => d.severity === 'error');
  const avisos = ide.diagnosticos.length - errores.length;
  if (r.exito) {
    perfil.exp = (perfil.exp ?? 0) + 1;
    guardarPerfil();
    sonar('exito');
    decir(
      avisos
        ? `¡Es superefectivo! Compiló, aunque hay ${avisos} ${avisos === 1 ? 'aviso' : 'avisos'} que conviene revisar.`
        : '¡Es superefectivo! Tu programa compiló sin errores.',
      'feliz',
    );
    if (logro('primera-compilacion')) {
      oak(
        `¡Excelente, ${perfil.nombre}! Tu primer programa pasó la revisión completa.`,
        'El lexer, el parser y el analizador lo leyeron de principio a fin sin encontrar problemas.',
        '¡Ahora presiona COMBATE para verlo correr!',
      );
    }
  } else {
    const d = errores[0];
    sonar('error');
    const mas = errores.length > 1 ? ` (y ${errores.length - 1} más)` : '';
    decir(
      `${d.heading} Línea ${d.line}: ${d.desc}${d.suggest ? ' ' + d.suggest : ''}${mas}`,
      'triste',
      INVITADOS[d.heading],
    );
  }
}

export async function compilar() {
  if (ide.compilando) return;
  if (ide.ejecutando)
    return oak(
      '¡No es el momento de usar eso!',
      'Hay un combate en curso. Presiona HUIR antes de volver a analizar.',
    );
  ide.compilando = true;
  sonar('escaneo');
  const inicio = Date.now();
  try {
    await guardarTodo();
    const r = await api.compilar(RUTA_ACTUAL());
    // La animación del escaneo dura al menos un momento.
    await new Promise((res) => setTimeout(res, Math.max(0, 1300 - (Date.now() - inicio))));
    reaccionar(r);
  } finally {
    ide.compilando = false;
  }
}

export async function ejecutar() {
  if (ide.ejecutando)
    return oak(
      '¡No es el momento de usar eso!',
      'Ya hay un combate en curso. Termínalo o presiona HUIR.',
    );
  await guardarTodo();
  // La salida se prepara antes de llamar al backend: el programa puede emitir
  // (por ejemplo, la pregunta de un capturar) antes de que vuelva la
  // respuesta, y esos mensajes no se deben perder.
  api.reiniciarOrden();
  const anterior = ide.salida;
  ide.salida = [
    { tipo: 'sistema', texto: `¡${perfil.nombre} y ${companero()} entran en combate!` },
  ];
  ide.ejecutando = true;
  const r = await api.ejecutar(RUTA_ACTUAL());
  reaccionar(r);
  if (!r.exito) {
    ide.salida = anterior;
    ide.ejecutando = false;
    return;
  }
  ide.transicion = (ide.transicion ?? 0) + 1;
  sonar('combate');
  if (ide.esperandoEntrada === null) decir('¡Vamos! Mira la salida del combate abajo.', 'feliz');
}

export async function detener() {
  if (!ide.ejecutando)
    return oak('¡No es el momento de usar eso!', 'No hay ningún combate del cual huir.');
  await api.detener();
}

export async function responder(texto) {
  if (ide.esperandoEntrada === null) return;
  // La respuesta queda en la misma línea que la pregunta, como en una terminal.
  const ultima = ide.salida.findLast((l) => l.tipo === 'pregunta' && l.respuesta === undefined);
  if (ultima) ultima.respuesta = texto;
  else ide.salida.push({ tipo: 'respuesta', texto });
  ide.esperandoEntrada = null;
  sonar('elegir');
  await api.enviarEntrada(texto);
}

let escuchando = false;
export function escucharEjecucion() {
  if (escuchando) return;
  escuchando = true;
  api.alEvento('salida', (e) => {
    ide.salida.push({ tipo: 'texto', texto: e.texto });
  });
  api.alEvento('pedir-entrada', (e) => {
    ide.salida.push({ tipo: 'pregunta', texto: e.texto });
    ide.esperandoEntrada = e.texto;
    sonar('pregunta');
    decir(
      'El programa te está preguntando algo. Escribe tu respuesta en la salida.',
      'normal',
      'chansey',
    );
  });
  api.alEvento('error-ejecucion', (e) => {
    const d = e.diagnostico;
    ide.salida.push({ tipo: 'error', texto: `${d.heading} Línea ${d.line}: ${d.desc}` });
    sonar('golpe');
    decir(`${d.heading} ${d.desc}`, 'triste', 'snorlax');
  });
  api.alEvento('fin-ejecucion', (e) => {
    ide.ejecutando = false;
    ide.esperandoEntrada = null;
    const fin = {
      terminado: 'El combate terminó.',
      detenido: '¡Escapaste sin problemas!',
      error: 'El combate terminó con un error.',
    };
    ide.salida.push({ tipo: 'sistema', texto: fin[e.texto] ?? 'Fin.' });
    if (e.texto === 'detenido') sonar('huir');
    else if (e.texto === 'terminado') sonar('victoria');
    if (e.texto === 'terminado') {
      decir('¡Ganamos! El programa terminó sin problemas.', 'feliz');
      if (logro('primer-combate')) {
        oak(
          `¡Felicidades, ${perfil.nombre}! Tu primer programa corrió de principio a fin.`,
          'Cada gran entrenador empezó con un solo combate. ¡Sigue así!',
        );
      }
    }
  });
}

export function aplicarArreglo(d) {
  if (!d.fix) return;
  ide.arreglo = { ...d.fix, archivo: d.file, vez: Date.now() };
  sonar('elegir');
  decir('¡Listo! Apliqué el arreglo. Presiona ANALIZAR para revisar de nuevo.', 'feliz');
}

export function irA(d) {
  ide.archivoActivo = d.file;
  ide.ir = { archivo: d.file, linea: d.line, col: d.col, len: d.len, vez: Date.now() };
}

// ─── Gestor de archivos ────────────────────────────────────────────────────

async function refrescar() {
  const info = await api.leerProyecto(RUTA_ACTUAL());
  ide.proyecto = { ruta: RUTA_ACTUAL(), ...info };
}

// nombrePks agrega .pks si hace falta y quita espacios.
const nombrePks = (n) => {
  const limpio = n.trim().replace(/\s+/g, '_');
  return limpio.endsWith('.pks') ? limpio : `${limpio}.pks`;
};

async function intentar(accion, exito) {
  try {
    await accion();
    sonar('elegir');
    if (exito) decir(exito, 'feliz');
    return true;
  } catch (e) {
    sonar('error');
    decir(`No se pudo: ${e?.message ?? e}.`, 'triste', 'psyduck');
    return false;
  }
}

// nuevoArchivo crea el archivo y, si viene de una plantilla, le escribe su
// contenido inicial.
export async function nuevoArchivo(nombre, contenido = '') {
  const archivo = nombrePks(nombre);
  const ok = await intentar(async () => {
    await api.nuevoArchivo(RUTA_ACTUAL(), archivo);
    if (contenido) await api.guardarArchivo(RUTA_ACTUAL(), archivo, contenido);
  }, `¡Nuevo archivo en la mochila: ${archivo}!`);
  if (!ok) return;
  ide.contenidos[archivo] = contenido;
  await refrescar();
  ide.archivoActivo = archivo;
}

// duplicarArchivo copia un archivo con el primer nombre libre: x_copia,
// x_copia2…
export async function duplicarArchivo(archivo) {
  const base = archivo.replace(/\.pks$/, '');
  const existentes = new Set(ide.proyecto?.archivos ?? []);
  let nombre = `${base}_copia`;
  for (let n = 2; existentes.has(`${nombre}.pks`); n++) nombre = `${base}_copia${n}`;
  await nuevoArchivo(nombre, ide.contenidos[archivo] ?? '');
}

export async function renombrarArchivo(de, nombre) {
  const a = nombrePks(nombre);
  if (a === de) return;
  await guardarTodo();
  const ok = await intentar(
    () => api.renombrarArchivo(RUTA_ACTUAL(), de, a),
    `Ahora se llama ${a}.`,
  );
  if (!ok) return;
  ide.contenidos[a] = ide.contenidos[de];
  delete ide.contenidos[de];
  if (ide.archivoActivo === de) ide.archivoActivo = a;
  await refrescar();
}

export async function borrarArchivo(archivo) {
  const ok = await intentar(
    () => api.borrarArchivo(RUTA_ACTUAL(), archivo),
    `Soltaste ${archivo}. ¡Adiós, ${archivo.replace(/\.pks$/, '')}!`,
  );
  if (!ok) return;
  delete ide.contenidos[archivo];
  delete ide.sucios[archivo];
  await refrescar();
  if (ide.archivoActivo === archivo) ide.archivoActivo = ide.proyecto.principal;
}

export async function marcarPrincipal(archivo) {
  const ok = await intentar(
    () => api.marcarPrincipal(RUTA_ACTUAL(), archivo),
    `${archivo} ahora es el archivo principal: ahí debe vivir el combate.`,
  );
  if (ok) await refrescar();
}

// ─── Proyectos ─────────────────────────────────────────────────────────────

// crearProyecto crea un proyecto nuevo con un combate de saludo. Con Wails
// primero se elige dónde guardarlo; en el simulador vive en memoria.
export async function crearProyecto(nombre) {
  let carpeta = '';
  if (api.enWails()) {
    carpeta = await api.elegirCarpeta();
    if (!carpeta) return null;
  }
  const info = await api.crearProyecto(carpeta, nombre.trim().replace(/\s+/g, '_'));
  recordar(info.ruta, info.nombre);
  return info.ruta;
}

// buscarProyecto abre el diálogo de carpetas del sistema (solo con Wails).
export async function buscarProyecto() {
  return api.enWails() ? await api.elegirCarpeta() : null;
}

// Proyectos que se pueden abrir sin buscar: los recientes y, en el
// simulador, los de ejemplo.
export function proyectosConocidos() {
  const recientes = perfil.recientes ?? [];
  if (api.enWails()) return recientes;
  const ejemplos = ['centro_pokemon', 'combate'].map((r) => ({ ruta: r, nombre: r }));
  return [...recientes, ...ejemplos.filter((e) => !recientes.some((p) => p.ruta === e.ruta))];
}

// salirAlMenu guarda todo, corta el combate si hay uno y vuelve al menú del
// título para abrir o crear otro proyecto.
export async function salirAlMenu() {
  if (ide.ejecutando) await api.detener();
  await guardarTodo();
  sonar('huir');
  navegacion.pantalla = 'menu';
}

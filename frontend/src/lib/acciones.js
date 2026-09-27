// Acciones del IDE: abrir el proyecto, compilar, ejecutar y responder a
// capturar. Cambian el estado compartido; los componentes solo lo pintan.
import { api } from './api.js';
import { ide, perfil, guardarPerfil, logro } from './estado.svelte.js';
import { POKEMON } from './pokemon.js';
import { sonar } from './sonido.js';

const RUTA = ''; // El simulador no usa ruta; con Wails será la carpeta abierta.

// Pokémon invitado según el encabezado del diagnóstico, como en el README.
const INVITADOS = {
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

export async function abrirProyecto() {
  const info = await api.leerProyecto(RUTA);
  ide.proyecto = { ruta: RUTA, ...info };
  for (const archivo of info.archivos) {
    ide.contenidos[archivo] = await api.leerArchivo(RUTA, archivo);
  }
  ide.archivoActivo = info.principal;
  decir(
    `¡Hola, ${perfil.nombre}! Estoy listo. Escribe tu programa y presiona ANALIZAR cuando quieras que lo revise.`,
    'feliz',
  );
}

export async function guardarTodo() {
  for (const archivo of Object.keys(ide.sucios)) {
    await api.guardarArchivo(RUTA, archivo, ide.contenidos[archivo]);
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
    const r = await api.compilar(RUTA);
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
  const anterior = ide.salida;
  ide.salida = [
    { tipo: 'sistema', texto: `¡${perfil.nombre} y ${companero()} entran en combate!` },
  ];
  ide.ejecutando = true;
  const r = await api.ejecutar(RUTA);
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
    sonar('linea');
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
  const info = await api.leerProyecto(RUTA);
  ide.proyecto = { ruta: RUTA, ...info };
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

export async function nuevoArchivo(nombre) {
  const archivo = nombrePks(nombre);
  const ok = await intentar(
    () => api.nuevoArchivo(RUTA, archivo),
    `¡Nuevo archivo en la mochila: ${archivo}!`,
  );
  if (!ok) return;
  ide.contenidos[archivo] = '';
  await refrescar();
  ide.archivoActivo = archivo;
}

export async function renombrarArchivo(de, nombre) {
  const a = nombrePks(nombre);
  if (a === de) return;
  await guardarTodo();
  const ok = await intentar(() => api.renombrarArchivo(RUTA, de, a), `Ahora se llama ${a}.`);
  if (!ok) return;
  ide.contenidos[a] = ide.contenidos[de];
  delete ide.contenidos[de];
  if (ide.archivoActivo === de) ide.archivoActivo = a;
  await refrescar();
}

export async function borrarArchivo(archivo) {
  const ok = await intentar(
    () => api.borrarArchivo(RUTA, archivo),
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
    () => api.marcarPrincipal(RUTA, archivo),
    `${archivo} ahora es el archivo principal: ahí debe vivir el combate.`,
  );
  if (ok) await refrescar();
}

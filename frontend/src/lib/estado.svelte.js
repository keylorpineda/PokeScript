// Estado compartido del IDE (sección 9: proyecto, archivoActivo,
// diagnosticos, salida, esperandoEntrada, estadoAsistente) más el perfil del
// entrenador, que se guarda en el navegador.
import { aplicarTema, TEMAS } from './temas.js';
import { INICIALES, forma, nivelDe } from './pokemon.js';
import { sonidoActivo } from './sonido.js';

const CLAVE = 'pokescript-perfil';

function cargarPerfil() {
  try {
    return JSON.parse(localStorage.getItem(CLAVE)) ?? null;
  } catch {
    return null;
  }
}

const guardado = cargarPerfil();

export const perfil = $state({
  nombre: guardado?.nombre ?? '',
  companero: INICIALES.includes(guardado?.companero) ? guardado.companero : null,
  tema: TEMAS[guardado?.tema] ? guardado.tema : 'rojo-fuego',
  sonido: guardado?.sonido ?? true,
  logros: guardado?.logros ?? {},
  recientes: guardado?.recientes ?? [], // { ruta, nombre, fecha }
  exp: guardado?.exp ?? 0, // compilaciones exitosas: suben de nivel al compañero
});

export function guardarPerfil() {
  try {
    localStorage.setItem(CLAVE, JSON.stringify(perfil));
  } catch {
    // Sin almacenamiento: el perfil dura solo esta sesión.
  }
}

export function cambiarTema(id) {
  perfil.tema = id;
  aplicarTema(id);
  guardarPerfil();
}

export function cambiarSonido(valor) {
  perfil.sonido = valor;
  sonidoActivo(valor);
  guardarPerfil();
}

aplicarTema(perfil.tema);
sonidoActivo(perfil.sonido);

export const ide = $state({
  proyecto: null, // { ruta, nombre, principal, archivos }
  archivoActivo: null,
  contenidos: {}, // archivo → texto en el editor
  sucios: {}, // archivo → true si tiene cambios sin guardar
  diagnosticos: [],
  resultado: null, // último ResultadoCompilacion
  salida: [], // { tipo: 'texto' | 'pregunta' | 'respuesta' | 'error' | 'sistema', texto }
  esperandoEntrada: null, // mensaje de capturar, o null
  ejecutando: false,
  compilando: false,
  // Lo que dice el compañero: { animo, texto, invitado, vez }
  asistente: { animo: 'feliz', texto: '', invitado: null, vez: 0 },
  transicion: 0, // sube cada vez que empieza un combate
  pokedex: false, // la Pokédex de consulta está abierta
  guardado: 0, // cuándo se guardó por última vez
  ir: null, // { archivo, linea, col, len }: el editor salta ahí
  arreglo: null, // { archivo, line, col, len, replacement }: el editor lo aplica
  // Oak aparece encima de todo cuando hay algo importante que decir.
  oak: null, // { lineas: [] }
  cursor: { linea: 1, col: 1 },
  // Bloques abiertos donde está el cursor, de afuera hacia adentro
  // («combate», «recorrer», «si»…), para la ruta de arriba del editor.
  ruta: [],
  // Nombre que está bajo el cursor, para mostrar su tipo en el cuadro.
  palabra: '',
  debilitado: false, // el último combate terminó con un error de ejecución
  celebracion: 0, // sube cada vez que un combate termina bien
  finTransicion: 0, // cuándo termina la transición del último combate
  evolucion: null, // { de, a, vez } mientras se ve la evolución
});

// miPokemon es la forma actual del compañero: el inicial o su evolución
// según el nivel.
export const miPokemon = () => forma(perfil.companero, nivelDe(perfil.exp));

// logro marca un logro y devuelve true solo la primera vez.
export function logro(nombre) {
  if (perfil.logros[nombre]) return false;
  perfil.logros[nombre] = true;
  guardarPerfil();
  return true;
}

// Pantalla actual del IDE y el proyecto elegido en el menú del título.
// archivo: el .pks que se eligió en el explorador, { ruta, nombre }, para
// abrirlo en lugar del principal.
export const navegacion = $state({ pantalla: 'titulo', ruta: null, aviso: '', archivo: null });

// El explorador (PC de Bill) se abre desde cualquier pantalla y devuelve la
// carpeta elegida, o null si se cierra sin elegir.
export const explorador = $state({ abierto: false, modo: 'abrir', nombre: '', resolver: null });

export function pedirCarpeta(modo = 'abrir', nombre = '') {
  return new Promise((resolver) => {
    Object.assign(explorador, { abierto: true, modo, nombre, resolver });
  });
}

export function cerrarExplorador(ruta = null) {
  const r = explorador.resolver;
  Object.assign(explorador, { abierto: false, resolver: null });
  r?.(ruta);
}

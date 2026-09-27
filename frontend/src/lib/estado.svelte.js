// Estado compartido del IDE (sección 9: proyecto, archivoActivo,
// diagnosticos, salida, esperandoEntrada, estadoAsistente) más el perfil del
// entrenador, que se guarda en el navegador.
import { aplicarTema, TEMAS } from './temas.js';
import { INICIALES } from './pokemon.js';
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
  ir: null, // { archivo, linea, col, len }: el editor salta ahí
  arreglo: null, // { archivo, line, col, len, replacement }: el editor lo aplica
  // Oak aparece encima de todo cuando hay algo importante que decir.
  oak: null, // { lineas: [] }
  cursor: { linea: 1, col: 1 },
});

// logro marca un logro y devuelve true solo la primera vez.
export function logro(nombre) {
  if (perfil.logros[nombre]) return false;
  perfil.logros[nombre] = true;
  guardarPerfil();
  return true;
}

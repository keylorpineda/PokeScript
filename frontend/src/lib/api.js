// Puente con el backend. Dentro de Wails usa los métodos de app.go
// (window.go.main.App) y sus eventos; en el navegador, sin Wails, usa un
// backend simulado para poder diseñar la interfaz.

import CONSULTA from './consulta.json';

const wails = () => window.go?.main?.App;

// ─── Backend simulado ──────────────────────────────────────────────────────

const EJEMPLO = {
  'principal.pks': `// Centro Pokémon: cura a tu Pokémon y revisa si puede combatir.
enseñar CURA desde "constantes.pks"

movimiento roca curar(roca vida)
    entregar vida + CURA
fin

combate
    planta nombre
    capturar(nombre, "¿Cómo se llama tu Pokémon? ")

    roca vida = 35
    vida = curar(vida)
    gritar "¡", nombre, " fue curado! Vida: ", vida

    si vida > 50
        gritar "¡Listo para el combate!"
    sino
        gritar "Necesita descansar."
    fin
fin
`,
  'constantes.pks': `medalla roca CURA = 20
`,
};

// El proyecto de ejemplo del repositorio (sección 10), leído en tiempo de
// compilación.
const COMBATE = Object.fromEntries(
  Object.entries(
    import.meta.glob('../../../ejemplos/combate/*.pks', {
      query: '?raw',
      import: 'default',
      eager: true,
    }),
  ).map(([ruta, texto]) => [ruta.split('/').pop(), texto]),
);

const CLAVE_SIMULADO = 'pokescript-simulado';

const simulado = {
  // Los proyectos del simulador se guardan en el navegador para que sigan ahí
  // al recargar la página.
  proyectos: (() => {
    const base = {
      centro_pokemon: { principal: 'principal.pks', archivos: { ...EJEMPLO } },
      combate: { principal: 'principal.pks', archivos: { ...COMBATE } },
    };
    try {
      return { ...base, ...JSON.parse(localStorage.getItem(CLAVE_SIMULADO) ?? '{}') };
    } catch {
      return base;
    }
  })(),
  guardar() {
    try {
      localStorage.setItem(CLAVE_SIMULADO, JSON.stringify(this.proyectos));
    } catch {
      // Sin almacenamiento: los cambios duran solo esta sesión.
    }
  },
  oyentes: {},
  entrada: null,

  proyecto(ruta) {
    const p = this.proyectos[ruta];
    if (!p) throw new Error(`no existe el proyecto ${ruta}`);
    return p;
  },
  async CrearProyecto(_carpeta, nombre) {
    if (this.proyectos[nombre]) throw new Error(`ya existe el proyecto ${nombre}`);
    this.proyectos[nombre] = {
      principal: 'principal.pks',
      archivos: { 'principal.pks': 'combate\n    gritar "¡Hola, mundo!"\nfin\n' },
    };
    return { ruta: nombre, ...(await this.LeerProyecto(nombre)) };
  },
  // El menú de consulta viene del backend: consulta.json lo genera
  // `pnpm docs:consulta` desde internal/consulta.
  async ObtenerPalabrasReservadas() {
    return CONSULTA.palabras;
  },
  async ObtenerTablaEfectividades() {
    return CONSULTA.efectividades;
  },
  async OtrosArchivos(ruta) {
    return ruta === 'centro_pokemon' ? ['notas.txt', 'portada.png'] : [];
  },
  async PlantillaPrincipal() {
    return '// principal.pks · aquí empieza tu programa\ncombate\n    gritar "¡Hola, mundo Pokémon!"\nfin\n';
  },
  // Explorador simulado: una sola carpeta con los proyectos del simulador.
  async LugaresExplorador() {
    return [{ nombre: 'Proyectos de ejemplo', ruta: '/', tipo: 'personal' }];
  },
  async ListarCarpeta(ruta) {
    if (ruta === '/') {
      return {
        ruta: '/',
        nombre: 'Proyectos de ejemplo',
        padre: '',
        proyecto: false,
        recortada: false,
        carpetas: Object.entries(this.proyectos).map(([n, p]) => ({
          nombre: n,
          ruta: n,
          proyecto: true,
          pks: Object.keys(p.archivos).length,
        })),
        archivos: [],
      };
    }
    const p = this.proyecto(ruta);
    return {
      ruta,
      nombre: ruta,
      padre: '/',
      proyecto: true,
      recortada: false,
      carpetas: [],
      archivos: [
        ...Object.keys(p.archivos).map((n) => ({ nombre: n, pks: true })),
        ...(await this.OtrosArchivos(ruta)).map((n) => ({ nombre: n, pks: false })),
      ],
    };
  },
  async ElegirCarpeta() {
    return null; // Sin Wails no hay diálogo de carpetas.
  },
  async LeerProyecto(ruta) {
    const p = this.proyecto(ruta);
    return { nombre: ruta, principal: p.principal, archivos: Object.keys(p.archivos).sort() };
  },
  async LeerArchivo(ruta, archivo) {
    return this.proyecto(ruta).archivos[archivo] ?? '';
  },
  async GuardarArchivo(ruta, archivo, contenido) {
    this.proyecto(ruta).archivos[archivo] = contenido;
  },
  async NuevoArchivo(ruta, archivo) {
    const p = this.proyecto(ruta);
    if (p.archivos[archivo] !== undefined) throw new Error(`ya existe ${archivo}`);
    p.archivos[archivo] = '';
  },
  async RenombrarArchivo(ruta, de, a) {
    const p = this.proyecto(ruta);
    if (p.archivos[a] !== undefined) throw new Error(`ya existe ${a}`);
    p.archivos[a] = p.archivos[de];
    delete p.archivos[de];
    if (p.principal === de) p.principal = a;
  },
  async BorrarArchivo(ruta, archivo) {
    const p = this.proyecto(ruta);
    if (archivo === p.principal) throw new Error('no se puede borrar el archivo principal');
    delete p.archivos[archivo];
  },
  async MarcarPrincipal(ruta, archivo) {
    this.proyecto(ruta).principal = archivo;
  },
  async CompilarProyecto(ruta) {
    await espera(250);
    const p = this.proyecto(ruta);
    const diagnosticos = [];
    for (const [archivo, texto] of Object.entries(p.archivos)) {
      texto.split('\n').forEach((l, i) => {
        const col = l.indexOf('curra');
        if (col >= 0) {
          diagnosticos.push({
            severity: 'error',
            category: 'semantico',
            code: 'nombre-no-declarado',
            heading: '¡No pasó nada!',
            file: archivo,
            line: i + 1,
            col: col + 1,
            len: 5,
            desc: '«curra» no está declarado.',
            cause: 'el nombre no coincide con ningún dato, movimiento o medalla visible.',
            suggest: '¿Quisiste decir «curar»?',
            fix: { line: i + 1, col: col + 1, len: 5, replacement: 'curar' },
            leccion: {
              titulo: 'Nombres que no existen',
              texto:
                'Antes de usar un nombre, tiene que estar declarado: un dato con su tipo, un movimiento o una medalla. Revisa que esté bien escrito: vida y Vida son nombres distintos.',
            },
          });
        }
      });
      const abre = (
        texto.match(/^\s*(combate|movimiento|si|mientras|recorrer|segun|especie|ficha)\b/gm) ?? []
      ).length;
      const cierra = (texto.match(/^\s*fin\b/gm) ?? []).length;
      if (abre > cierra) {
        const n = texto.split('\n').length;
        diagnosticos.push({
          severity: 'error',
          category: 'sintactico',
          code: 'bloque-sin-cerrar',
          heading: '¡Se escapó!',
          file: archivo,
          line: n,
          col: 1,
          len: 1,
          desc: 'un bloque quedó abierto: falta su fin.',
          cause: 'cada combate, movimiento, si, mientras o recorrer se cierra con fin.',
          suggest: 'Agrega fin donde termina el bloque.',
          leccion: {
            titulo: 'Cada bloque tiene su fin',
            texto:
              'Los bloques se abren con una palabra (si, mientras, movimiento…) y se cierran con fin. La sangría ayuda a ver cuál quedó abierto.',
          },
        });
      }
    }
    const exito = !diagnosticos.some((d) => d.severity === 'error');
    return {
      exito,
      encabezado: exito ? '¡Es superefectivo!' : diagnosticos[0].heading,
      principal: p.principal,
      archivos: Object.keys(p.archivos),
      diagnosticos,
      simbolos: {
        'principal.pks': [
          { nombre: 'curar', clase: 'movimiento', tipo: 'roca', detalle: 'curar(roca vida)' },
          { nombre: 'CURA', clase: 'medalla', tipo: 'roca' },
          { nombre: 'nombre', clase: 'dato', tipo: 'planta' },
          { nombre: 'vida', clase: 'dato', tipo: 'roca' },
        ],
      },
    };
  },
  async EjecutarProyecto(ruta) {
    const r = await this.CompilarProyecto(ruta);
    if (!r.exito) return r;
    this.simular(this.proyecto(ruta));
    return r;
  },
  // simular recorre el combate del principal de arriba abajo, sin evaluar de
  // verdad: pregunta en cada capturar y muestra cada gritar con los datos que
  // conoce. Alcanza para diseñar la interfaz; el intérprete real es el de Go.
  async simular(p) {
    const texto = p.archivos[p.principal] ?? '';
    const cuerpo = texto
      .split('\n')
      .slice(texto.split('\n').findIndex((l) => /^\s*combate\b/.test(l)) + 1);
    const datos = {};
    const valor = (parte) => {
      const t = parte.trim();
      if (/^".*"$/.test(t)) return t.slice(1, -1);
      if (/^\d+(\.\d+)?$/.test(t)) return t;
      return datos[t] ?? `«${t}»`;
    };
    for (const linea of cuerpo) {
      const pide = linea.match(/capturar\((\w+),\s*"([^"]*)"\)/);
      const grita = linea.match(/^\s*gritar\s+(.*)$/);
      const asigna = linea.match(/^\s*(?:roca|agua|planta|fuego|electrico)\s+(\w+)\s*=\s*(.+)$/);
      if (pide) {
        this.emitir('pedir-entrada', { tipo: 'pedir-entrada', texto: pide[2] });
        const r = await new Promise((res) => (this.entrada = res));
        if (r === null)
          return this.emitir('fin-ejecucion', { tipo: 'fin-ejecucion', texto: 'detenido' });
        datos[pide[1]] = r;
      } else if (asigna) {
        datos[asigna[1]] = valor(asigna[2]);
      } else if (grita) {
        await espera(300);
        const partes = grita[1].match(/"[^"]*"|[^,]+/g) ?? [];
        this.emitir('salida', { tipo: 'salida', texto: partes.map(valor).join('') });
      }
    }
    this.emitir('fin-ejecucion', { tipo: 'fin-ejecucion', texto: 'terminado' });
  },
  async EnviarEntrada(texto) {
    const r = this.entrada;
    this.entrada = null;
    r?.(texto);
  },
  async DetenerEjecucion() {
    const r = this.entrada;
    this.entrada = null;
    r?.(null);
  },
  emitir(evento, datos) {
    for (const f of this.oyentes[evento] ?? []) f(datos);
  },
  on(evento, f) {
    (this.oyentes[evento] ??= []).push(f);
  },
};

// Lo que cambia un proyecto del simulador se guarda enseguida.
for (const m of [
  'CrearProyecto',
  'GuardarArchivo',
  'NuevoArchivo',
  'RenombrarArchivo',
  'BorrarArchivo',
  'MarcarPrincipal',
]) {
  const original = simulado[m];
  simulado[m] = async function (...args) {
    const r = await original.apply(this, args);
    this.guardar();
    return r;
  };
}

const espera = (ms) => new Promise((r) => setTimeout(r, ms));

// ─── Interfaz única para los componentes ───────────────────────────────────

const llamar = (metodo, ...args) =>
  wails() ? wails()[metodo](...args) : simulado[metodo](...args);

export const api = {
  enWails: () => Boolean(wails()),
  crearProyecto: (carpeta, nombre) => llamar('CrearProyecto', carpeta, nombre),
  // Abre el diálogo de carpetas del sistema; null si se cancela o sin Wails.
  elegirCarpeta: () => llamar('ElegirCarpeta'),
  lugaresExplorador: () => llamar('LugaresExplorador'),
  listarCarpeta: (ruta) => llamar('ListarCarpeta', ruta),
  leerProyecto: (ruta) => llamar('LeerProyecto', ruta),
  otrosArchivos: (ruta) => llamar('OtrosArchivos', ruta),
  plantillaPrincipal: () => llamar('PlantillaPrincipal'),
  palabrasReservadas: () => llamar('ObtenerPalabrasReservadas'),
  tablaEfectividades: () => llamar('ObtenerTablaEfectividades'),
  leerArchivo: (ruta, archivo) => llamar('LeerArchivo', ruta, archivo),
  guardarArchivo: (ruta, archivo, contenido) => llamar('GuardarArchivo', ruta, archivo, contenido),
  nuevoArchivo: (ruta, archivo) => llamar('NuevoArchivo', ruta, archivo),
  renombrarArchivo: (ruta, de, a) => llamar('RenombrarArchivo', ruta, de, a),
  borrarArchivo: (ruta, archivo) => llamar('BorrarArchivo', ruta, archivo),
  marcarPrincipal: (ruta, archivo) => llamar('MarcarPrincipal', ruta, archivo),
  compilar: (ruta) => llamar('CompilarProyecto', ruta),
  ejecutar: (ruta) => llamar('EjecutarProyecto', ruta),
  enviarEntrada: (texto) => llamar('EnviarEntrada', texto),
  detener: () => llamar('DetenerEjecucion'),
  // Eventos de la ejecución: salida, pedir-entrada, error-ejecucion, fin-ejecucion.
  alEvento(evento, f) {
    const recibir = (e) => enOrden(evento, e, f);
    if (window.runtime?.EventsOn) window.runtime.EventsOn(evento, recibir);
    else simulado.on(evento, recibir);
  },
  // Cada ejecución numera sus eventos desde 1 (ver EventoNumerado en app.go).
  reiniciarOrden() {
    esperado = 1;
    pendientes.clear();
  },
};

// Wails puede entregar los eventos desordenados: los que llegan antes de
// tiempo esperan a los que faltan. Los del simulador no traen número y pasan
// directo.
let esperado = 1;
const pendientes = new Map();
function enOrden(evento, e, f) {
  if (e?.n == null) return f(e);
  pendientes.set(e.n, () => f(e));
  while (pendientes.has(esperado)) {
    const siguiente = pendientes.get(esperado);
    pendientes.delete(esperado);
    esperado += 1;
    siguiente();
  }
}

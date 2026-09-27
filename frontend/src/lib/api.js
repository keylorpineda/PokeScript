// Puente con el backend. Dentro de Wails usa los métodos de app.go
// (window.go.main.App) y sus eventos; en el navegador, sin Wails, usa un
// backend simulado para poder diseñar la interfaz.

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

const simulado = {
  archivos: { ...EJEMPLO },
  principal: 'principal.pks',
  oyentes: {},
  entrada: null,

  async LeerProyecto() {
    return {
      nombre: 'centro_pokemon',
      principal: this.principal,
      archivos: Object.keys(this.archivos).sort(),
    };
  },
  async LeerArchivo(_ruta, archivo) {
    return this.archivos[archivo] ?? '';
  },
  async GuardarArchivo(_ruta, archivo, contenido) {
    this.archivos[archivo] = contenido;
  },
  async NuevoArchivo(_ruta, archivo) {
    if (this.archivos[archivo] !== undefined) throw new Error(`ya existe ${archivo}`);
    this.archivos[archivo] = '';
  },
  async RenombrarArchivo(_ruta, de, a) {
    if (this.archivos[a] !== undefined) throw new Error(`ya existe ${a}`);
    this.archivos[a] = this.archivos[de];
    delete this.archivos[de];
    if (this.principal === de) this.principal = a;
  },
  async BorrarArchivo(_ruta, archivo) {
    if (archivo === this.principal) throw new Error('no se puede borrar el archivo principal');
    delete this.archivos[archivo];
  },
  async MarcarPrincipal(_ruta, archivo) {
    this.principal = archivo;
  },
  async CompilarProyecto() {
    await espera(250);
    const diagnosticos = [];
    for (const [archivo, texto] of Object.entries(this.archivos)) {
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
      principal: 'principal.pks',
      archivos: Object.keys(this.archivos),
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
    (async () => {
      this.emitir('pedir-entrada', { tipo: 'pedir-entrada', texto: '¿Cómo se llama tu Pokémon? ' });
      const nombre = await new Promise((res) => (this.entrada = res));
      if (nombre === null)
        return this.emitir('fin-ejecucion', { tipo: 'fin-ejecucion', texto: 'detenido' });
      for (const t of [`¡${nombre} fue curado! Vida: 55`, '¡Listo para el combate!']) {
        await espera(350);
        this.emitir('salida', { tipo: 'salida', texto: t });
      }
      this.emitir('fin-ejecucion', { tipo: 'fin-ejecucion', texto: 'terminado' });
    })();
    return r;
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

const espera = (ms) => new Promise((r) => setTimeout(r, ms));

// ─── Interfaz única para los componentes ───────────────────────────────────

const llamar = (metodo, ...args) =>
  wails() ? wails()[metodo](...args) : simulado[metodo](...args);

export const api = {
  enWails: () => Boolean(wails()),
  leerProyecto: (ruta) => llamar('LeerProyecto', ruta),
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
    if (window.runtime?.EventsOn) window.runtime.EventsOn(evento, f);
    else simulado.on(evento, f);
  },
};

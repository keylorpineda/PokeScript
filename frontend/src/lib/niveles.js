// Los «niveles» del editor: sangría automática por bloques, líneas guía de
// colores, plegado de bloques y la pareja de cada bloque (si ↔ sino ↔ fin).
// Todo sale de bloques.js, que lee la estructura del texto.
import { EditorView, Decoration, ViewPlugin } from '@codemirror/view';
import { StateField, RangeSetBuilder } from '@codemirror/state';
import {
  indentService,
  foldService,
  foldGutter,
  foldKeymap,
  getIndentation,
  getIndentUnit,
  indentString,
  IndentContext,
} from '@codemirror/language';
import { insertNewlineAndIndent } from '@codemirror/commands';
import { ABREN, analizar, clase, palabras, sangria } from './bloques.js';

const lineasDe = (doc) => doc.toString().split('\n');

// La estructura del documento, que se vuelve a calcular solo cuando cambia el texto.
const estructura = StateField.define({
  create: (estado) => analizar(lineasDe(estado.doc), estado.tabSize),
  update: (valor, tr) => (tr.docChanged ? analizar(lineasDe(tr.newDoc), tr.state.tabSize) : valor),
});

// ─── Sangría automática ─────────────────────────────────────────────────────

// Enter, indentSelection e indentOnInput preguntan aquí cuánta sangría lleva
// una línea. `cx.lineAt` ya tiene en cuenta el salto de línea simulado de
// Enter, así que el texto de antes llega hasta donde está el cursor.
const sangriaPorBloques = indentService.of((cx, pos) => {
  const linea = cx.lineAt(pos, 1);
  const antes = cx.state.doc.sliceString(0, linea.from).split('\n');
  // El último trozo es el pedazo de línea que queda antes del corte: si está
  // vacío (Enter al final de una línea), no aporta nada.
  if (antes[antes.length - 1] === '') antes.pop();
  return sangria(antes, linea.text, getIndentUnit(cx.state), cx.state.tabSize);
});

// Al terminar una línea que empieza con fin o sino, primero se acomoda esa
// línea en el nivel de su bloque y después se abre la siguiente. Así se
// corrige también si se escribió «final…» y la línea se movió antes de
// tiempo.
function enterConNiveles(vista) {
  const { state } = vista;
  const sel = state.selection.main;
  if (!sel.empty) return false;
  const linea = state.doc.lineAt(sel.head);
  const primera = palabras(linea.text)[0] ?? '';
  if (primera.startsWith('fin') || primera.startsWith('sino')) {
    const cx = new IndentContext(state);
    const quiere = getIndentation(cx, linea.from);
    const actual = /^\s*/.exec(linea.text)[0];
    if (quiere !== null) {
      const nueva = indentString(state, quiere);
      if (nueva !== actual) {
        vista.dispatch({
          changes: { from: linea.from, to: linea.from + actual.length, insert: nueva },
          userEvent: 'input.indent',
        });
      }
    }
  }
  return insertNewlineAndIndent(vista);
}

// ─── Líneas guía ────────────────────────────────────────────────────────────

// Cada nivel lleva el color de un tipo, en ronda.
const COLORES = ['--t-agua', '--t-planta', '--t-fuego', '--t-electrico', '--t-roca'];
const RELLENO = 10; // el padding izquierdo de .cm-line en pokescript.js

function guiasDeLinea(bloques, dentro) {
  if (!dentro.length) return null;
  const capas = dentro.map((b, n) => {
    const color = `color-mix(in srgb, var(${COLORES[n % COLORES.length]}) 55%, transparent)`;
    return {
      imagen: `linear-gradient(${color}, ${color})`,
      pos: `calc(${RELLENO}px + ${bloques[b].col}ch) 0`,
    };
  });
  return Decoration.line({
    attributes: {
      class: 'cm-guia-niveles',
      style:
        `background-image: ${capas.map((c) => c.imagen).join(', ')};` +
        `background-position: ${capas.map((c) => c.pos).join(', ')};`,
    },
  });
}

const guias = ViewPlugin.fromClass(
  class {
    constructor(vista) {
      this.decorations = this.construir(vista);
    }
    update(u) {
      if (u.docChanged || u.viewportChanged) this.decorations = this.construir(u.view);
    }
    construir(vista) {
      const { info, bloques } = vista.state.field(estructura);
      const b = new RangeSetBuilder();
      for (const { from, to } of vista.visibleRanges) {
        for (let pos = from; pos <= to;) {
          const linea = vista.state.doc.lineAt(pos);
          const i = linea.number - 1;
          // La guía de un bloque baja por su cuerpo: no se dibuja en la línea
          // que lo abre ni en su fin (dentro ya los excluye).
          const deco = info[i] ? guiasDeLinea(bloques, info[i].dentro) : null;
          if (deco) b.add(linea.from, linea.from, deco);
          pos = linea.to + 1;
        }
      }
      return b.finish();
    }
  },
  { decorations: (v) => v.decorations },
);

// ─── Plegado ────────────────────────────────────────────────────────────────

// Un bloque se pliega desde el final de la línea que lo abre hasta el final
// de la línea anterior a su fin, así el fin queda a la vista.
const plegado = foldService.of((estado, desde) => {
  const linea = estado.doc.lineAt(desde);
  const { bloques } = estado.field(estructura);
  const b = bloques.find((x) => x.abre === linea.number - 1);
  if (!b || b.cierra === null || b.cierra - b.abre < 2) return null;
  return { from: linea.to, to: estado.doc.line(b.cierra).to };
});

// ─── Pareja del bloque ──────────────────────────────────────────────────────

const marcaPareja = Decoration.mark({ class: 'cm-pareja-bloque' });

// rangoClave ubica la palabra que marca la línea: la primera (si, sino,
// fin…) o, en «X entonces si …», la que abre el bloque.
function rangoClave(linea) {
  const texto = linea.text;
  const ps = palabras(texto);
  let objetivo = ps[0];
  if (clase(texto) === 'abre' && !ABREN.has(objetivo)) objetivo = ps[ps.indexOf('entonces') + 1];
  const re = new RegExp(`(^|[^\\p{L}\\p{Nd}_])(${objetivo})(?![\\p{L}\\p{Nd}_])`, 'u');
  const m = re.exec(texto);
  if (!m) return null;
  const inicio = linea.from + m.index + m[1].length;
  return { from: inicio, to: inicio + objetivo.length };
}

const pareja = ViewPlugin.fromClass(
  class {
    constructor(vista) {
      this.decorations = this.construir(vista.state);
    }
    update(u) {
      if (u.docChanged || u.selectionSet) this.decorations = this.construir(u.state);
    }
    construir(estado) {
      const sel = estado.selection.main;
      const n = estado.doc.lineAt(sel.head).number - 1;
      const { bloques } = estado.field(estructura);
      const b = bloques.find((x) => x.abre === n || x.cierra === n || x.medios.includes(n));
      if (!b || b.cierra === null) return Decoration.none;
      const rangos = [b.abre, ...b.medios, b.cierra]
        .map((i) => rangoClave(estado.doc.line(i + 1)))
        .filter(Boolean)
        .sort((x, y) => x.from - y.from);
      return Decoration.set(rangos.map((r) => marcaPareja.range(r.from, r.to)));
    }
  },
  { decorations: (v) => v.decorations },
);

// ─── Estilos ────────────────────────────────────────────────────────────────

const estilos = EditorView.theme({
  '.cm-guia-niveles': {
    backgroundRepeat: 'no-repeat',
    backgroundSize: '2px 100%',
  },
  '.cm-pareja-bloque': {
    backgroundColor: 'var(--editor-seleccion)',
    outline: '1px solid var(--acento)',
    borderRadius: '2px',
  },
  '.cm-foldGutter .cm-gutterElement': {
    cursor: 'pointer',
    color: 'var(--texto-suave)',
    padding: '0 4px',
  },
  '.cm-foldGutter .cm-gutterElement:hover': { color: 'var(--acento)' },
  '.cm-foldPlaceholder': {
    backgroundColor: 'var(--editor-seleccion)',
    border: '1px solid var(--acento)',
    color: 'var(--editor-texto)',
    padding: '0 6px',
    margin: '0 4px',
  },
});

// atajosNiveles va en el keymap de Editor.svelte después de completionKeymap
// (su Enter acepta una sugerencia) y antes de defaultKeymap (su Enter no
// acomoda el fin).
export const atajosNiveles = [{ key: 'Enter', run: enterConNiveles }, ...foldKeymap];

export const niveles = [
  estructura,
  sangriaPorBloques,
  guias,
  plegado,
  foldGutter({ openText: '▾', closedText: '▸' }),
  pareja,
  estilos,
];

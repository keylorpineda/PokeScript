<script>
  // Editor de código (CodeMirror 6). Un EditorState por archivo, para que cada
  // pestaña conserve su historial de deshacer; los diagnósticos se subrayan
  // con Line/Col/Len del compilador.
  import { onMount, onDestroy, untrack } from 'svelte';
  import {
    EditorView,
    keymap,
    lineNumbers,
    highlightActiveLine,
    highlightActiveLineGutter,
    drawSelection,
  } from '@codemirror/view';
  import { EditorState } from '@codemirror/state';
  import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands';
  import { indentOnInput, bracketMatching } from '@codemirror/language';
  import { autocompletion, completionKeymap, closeBrackets } from '@codemirror/autocomplete';
  import { lintGutter, setDiagnostics } from '@codemirror/lint';
  import { pokescript, desplazamiento, completar } from '../lib/pokescript.js';
  import { ide } from '../lib/estado.svelte.js';
  import { compilar, ejecutar } from '../lib/acciones.js';
  import { sonar } from '../lib/sonido.js';

  let contenedor;
  let vista;
  const estados = {};
  const scrolls = {}; // dónde iba cada archivo al cambiar de pestaña
  let mostrado = null;

  function crearEstado(archivo) {
    return EditorState.create({
      doc: ide.contenidos[archivo] ?? '',
      extensions: [
        lineNumbers(),
        highlightActiveLineGutter(),
        highlightActiveLine(),
        drawSelection(),
        history(),
        indentOnInput(),
        bracketMatching(),
        closeBrackets(),
        autocompletion({ override: [completar(() => ide.resultado?.simbolos?.[archivo])] }),
        lintGutter(),
        EditorState.tabSize.of(4),
        keymap.of([
          { key: 'F5', run: () => (ejecutar(), true) },
          { key: 'Mod-Enter', run: () => (compilar(), true) },
          ...completionKeymap,
          ...defaultKeymap,
          ...historyKeymap,
          indentWithTab,
        ]),
        pokescript,
        EditorView.updateListener.of((u) => {
          if (u.docChanged) {
            if (u.transactions.some((tr) => tr.isUserEvent('input') || tr.isUserEvent('delete')))
              sonar('tecla');
            ide.contenidos[archivo] = u.state.doc.toString();
            ide.sucios[archivo] = true;
          }
          if (u.selectionSet || u.docChanged) {
            const pos = u.state.selection.main.head;
            const l = u.state.doc.lineAt(pos);
            ide.cursor = {
              linea: l.number,
              col: Array.from(l.text.slice(0, pos - l.from)).length + 1,
            };
          }
        }),
      ],
    });
  }

  function mostrar(archivo) {
    if (!vista || !archivo || archivo === mostrado) return;
    if (mostrado) {
      estados[mostrado] = vista.state;
      scrolls[mostrado] = vista.scrollDOM.scrollTop;
    }
    estados[archivo] ??= crearEstado(archivo);
    vista.setState(estados[archivo]);
    vista.scrollDOM.scrollTop = scrolls[archivo] ?? 0;
    mostrado = archivo;
    pintarDiagnosticos();
    vista.focus();
  }

  function pintarDiagnosticos() {
    if (!vista || !mostrado) return;
    const doc = vista.state.doc;
    const lista = ide.diagnosticos
      .filter((d) => d.file === mostrado)
      .map((d) => {
        const from = desplazamiento(doc, d.line, d.col);
        const to = Math.max(from + 1, desplazamiento(doc, d.line, d.col + Math.max(1, d.len)));
        return {
          from,
          to: Math.min(to, doc.length),
          severity: d.severity === 'error' ? 'error' : 'warning',
          message: `${d.heading}\n${d.desc}${d.suggest ? '\n' + d.suggest : ''}`,
        };
      });
    vista.dispatch(setDiagnostics(vista.state, lista));
  }

  onMount(() => {
    vista = new EditorView({ parent: contenedor });
    mostrar(ide.archivoActivo);
  });
  onDestroy(() => vista?.destroy());

  $effect(() => {
    const a = ide.archivoActivo;
    untrack(() => mostrar(a));
  });

  $effect(() => {
    ide.diagnosticos;
    untrack(pintarDiagnosticos);
  });

  // Saltar a un diagnóstico.
  $effect(() => {
    const ir = ide.ir;
    if (!ir) return;
    untrack(() => {
      mostrar(ir.archivo);
      const doc = vista.state.doc;
      const from = desplazamiento(doc, ir.linea, ir.col);
      const to = desplazamiento(doc, ir.linea, ir.col + Math.max(1, ir.len));
      vista.dispatch({
        selection: { anchor: from, head: to },
        effects: EditorView.scrollIntoView(from, { y: 'center' }),
      });
      vista.focus();
    });
  });

  // Aplicar la corrección del asistente.
  $effect(() => {
    const f = ide.arreglo;
    if (!f) return;
    untrack(() => {
      mostrar(f.archivo);
      const doc = vista.state.doc;
      const from = desplazamiento(doc, f.line, f.col);
      const to = desplazamiento(doc, f.line, f.col + f.len);
      vista.dispatch({
        changes: { from, to, insert: f.replacement },
        selection: { anchor: from + f.replacement.length },
      });
      vista.focus();
    });
  });
</script>

<div class="editor" bind:this={contenedor}></div>

<style>
  .editor {
    height: 100%;
    overflow: hidden;
    user-select: text;
  }
</style>

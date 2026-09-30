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
  import { indentOnInput, bracketMatching, indentUnit } from '@codemirror/language';
  import { autocompletion, completionKeymap, closeBrackets } from '@codemirror/autocomplete';
  import { lintGutter, setDiagnostics } from '@codemirror/lint';
  import { pokescript, desplazamiento, completar } from '../lib/pokescript.js';
  import { niveles, atajosNiveles } from '../lib/niveles.js';
  import { ide } from '../lib/estado.svelte.js';
  import { compilar, ejecutar, programarGuardado, guardarAhora } from '../lib/acciones.js';
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
        // Tab y Shift+Tab mueven 4 espacios, la sangría del lenguaje.
        indentUnit.of('    '),
        keymap.of([
          { key: 'F5', run: () => (ejecutar(), true) },
          { key: 'Mod-Enter', run: () => (compilar(), true) },
          { key: 'Mod-s', preventDefault: true, run: () => (guardarAhora(), true) },
          ...completionKeymap,
          ...atajosNiveles,
          ...defaultKeymap,
          ...historyKeymap,
          indentWithTab,
        ]),
        pokescript,
        niveles,
        EditorView.updateListener.of((u) => {
          if (u.docChanged) {
            ide.contenidos[archivo] = u.state.doc.toString();
            ide.sucios[archivo] = true;
            programarGuardado();
          }
          if (u.selectionSet || u.docChanged) marcarCursor(u.state);
        }),
      ],
    });
  }

  // marcarCursor pone en la barra de abajo la línea y columna del cursor.
  function marcarCursor(estado) {
    const pos = estado.selection.main.head;
    const l = estado.doc.lineAt(pos);
    ide.cursor = {
      linea: l.number,
      col: Array.from(l.text.slice(0, pos - l.from)).length + 1,
    };
  }

  function mostrar(archivo) {
    if (!vista || !archivo || archivo === mostrado) return;
    if (mostrado) {
      estados[mostrado] = vista.state;
      scrolls[mostrado] = vista.scrollDOM.scrollTop;
    }
    estados[archivo] ??= crearEstado(archivo);
    vista.setState(estados[archivo]);
    // Al cambiar de archivo (o de proyecto) no hay evento de selección: sin
    // esto, la barra seguiría mostrando la posición del archivo anterior.
    marcarCursor(vista.state);
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
      // Se usa una sola vez: si quedara guardado, el próximo editor que se
      // monte (por ejemplo, al abrir otro proyecto) volvería a saltar.
      ide.ir = null;
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
      // Se aplica una sola vez: si quedara guardado, el próximo editor que se
      // monte lo volvería a aplicar sobre otro archivo con el mismo nombre.
      ide.arreglo = null;
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

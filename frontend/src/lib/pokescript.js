// PokeScript para CodeMirror 6: resaltado (cada tipo con el color de su tipo
// Pokémon), tema del editor con las variables del tema del IDE,
// autocompletado y conversión de posiciones de los diagnósticos.
import { StreamLanguage, HighlightStyle, syntaxHighlighting } from '@codemirror/language';
import { Tag, tags as t } from '@lezer/highlight';
import { EditorView } from '@codemirror/view';

const et = {
  roca: Tag.define(),
  agua: Tag.define(),
  fuego: Tag.define(),
  planta: Tag.define(),
  electrico: Tag.define(),
  coleccion: Tag.define(),
  estructura: Tag.define(),
  control: Tag.define(),
  funcion: Tag.define(),
  constante: Tag.define(),
  tipoNombrado: Tag.define(),
};

const TIPOS = new Set(['roca', 'agua', 'fuego', 'planta', 'electrico']);
const COLECCION = new Set([
  'equipo',
  'mochila',
  'posible',
  'de',
  'a',
  'especie',
  'ficha',
  'medalla',
]);
const ESTRUCTURA = new Set(['combate', 'movimiento', 'entregar', 'fin', 'enseñar', 'desde']);
const CONTROL = new Set([
  'si',
  'sino',
  'segun',
  'entonces',
  'otro',
  'mientras',
  'recorrer',
  'en',
  'hasta',
  'huir',
  'siguiente',
]);
const FUNCIONES = new Set([
  'gritar',
  'capturar',
  'convertir',
  'tamaño',
  'aleatorio',
  'redondear',
  'sumar',
  'quitar',
  'contiene',
]);
const OPERADORES = new Set(['resto', 'igual', 'diferente', 'y', 'o', 'no']);
const LITERALES = new Set(['verdadero', 'falso', 'fantasma']);

export const RESERVADAS = [
  ...TIPOS,
  ...COLECCION,
  ...ESTRUCTURA,
  ...CONTROL,
  ...FUNCIONES,
  ...OPERADORES,
  ...LITERALES,
];

const lenguaje = StreamLanguage.define({
  name: 'pokescript',
  token(stream) {
    if (stream.eatSpace()) return null;
    if (stream.match('//')) {
      stream.skipToEnd();
      return 'comment';
    }
    if (stream.match(/^"(?:[^"\\]|\\.)*"?/)) return 'string';
    if (stream.match(/^'(?:[^'\\]|\\.)*'?/)) return 'character';
    if (stream.match(/^\d+\.\d+/) || stream.match(/^\d+/)) return 'number';
    if (stream.match(/^[\p{L}][\p{L}\p{Nd}_]*/u)) {
      const w = stream.current();
      if (TIPOS.has(w)) return w;
      if (ESTRUCTURA.has(w)) return 'estructura';
      if (CONTROL.has(w)) return 'control';
      if (FUNCIONES.has(w)) return 'funcion';
      if (OPERADORES.has(w)) return 'operatorKeyword';
      if (LITERALES.has(w)) return 'bool';
      if (COLECCION.has(w)) return 'coleccion';
      if (/^[\p{Lu}][\p{Lu}\p{Nd}_]+$/u.test(w)) return 'constante';
      if (/^\p{Lu}/u.test(w)) return 'tipoNombrado';
      return 'variableName';
    }
    if (stream.match(/^(>=|<=|[+\-*/=<>])/)) return 'operator';
    stream.next();
    return 'punctuation';
  },
  tokenTable: {
    roca: et.roca,
    agua: et.agua,
    fuego: et.fuego,
    planta: et.planta,
    electrico: et.electrico,
    coleccion: et.coleccion,
    estructura: et.estructura,
    control: et.control,
    funcion: et.funcion,
    constante: et.constante,
    tipoNombrado: et.tipoNombrado,
    operatorKeyword: t.operatorKeyword,
    bool: t.bool,
    character: t.character,
  },
  languageData: { commentTokens: { line: '//' } },
});

// Cada tipo de dato se pinta con el color de su tipo Pokémon.
const chip = (color) => ({ color: `var(${color})`, fontWeight: '700' });

const resaltado = HighlightStyle.define([
  { tag: et.roca, ...chip('--t-roca') },
  { tag: et.agua, ...chip('--t-agua') },
  { tag: et.fuego, ...chip('--t-fuego') },
  { tag: et.planta, ...chip('--t-planta') },
  { tag: et.electrico, ...chip('--t-electrico') },
  { tag: et.coleccion, color: 'var(--nombre_tipo)', fontWeight: '700' },
  { tag: et.estructura, color: 'var(--palabra)', fontWeight: '700' },
  { tag: et.control, color: 'var(--control)', fontWeight: '700' },
  { tag: et.funcion, color: 'var(--constante)', fontWeight: '700' },
  { tag: et.constante, color: 'var(--constante)' },
  { tag: et.tipoNombrado, color: 'var(--nombre_tipo)', fontWeight: '700' },
  { tag: t.operatorKeyword, color: 'var(--control)' },
  { tag: t.bool, color: 'var(--numero)', fontWeight: '700' },
  { tag: t.number, color: 'var(--numero)' },
  { tag: [t.string, t.character], color: 'var(--texto_lit)' },
  { tag: t.comment, color: 'var(--comentario)', fontStyle: 'italic' },
  { tag: t.operator, color: 'var(--texto-suave)' },
  { tag: t.punctuation, color: 'var(--texto-suave)' },
]);

const tema = EditorView.theme({
  '&': {
    height: '100%',
    fontSize: 'var(--tam-codigo, 15px)',
    backgroundColor: 'var(--editor-fondo)',
    color: 'var(--editor-texto)',
  },
  '.cm-scroller': {
    fontFamily: 'var(--codigo)',
    lineHeight: '1.7',
    fontVariantLigatures: 'none',
  },
  '.cm-line': { padding: '0 16px 0 10px' },
  '.cm-content': { caretColor: 'var(--acento)', padding: '10px 0' },
  '.cm-cursor, .cm-dropCursor': { borderLeft: '3px solid var(--acento)' },
  '&.cm-focused': { outline: 'none' },
  '.cm-gutters': {
    backgroundColor: 'var(--editor-gutter)',
    color: 'var(--texto-suave)',
    border: 'none',
    borderRight: '3px solid var(--borde-suave)',
    fontFamily: 'var(--titulo)',
    fontSize: '15px',
  },
  '.cm-activeLine': { backgroundColor: 'var(--editor-linea)' },
  '.cm-activeLineGutter': {
    backgroundColor: 'var(--editor-linea)',
    color: 'var(--acento)',
    fontWeight: '700',
  },
  '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, ::selection': {
    backgroundColor: 'var(--editor-seleccion) !important',
  },
  '.cm-matchingBracket': {
    backgroundColor: 'var(--editor-seleccion)',
    outline: '1px solid var(--acento)',
  },
  '.cm-tooltip': {
    border: '3px solid var(--borde)',
    backgroundColor: 'var(--panel)',
    color: 'var(--texto)',
    fontFamily: 'var(--cuerpo)',
    boxShadow: '4px 4px 0 var(--borde)',
  },
  '.cm-tooltip-autocomplete > ul': { fontFamily: 'var(--codigo)', maxHeight: '14em' },
  '.cm-tooltip-autocomplete > ul > li[aria-selected]': {
    backgroundColor: 'var(--acento)',
    color: 'var(--acento-texto)',
  },
  '.cm-completionDetail': { fontStyle: 'normal', opacity: '0.7', marginLeft: '8px' },
  '.cm-diagnostic': { fontFamily: 'var(--cuerpo)', fontSize: '14px', padding: '6px 10px' },
  '.cm-diagnostic-error': { borderLeft: '5px solid var(--error)' },
  '.cm-diagnostic-warning': { borderLeft: '5px solid var(--aviso)' },
  '.cm-lintRange-error': {
    backgroundImage: 'none',
    textDecoration: 'underline wavy var(--error)',
    textDecorationThickness: '2px',
    textUnderlineOffset: '4px',
  },
  '.cm-lintRange-warning': {
    backgroundImage: 'none',
    textDecoration: 'underline wavy var(--aviso)',
    textDecorationThickness: '2px',
    textUnderlineOffset: '4px',
  },
  '.cm-lint-marker': { width: '16px', height: '16px', content: 'none' },
  '.cm-lint-marker-error': {
    content: 'none',
    background: 'url(/objetos/poke-ball.png) center / 18px no-repeat',
    imageRendering: 'pixelated',
  },
  '.cm-lint-marker-warning': {
    content: 'none',
    background: 'url(/objetos/great-ball.png) center / 18px no-repeat',
    imageRendering: 'pixelated',
  },
});

export const pokescript = [lenguaje, syntaxHighlighting(resaltado), tema];

// desplazamiento convierte línea y columna (base 1, en runas, como las da
// el compilador) a una posición de CodeMirror (en unidades UTF-16).
export function desplazamiento(doc, linea, col) {
  const n = Math.min(Math.max(1, linea), doc.lines);
  const l = doc.line(n);
  const runas = Array.from(l.text);
  const antes = runas.slice(0, Math.max(0, col - 1)).join('');
  return Math.min(l.from + antes.length, l.to);
}

// Fuente de autocompletado: palabras reservadas y los símbolos que devolvió
// la última compilación para este archivo.
export function completar(simbolos) {
  const clase = {
    movimiento: 'function',
    medalla: 'constant',
    especie: 'type',
    ficha: 'type',
    valor: 'enum',
    dato: 'variable',
    parametro: 'variable',
  };
  return (ctx) => {
    const palabra = ctx.matchBefore(/[\p{L}][\p{L}\p{Nd}_]*/u);
    if (!palabra || (palabra.from === palabra.to && !ctx.explicit)) return null;
    const opciones = [
      ...RESERVADAS.map((p) => ({ label: p, type: 'keyword' })),
      ...(simbolos() ?? []).map((s) => ({
        label: s.nombre,
        type: clase[s.clase] ?? 'variable',
        detail: s.detalle || s.tipo || s.clase,
      })),
    ];
    return { from: palabra.from, options: opciones, validFor: /^[\p{L}\p{Nd}_]*$/u };
  };
}

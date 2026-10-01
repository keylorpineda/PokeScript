// Qué es un nombre del programa y de qué tipo, según la tabla de símbolos que
// devolvió la última compilación (ide.resultado.simbolos). Lo usan el cuadro
// del Pokémon (el tipo del nombre bajo el cursor) y el editor (el tipo al
// pasar el mouse).

const CLASES = {
  dato: 'dato',
  parametro: 'parámetro',
  medalla: 'medalla',
  movimiento: 'movimiento',
  especie: 'especie',
  valor: 'valor',
  ficha: 'ficha',
};

// buscar devuelve los símbolos con ese nombre en la lista de un archivo.
export function buscar(lista, nombre) {
  if (!nombre || !Array.isArray(lista)) return [];
  return lista.filter((s) => s.nombre === nombre);
}

// describir arma el texto de un símbolo: su clase y su tipo o su firma.
export function describir(s) {
  const clase = CLASES[s.clase] ?? s.clase;
  switch (s.clase) {
    case 'movimiento':
      return { clase, texto: s.detalle || s.nombre, tipo: s.tipo || '' };
    case 'especie':
      return { clase, texto: s.nombre, tipo: '', detalle: s.detalle || '' };
    case 'ficha':
      return { clase, texto: s.nombre, tipo: '', detalle: s.detalle || '' };
    case 'valor':
      return { clase, texto: s.nombre, tipo: s.tipo || '' };
    default:
      return { clase, texto: s.nombre, tipo: s.tipo || '' };
  }
}

// colorTipo da la variable de color del tipo base: «posible roca» y «equipo
// de roca» se pintan como roca. Las especies y fichas usan el color de los
// nombres de tipo.
export function colorTipo(tipo) {
  const base = /(roca|agua|fuego|planta|electrico)$/.exec(tipo ?? '');
  return base ? `var(--t-${base[1]})` : 'var(--nombre_tipo)';
}

// Plantillas para archivos nuevos: cada una trae un ejemplo corto que
// compila y un nombre sugerido. El objeto es la Pokéball con que se muestra.

export const PLANTILLAS = [
  {
    id: 'vacio',
    titulo: 'En blanco',
    descripcion: 'Sin nada adentro.',
    objeto: 'poke-ball',
    nombre: 'nuevo',
    contenido: '',
  },
  {
    id: 'movimientos',
    titulo: 'Con movimientos',
    descripcion: 'Funciones para usar desde el combate.',
    objeto: 'great-ball',
    nombre: 'movimientos',
    contenido: `// Movimientos: las funciones del proyecto.
// Para usarlos en otro archivo: enseñar doble desde "movimientos.pks"

movimiento roca doble(roca n)
    entregar n * 2
fin

movimiento saludar(planta nombre)
    gritar "¡Hola, ", nombre, "!"
fin
`,
  },
  {
    id: 'tipos',
    titulo: 'Con especies y fichas',
    descripcion: 'Tus propios tipos de datos.',
    objeto: 'ultra-ball',
    nombre: 'tipos',
    contenido: `// Especies y fichas: los tipos propios del proyecto.

especie Estado
    SANO, DORMIDO, ENVENENADO
fin

ficha Pokemon
    planta nombre
    roca   nivel
    Estado estado
fin
`,
  },
  {
    id: 'medallas',
    titulo: 'Con medallas',
    descripcion: 'Valores que nunca cambian.',
    objeto: 'rare-candy',
    nombre: 'constantes',
    contenido: `// Medallas: valores que no cambian nunca. Van en mayúsculas.

medalla roca NIVEL_MAXIMO = 100
medalla agua PRECISION = 0.95
medalla planta REGION = "Kanto"
`,
  },
  {
    id: 'lista',
    titulo: 'Con un recorrido',
    descripcion: 'Un movimiento que suma un equipo.',
    objeto: 'potion',
    nombre: 'colecciones',
    contenido: `// Recorrer un equipo con un movimiento.

movimiento roca sumar_niveles(equipo de roca niveles)
    roca total = 0
    recorrer nivel en niveles
        total = total + nivel
    fin
    entregar total
fin
`,
  },
];

// tipoDeArchivo mira lo que tiene un archivo para mostrarlo en la mochila:
// todos son .pks, pero no es lo mismo el que tiene el combate que uno de
// medallas.
export function tipoDeArchivo(contenido = '') {
  const lineas = contenido
    .split('\n')
    .map((l) => l.replace(/\/\/.*$/, '').trim())
    .filter(Boolean);
  const hay = (re) => lineas.some((l) => re.test(l));
  if (!lineas.length) return { objeto: 'poke-ball', nombre: 'Archivo vacío', clase: 'vacio' };
  if (hay(/^combate\b/))
    return { objeto: 'master-ball', nombre: 'Tiene el combate', clase: 'combate' };
  if (lineas.every((l) => /^(medalla|enseñar)\b/.test(l)))
    return { objeto: 'rare-candy', nombre: 'Medallas', clase: 'medallas' };
  if (hay(/^movimiento\b/))
    return { objeto: 'great-ball', nombre: 'Movimientos', clase: 'movimientos' };
  if (hay(/^(especie|ficha)\b/))
    return { objeto: 'ultra-ball', nombre: 'Especies y fichas', clase: 'tipos' };
  return { objeto: 'poke-ball', nombre: 'PokeScript', clase: 'otro' };
}

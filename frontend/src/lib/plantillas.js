// Plantillas para archivos nuevos: cada una trae un ejemplo corto que
// compila y un nombre sugerido. El objeto es la Pokéball con que se muestra.

export const PLANTILLAS = [
  {
    id: 'vacio',
    titulo: 'Archivo vacío',
    descripcion: 'Una hoja en blanco.',
    objeto: 'poke-ball',
    nombre: 'nuevo',
    contenido: '',
  },
  {
    id: 'movimientos',
    titulo: 'Movimientos',
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
    titulo: 'Especies y fichas',
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
    titulo: 'Medallas',
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
    titulo: 'Equipo y mochila',
    descripcion: 'Un movimiento que recorre una colección.',
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

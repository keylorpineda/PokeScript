// estados.pks · especie, ficha, segun y movimientos
especie Estado
    SANO, ENVENENADO, DORMIDO
fin

ficha Pokemon
    planta nombre
    roca   vida
    Estado estado
fin

movimiento roca pasar_turno(Pokemon p)
    roca dano = 0
    segun p.estado
        SANO        entonces gritar p.nombre, " está en plena forma"
        ENVENENADO  entonces dano = 10
        DORMIDO     entonces gritar p.nombre, " sigue dormido…"
    fin
    entregar p.vida - dano
fin

combate
    Pokemon bulbi = {nombre: "Bulbasaur", vida: 45, estado: ENVENENADO}
    recorrer t de 1 hasta 3
        bulbi.vida = pasar_turno(bulbi)
        gritar "Turno ", t, ": ", bulbi.nombre, " tiene ", bulbi.vida, " PS"
    fin
fin

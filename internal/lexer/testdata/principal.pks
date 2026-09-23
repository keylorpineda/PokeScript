enseñar calcular_dano desde "operaciones.pks"
enseñar Estado, Pokemon desde "tipos.pks"
enseñar VIDA_MAXIMA desde "constantes.pks"

movimiento describir(Estado actual)
    segun actual
        SANO                 entonces gritar "  Puede atacar"
        ENVENENADO           entonces gritar "  Pierde vida cada turno"
        DORMIDO, PARALIZADO  entonces gritar "  Podría no atacar"
    fin
fin

combate
    planta nombre
    capturar(nombre, "¿Cómo se llama tu Pokemon? ")

    Pokemon mio   = {nombre: nombre,   vida: VIDA_MAXIMA, estado: SANO}
    Pokemon rival = {nombre: "Bulbi", vida: VIDA_MAXIMA, estado: SANO}

    gritar "¡", mio.nombre, " entra en combate!"
    describir(mio.estado)

    recorrer turno de 1 hasta 20
        gritar "--- Turno ", turno, " ---"

        roca dano = calcular_dano(40)
        rival.vida = rival.vida - dano
        gritar mio.nombre, " ataca y hace ", dano, " de daño"

        si rival.vida <= 0
            gritar rival.nombre, " se debilitó. ¡Ganaste!"
            huir
        fin

        roca contra = calcular_dano(35)
        mio.vida = mio.vida - contra
        gritar rival.nombre, " contraataca: ", contra, " de daño"
        gritar "Vida de ", mio.nombre, ": ", mio.vida

        si mio.vida <= 0
            gritar mio.nombre, " se debilitó. Perdiste."
            huir
        fin
    fin
fin

// entrenamiento.pks · equipo, mochila y ciclos
combate
    equipo de planta mi_equipo = ["Pikachu", "Charmander", "Squirtle"]
    sumar "Bulbasaur" a mi_equipo
    gritar "Tu equipo tiene ", tamaño(mi_equipo), " Pokémon"

    mochila de planta a roca bayas = {"Aranja": 3, "Zreza": 0, "Meloc": 5}
    recorrer baya, cantidad en bayas
        si cantidad igual 0
            siguiente
        fin
        gritar "  ", baya, " x", cantidad
    fin

    roca nivel = 5
    mientras nivel < 10
        nivel = nivel + 2
    fin
    gritar "Nivel final: ", nivel
fin

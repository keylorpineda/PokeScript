// centro.pks · capturar, contiene y sino si
movimiento electrico es_legendario(planta nombre)
    equipo de planta legendarios = ["Mewtwo", "Lugia", "Rayquaza"]
    entregar legendarios contiene nombre
fin

combate
    planta nombre
    roca nivel
    capturar(nombre, "¿Qué Pokémon atrapaste? ")
    capturar(nivel, "¿De qué nivel? ")

    si es_legendario(nombre)
        gritar "¡Increíble! ", nombre, " es legendario"
    sino si nivel >= 50
        gritar nombre, " ya es todo un veterano"
    sino
        agua progreso = nivel / 100
        gritar nombre, " va al ", redondear(progreso * 100), "% del camino"
    fin
fin

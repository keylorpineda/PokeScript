enseñar NIVEL desde "constantes.pks"

// Daño base con variación al azar del 85% al 100%
movimiento roca calcular_dano(roca poder)
    agua base      = poder * NIVEL / 50
    roca variacion = aleatorio(85, 100)
    agua total     = base * variacion / 100
    entregar redondear(total) + 2
fin

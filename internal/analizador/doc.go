// Package analizador hace las dos pasadas semánticas de la sección 4.
//
// Pasada 1, Recolectar (recoleccion.go): recorre los archivos alcanzables
// de un proyecto.Proyecto y arma la Tabla de símbolos: especies y sus
// valores, fichas con sus campos, medallas de archivo y cabeceras de
// movimientos, todos con el tipo ya resuelto a tipos.Type. Reporta
// nombres duplicados en el proyecto, campos y parámetros repetidos, tipos
// desconocidos o inválidos, y que haya exactamente un combate en el
// archivo principal.
//
// Pasada 2 (validaciones de cuerpos): cada archivo del paquete revisa un
// grupo de reglas y recibe la Tabla ya armada. Lo que ofrece la Tabla:
//
//	t.Buscar(archivo, nombre)       qué es un nombre visible desde un archivo
//	t.ResolverTipo(archivo, tipo)   tipo escrito en el código → tipos.Type
//	t.Movimientos["f"].Params       firma de un movimiento, para las llamadas
//	t.Fichas["P"].Campo("vida")     tipo de un campo, para los accesos .campo
//	t.Valores["SANO"]               especie de un valor, para segun y literales
//	t.Visibles(archivo)             nombres para las sugerencias del asistente
//
// Los datos declarados dentro de los bloques (variables locales, parámetros
// y variables de recorrer) no están en la Tabla: cada validación lleva su
// propio ámbito mientras recorre el cuerpo.
package analizador

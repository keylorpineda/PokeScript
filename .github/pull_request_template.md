<!--
Título del PR: ID de la tarea + descripción en inglés, en imperativo.
Ejemplo: T2.1 Parse expressions with precedence
-->

## Tareas

<!-- IDs de docs/PLAN.md que este PR cierra o avanza. -->

- Cierra: T
- Avanza:

## Qué cambia

<!-- Qué hace el cambio y por qué, en dos o tres frases. -->

## Cómo probarlo

```bash
go test ./internal/...
```

<!-- Pasos extra para probarlo a mano, si aplica. -->

## Capturas

<!-- Obligatorio si cambia la interfaz: captura o GIF corto. Si no, borra esta sección. -->

## Decisiones

<!-- Algo que la especificación no define y se decidió aquí. También va en el registro de docs/PLAN.md. Si no hay, escribe "Ninguna". -->

## Checklist

- [ ] Los commits siguen el formato `PKS type(scope): description`
- [ ] El CI está en verde
- [ ] Hay pruebas (Go) o captura/GIF (interfaz)
- [ ] Los diagnósticos nuevos llevan `Line`, `Col` y `Len` correctos y están en español
- [ ] Si cambió un contrato (`token`, `diag`, `ast`, `ES`), los otros dos lo aprobaron
- [ ] Actualicé la sección **Estado actual** de `docs/PLAN.md` y marqué la tarea con ✅

## Para quien revisa

<!-- Qué conviene mirar con más cuidado. Quien revisa debe poder explicar el cambio en la defensa. -->

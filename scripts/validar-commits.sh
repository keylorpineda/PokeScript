#!/bin/sh
# Valida con commitlint los mensajes de los commits entre DESDE y HASTA.
#
#   sh scripts/validar-commits.sh DESDE HASTA
#
# Los commits de merge se saltan: su título es el del PR ("T2.1 Parse
# expressions…"), que sigue la convención de PR y no la de commits.
set -eu
desde=$1
hasta=$2
fallos=0
for sha in $(git rev-list --no-merges "$desde..$hasta"); do
  if ! git log -1 --format=%B "$sha" | pnpm exec commitlint --verbose; then
    echo "✖ commit $(git log -1 --format='%h %s' "$sha")"
    fallos=$((fallos + 1))
  fi
done
if [ "$fallos" -gt 0 ]; then
  echo "✖ $fallos commit(s) no siguen el formato PKS type(scope): description"
  exit 1
fi
echo "✔ Todos los commits siguen el formato."

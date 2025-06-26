#!/bin/bash

set -euo pipefail

command -v git >/dev/null 2>&1 || { echo "Errore: git non trovato."; exit 1; }
command -v git-flow >/dev/null 2>&1 || { echo "Errore: git-flow non trovato."; exit 1; }

# Parametri
type=${1:-}
version=${2:-}
FILES_TO_REMOVE="tests/ scripts/ monitoring/ RELEASE_TEMPLATE.md NOTES.md BENCHMARKS.md"
SOURCE_BRANCH="${type}/${version}"

# Validazione input
if [[ "$type" != "release" && "$type" != "hotfix" ]]; then
  echo "Errore: <type> deve essere 'release' o 'hotfix'."
  echo "Usage: $0 <release|hotfix> <version> (es. release 1.2.0)"
  exit 1
fi

if [[ -z "$version" ]]; then
  echo "Errore: <version> mancante."
  echo "Usage: $0 <release|hotfix> <version> (es. release 1.2.0)"
  exit 1
fi

# Controlla se il branch esiste
if ! git show-ref --verify --quiet "refs/heads/$SOURCE_BRANCH"; then
  echo "Errore: branch '$SOURCE_BRANCH' non trovato. Crea il branch con: git flow $type start $version"
  exit 1
fi

echo ">>> Switch al branch $SOURCE_BRANCH"
git switch "$SOURCE_BRANCH"

echo ">>> Rimozione file non destinati alla produzione"
git rm -r --ignore-unmatch $FILES_TO_REMOVE || echo "Alcuni file non esistono, continuo..."
git commit -m "Pulizia file non destinati alla produzione"

echo ">>> Pubblicazione del $type $version"
git flow "$type" publish "$version"

echo ">>> Push branch $SOURCE_BRANCH"
git push origin "$SOURCE_BRANCH"

echo ">>> Finish del $type $version"
git flow "$type" finish "$version" --nodevelopmerge -Fp

echo ">>> Switch al branch main"
git switch main

echo ">>> Operazione completata con successo."

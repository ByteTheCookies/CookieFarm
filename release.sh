#!/bin/bash

type=$1       # release o hotfix
version=$2    # es. 1.2.0 o 1.2.1-hotfix

FILES_TO_REMOVE="tests/ scripts/ monitoring/ RELEASE_TEMPLATE.md NOTES.md BENCHMARKS.md"

if [ -z "$type" ] || [ -z "$version" ]; then
  echo "Usage: $0 <release|hotfix> <version> (e.g., release 1.2.0)"
  exit 1
fi


SOURCE_BRANCH="release/$version"
git switch $SOURCE_BRANCH
git rm -r $FILES_TO_REMOVE
git commit -m "Pulizia file non destinati alla produzione"
git flow release publish "$version"
git push
git flow release finish "$version" --nodevelopmerge -Fp

git switch main

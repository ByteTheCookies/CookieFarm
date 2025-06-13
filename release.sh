#!/bin/bash

version=$1

if [ -z "$version" ]; then
  echo "Usage: $0 <version> (e.g., 1.0.0)"
  exit 1
fi

git flow release finish "$version" --nodevelopmerge --nomainmerge -Fp

git checkout -b temp-clean-release main

git rm -r tests/ scripts/ monitoring/ RELEASE_TEMPLATE.md NOTES.md BENCHMARKS.md
git commit -m "Pulizia file non destinati alla produzione"

git checkout main
git merge temp-clean-release

git push origin main
git push origin --tags

git branch -d temp-clean-release

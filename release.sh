#!/bin/bash

type=$1       # release o hotfix
version=$2    # es. 1.2.0 o 1.2.1-hotfix

FILES_TO_REMOVE="tests/ scripts/ monitoring/ RELEASE_TEMPLATE.md NOTES.md BENCHMARKS.md"

if [ -z "$type" ] || [ -z "$version" ]; then
  echo "Usage: $0 <release|hotfix> <version> (e.g., release 1.2.0)"
  exit 1
fi

if [ "$type" == "release" ]; then
  git flow release finish "$version" --nodevelopmerge --nomainmerge -Fp
  SOURCE_BRANCH="release/$version"
elif [ "$type" == "hotfix" ]; then
  git flow hotfix finish "$version" --nomainmerge -Fp
  SOURCE_BRANCH="hotfix/$version"
else
  echo "Invalid type: must be 'release' or 'hotfix'"
  exit 1
fi

TEMP_BRANCH="temp-clean-$type-$version"
git checkout -b "$TEMP_BRANCH" main

git merge --no-commit "$SOURCE_BRANCH"

git rm -r --cached $FILES_TO_REMOVE 2>/dev/null
git commit -m "Pulizia file non destinati alla produzione"

git checkout main
git merge "$TEMP_BRANCH"

git push origin main
git push origin --tags

if [ "$type" == "release" ]; then
  git checkout develop
  git merge "$SOURCE_BRANCH"
  git push origin develop
fi

git branch -d "$TEMP_BRANCH"

## Guide to release a new version of CookieFarm

### 1. Update the version number

- Make sure the version number in `pyproject.toml` is updated to the new version you want to release.
- Make sure the version number in `README.md` is updated to the new version you want to release.

### 2. Prepare branch

- Make sure you are on the `dev` branch and it is up to date with the remote repository.
- Run `git flow release start <version>` to create a new release branch `version` without the v.
- Make eveny necessary changes in the release branch, such as updating the changelog, fixing any bugs, etc.
- Run `git flow release finish <version>` to finish the release and merge the release branch

### 3. Manage the tags

- After finishing the release, a new tag will be created with the format `v<version>`.
- You can check the tags by running `git tag`.
- If you want to push the tags to the remote repository, run `git push --tags

### 4. Checking and Edits

- If all is good, you can push the changes to the remote repository by running `git push origin main`.
- Wait gorelease to create the release on GitHub. This may take a few minutes.
- After that, you can check the release on GitHub and make any necessary edits to the release notes, such as adding more details, fixing typos, etc.
- Make sure all workflows and packages are working correctly after the release. If some workflow not working and is not possible to publish ghrc or pypi, you can do it manually and fix later the problem on workflow

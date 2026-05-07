# CookieFarm Attacker Container

Docker-based attacker environment for testing.
It builds a privileged Debian 13 container with Docker-in-Docker enabled, starts dockerd manually, builds the Linux client, and installs the `cookiefarm` Python package to interact with the project from inside the container.

## What this does

The attacker container is meant to provide a self-contained environment to test the current CookieFarm version up to working tree changes.

When the container starts, it:

1. Launches its own Docker daemon (`dockerd`);
2. Starts the CookieFarm server stack with Docker Compose;
3. Builds the Linux client binary;
4. Exposes the client as `ckc`;
5. Creates a Python virtual environment and installs `cookiefarm`.

## Required `.env`

Create a `.env` file with the values expected by the project.  
At minimum, the project expects the following variables:

- `PORT`
- `PASSWORD`
- `CONFIG_FILE`
- `DEBUG`

## How to build

From this directory:

```sh
just build
```

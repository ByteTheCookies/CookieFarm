#!/bin/sh
set -eu

rm -f /var/run/docker.pid /var/run/docker.sock
export DOCKER_HOST=unix:///var/run/docker.sock
dockerd --host=unix:///var/run/docker.sock --storage-driver=overlay2 2>&1 | tee /var/log/dockerd.log &

for _ in $(seq 1 30); do
    if docker info >/dev/null 2>&1; then
        break
    fi
    sleep 1
done

if ! docker info >/dev/null 2>&1; then
    echo "dockerd failed to start. See /var/log/dockerd.log" >&2
    exit 1
fi

cd CookieFarm/

docker compose -f cookiefarm/compose.build.yml up -d --build

just client-build-linux-prod

exec "$@"

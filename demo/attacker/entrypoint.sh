#!/bin/sh
set -eu

rm -f /var/run/docker.pid /var/run/docker.sock
export DOCKER_HOST=unix:///var/run/docker.sock
nohup dockerd --host=unix:///var/run/docker.sock --storage-driver=overlay2 > /var/log/dockerd.log 2>&1 &

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

export CGO_ENABLED=1
just client-build-linux

ln -P bin/ckc /usr/local/bin/ckc

python3 -m venv .venv
. ./.venv/bin/activate
pip install --upgrade -i https://test.pypi.org/simple cookiefarm


exec "$@"

FROM debian:13-slim

ENV DEBIAN_FRONTEND=noninteractive
ENV DOCKER_TLS_CERTDIR=

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       ca-certificates curl iptables iproute2 uidmap docker-cli docker.io openssh-server \
    && (apt-get install -y --no-install-recommends docker-compose-plugin || true) \
    && if ! docker compose version >/dev/null 2>&1; then \
         arch="$(dpkg --print-architecture)"; \
         case "$arch" in \
           amd64) compose_arch="x86_64" ;; \
           arm64) compose_arch="aarch64" ;; \
           *) echo "Unsupported architecture for compose plugin: $arch" >&2; exit 1 ;; \
         esac; \
         mkdir -p /usr/libexec/docker/cli-plugins; \
         curl -fsSL "https://github.com/docker/compose/releases/latest/download/docker-compose-linux-${compose_arch}" \
           -o /usr/libexec/docker/cli-plugins/docker-compose; \
         chmod +x /usr/libexec/docker/cli-plugins/docker-compose; \
       fi

RUN apt-get install -y --no-install-recommends \
		vim python3 git just golang-go build-essential python3.13-venv

COPY ./entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

COPY ./CookieFarm /CookieFarm
COPY .env /CookieFarm/cookiefarm/.env

# web interface port on local machine
EXPOSE 7331

VOLUME ["/var/lib/docker"]
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

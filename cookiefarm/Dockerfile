FROM golang:1.26.0-alpine AS build

RUN apk add --no-cache alpine-sdk

WORKDIR /app

COPY ./ /app/

ARG VERSION=dev

# server builds
RUN GOOS="linux" GOARCH="amd64" \
  go build -trimpath \
  -ldflags="-s -w -X 'main.Version=${VERSION}'" \
  -o ./bin/cks ./server/main.go

# plugins build
RUN for file in $(find ./pkg/protocols -name '*.go' ! -name 'protocols.go'); do \
 if grep -q '^package main' "$file"; then \
  filename=$(basename "$file"); \
  pluginname="${filename%.go}"; \
  GOOS="linux" GOARCH="amd64" \
    go build -trimpath \
        -ldflags="-s -w -X 'main.Version=${VERSION}'" \
        -buildmode=plugin -o "./pkg/protocols/$pluginname.so" "$file"; \
 else \
  echo "Skipping $file: not a main package"; \
 fi; \
done


FROM oven/bun:1.3.0-alpine AS bun

WORKDIR /app/server/frontend

COPY ./server/frontend/ /app/server/frontend/

RUN bun install --frozen-lockfile --ignore-scripts
RUN bun run build

FROM alpine:3.23.3 AS prod

WORKDIR /app

RUN apk add --no-cache libc6-compat; addgroup -S vincenzino \
    && adduser -S vincenzino -G vincenzino

COPY --from=build /app/bin/cks /app/bin/cks
COPY --from=build /app/server/public /app/server/public
COPY --from=build /app/config.yml /app/config.yml
COPY --from=build /app/pkg/protocols /app/pkg/protocols
COPY --from=bun /app/server/frontend/dist /app/server/frontend/dist
COPY ./run.sh /app/run.sh

RUN touch /app/cookiefarm.db \
    && chmod +x /app/run.sh \
    && chown -R vincenzino:vincenzino /app

USER vincenzino

EXPOSE 8080

ENTRYPOINT ["/bin/sh", "/app/run.sh"]

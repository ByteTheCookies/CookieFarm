FROM golang:1.24.4-alpine AS build

RUN apk add --no-cache alpine-sdk make

WORKDIR /app

COPY go.sum go.mod Makefile ./
RUN go mod download

COPY . .

RUN make server-build-prod
RUN make server-build-plugins-prod

# Runtime stage
FROM alpine:3.20.1 AS prod

WORKDIR /app

RUN apk add --no-cache libc6-compat dos2unix

COPY --from=build /app/bin/cookieserver /app/bin/cookieserver
COPY --from=build /app/internal/server/public /app/internal/server/public
COPY --from=build /app/config.yml /app/config.yml
COPY --from=build /app/pkg/protocols /app/protocols
COPY --from=build /app/internal/server/ui/views /app/internal/server/ui/views

RUN touch ./cookiefarm.db

COPY run.sh run.sh
RUN dos2unix run.sh && chmod +x run.sh

EXPOSE ${PORT}

ENTRYPOINT ["/bin/sh", "/app/run.sh"]

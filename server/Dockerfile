FROM golang:1.24.2-alpine AS build

RUN apk add --no-cache alpine-sdk make

WORKDIR /app

COPY go.sum go.mod Makefile ./
RUN go mod download

COPY . .

RUN make build-prod
RUN make build-plugins-prod

# Runtime stage
FROM alpine:3.20.1 AS prod

WORKDIR /app

RUN apk add --no-cache libc6-compat dos2unix

COPY --from=build /app/bin/cookieserver /app/bin/cookieserver
COPY --from=build /app/public /app/public
COPY --from=build /app/config.yml /app/config.yml
COPY --from=build /app/protocols /app/protocols
COPY --from=build /app/internal/ui/views /app/internal/ui/views

RUN touch ./cookiefarm.db

COPY run.sh run.sh
RUN dos2unix run.sh && chmod +x run.sh

EXPOSE ${PORT}

ENTRYPOINT ["/bin/sh", "/app/run.sh"]

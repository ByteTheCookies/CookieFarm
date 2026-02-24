FROM golang:1.26.0-alpine AS build

RUN apk add --no-cache alpine-sdk make

WORKDIR /app

COPY justfile ./

COPY . .
RUN apk add --no-cache just
RUN just server-build-prod
RUN just server-build-plugins-prod

# Runtime stage
FROM alpine:latest AS prod

WORKDIR /app

RUN apk add --no-cache libc6-compat

COPY --from=build /app/bin/cks /app/bin/cks
COPY --from=build /app/server/public /app/server/public
COPY --from=build /app/config.yml /app/config.yml
COPY --from=build /app/pkg/protocols /app/pkg/protocols
COPY --from=build /app/server/ui/views /app/server/ui/views

RUN touch ./cookiefarm.db

COPY run.sh run.sh
RUN chmod +x run.sh

EXPOSE ${PORT}

ENTRYPOINT ["/bin/sh", "/app/run.sh"]

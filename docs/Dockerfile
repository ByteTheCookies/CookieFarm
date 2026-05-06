# syntax=docker.io/docker/dockerfile:1

# use the official Bun image
# see all versions at https://hub.docker.com/r/oven/bun/tags
FROM oven/bun:alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Install dependencies based on the preferred package manager
COPY ./docs/package.json ./docs/bun.lock* ./docs/package-lock.json* ./docs/source.config.ts ./docs/next.config.* ./
RUN \
    if [ -f bun.lock ]; then bun install --ignore-scripts --frozen-lockfile; \
    else echo "Lockfile not found." && exit 1; \
    fi

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY ./docs .
COPY ./cookiefarm/server/api/docs/swagger.json /cookiefarm/server/api/docs/swagger.json

RUN apk add --no-cache git

ENV NEXT_TELEMETRY_DISABLED=1

RUN \
    if [ -f bun.lock ]; then bun run build; \
    elif [ -f yarn.lock ]; then yarn run build; \
    elif [ -f package-lock.json ]; then npm run build; \
    elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm run build; \
    else echo "Lockfile not found." && exit 1; \
    fi

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs; adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=nextjs:nodejs --chmod=755 /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs --chmod=755 /app/.next/static ./.next/static

USER nextjs

EXPOSE 8882
ENV PORT=8882
ENV HOSTNAME="0.0.0.0"

# server.js is created by next build from the standalone output
CMD ["bun", "server.js"]

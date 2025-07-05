# 1) build stage
FROM oven/bun:1.2.5-alpine AS builder
WORKDIR /app

# install deps e build
COPY package.json bun.lock* ./
RUN bun install --frozen-lockfile
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN bun run build

# 2) runtime stage
FROM oven/bun:1.2.5-alpine AS runner
WORKDIR /app

# copia tutto il necessario
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/.next        ./.next
COPY --from=builder /app/public       ./public
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/bun.lock*    ./

ENV NODE_ENV=production
EXPOSE 3000

# avvia con bun run start (che esegue `next start`)
CMD ["bun", "run", "start"]

FROM node:lts-alpine AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app

FROM base AS deps
COPY --from=frontend package.json pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile 

FROM base AS builder
COPY --from=deps /app/node_modules ./node_modules
COPY --from=frontend . ./
RUN pnpm run build

FROM nginxinc/nginx-unprivileged:1.29.3-alpine-slim AS final
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 8080
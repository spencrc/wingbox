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

FROM openresty/openresty:alpine AS final
RUN apk add --no-cache curl perl
RUN opm install SkyLothar/lua-resty-jwt
# FOR NEXT LINE SEE: https://stackoverflow.com/questions/57734748/changing-nginx-conf-file
COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf 
COPY lua/ /usr/local/openresty/nginx/lua
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 8080
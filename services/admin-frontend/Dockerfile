FROM node:20-alpine AS deps

WORKDIR /app

COPY services/admin-frontend/package.json services/admin-frontend/package-lock.json ./
RUN npm ci

FROM node:20-alpine AS builder

WORKDIR /app

RUN apk add --no-cache curl
RUN curl -sSL https://github.com/bufbuild/buf/releases/download/v1.63.0/buf-Linux-x86_64 -o /usr/local/bin/buf \
  && chmod +x /usr/local/bin/buf

COPY --from=deps /app/node_modules ./node_modules
COPY services/admin-frontend ./services/admin-frontend
COPY buf.yaml buf.lock buf.gen.admin-frontend.yaml ./
COPY proto ./proto

ENV NEXT_TELEMETRY_DISABLED=1

RUN buf generate --template buf.gen.admin-frontend.yaml
RUN npm --prefix services/admin-frontend run build

FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3001

COPY --from=builder /app/services/admin-frontend/package.json ./
COPY --from=builder /app/services/admin-frontend/public ./public
COPY --from=builder /app/services/admin-frontend/.next ./.next
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3001

CMD ["npm", "run", "start"]

FROM node:20-alpine AS deps

WORKDIR /app/services/admin-frontend

COPY services/admin-frontend/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

FROM node:20-alpine AS builder

WORKDIR /app

RUN apk add --no-cache curl
RUN curl -sSL https://github.com/bufbuild/buf/releases/download/v1.63.0/buf-Linux-x86_64 -o /usr/local/bin/buf \
  && chmod +x /usr/local/bin/buf

COPY buf.yaml buf.lock buf.gen.admin-frontend.yaml ./
COPY proto ./proto
COPY services/admin-frontend ./services/admin-frontend
COPY --from=deps /app/services/admin-frontend/node_modules ./services/admin-frontend/node_modules

ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_DISABLE_TURBOPACK=1

RUN buf generate --template buf.gen.admin-frontend.yaml

WORKDIR /app/services/admin-frontend
RUN npm run build

FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3001

COPY --from=builder /app/services/admin-frontend/package.json ./
COPY --from=builder /app/services/admin-frontend/.next ./.next
COPY --from=builder /app/services/admin-frontend/node_modules ./node_modules

EXPOSE 3001

CMD ["npm", "run", "start"]

FROM node:20-alpine AS deps

WORKDIR /app/services/frontend

COPY services/frontend/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

FROM node:20-alpine AS builder

WORKDIR /app

RUN apk add --no-cache curl
RUN curl -sSL https://github.com/bufbuild/buf/releases/download/v1.63.0/buf-Linux-x86_64 -o /usr/local/bin/buf \
  && chmod +x /usr/local/bin/buf

COPY buf.yaml buf.lock buf.gen.frontend.yaml ./
COPY proto ./proto
COPY services/frontend ./services/frontend
COPY --from=deps /app/services/frontend/node_modules ./services/frontend/node_modules

ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_DISABLE_TURBOPACK=1
ENV TURBOPACK=0

RUN buf generate --template buf.gen.frontend.yaml

WORKDIR /app/services/frontend
RUN npm run build

FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000

COPY --from=builder /app/services/frontend/package.json ./
COPY --from=builder /app/services/frontend/.next ./.next
COPY --from=builder /app/services/frontend/node_modules ./node_modules

EXPOSE 3000

CMD ["npm", "run", "start"]

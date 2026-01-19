FROM node:20-alpine AS deps

WORKDIR /app/services/frontend

COPY services/frontend/package.json services/frontend/package-lock.json ./
RUN npm ci

FROM node:20-alpine AS builder

WORKDIR /app/services/frontend

RUN apk add --no-cache curl
RUN curl -sSL https://github.com/bufbuild/buf/releases/download/v1.63.0/buf-Linux-x86_64 -o /usr/local/bin/buf \
  && chmod +x /usr/local/bin/buf

COPY --from=deps /app/services/frontend/node_modules ./node_modules
COPY services/frontend ./
COPY buf.yaml buf.lock buf.gen.frontend.yaml /app/
COPY proto /app/proto

ENV NEXT_TELEMETRY_DISABLED=1

WORKDIR /app
RUN buf generate --template /app/buf.gen.frontend.yaml

WORKDIR /app/services/frontend
RUN npm run build

FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000

COPY --from=builder /app/services/frontend/package.json ./
COPY --from=builder /app/services/frontend/public ./public
COPY --from=builder /app/services/frontend/.next ./.next
COPY --from=builder /app/services/frontend/node_modules ./node_modules

EXPOSE 3000

CMD ["npm", "run", "start"]

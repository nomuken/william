FROM node:20-alpine AS deps

WORKDIR /app

COPY services/frontend/package.json services/frontend/package-lock.json ./
RUN npm ci

FROM node:20-alpine AS builder

WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY services/frontend ./services/frontend
COPY buf.yaml buf.lock buf.gen.frontend.yaml ./
COPY proto ./proto

ENV NEXT_TELEMETRY_DISABLED=1

RUN npx --yes buf generate --template buf.gen.frontend.yaml
RUN npm --prefix services/frontend run build

FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000

COPY --from=builder /app/services/frontend/package.json ./
COPY --from=builder /app/services/frontend/public ./public
COPY --from=builder /app/services/frontend/.next ./.next
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000

CMD ["npm", "run", "start"]

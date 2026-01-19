FROM node:20-alpine AS deps

WORKDIR /app/services/frontend

COPY services/frontend/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

FROM node:20-alpine AS builder

WORKDIR /app/services/frontend

COPY --from=deps /app/services/frontend/node_modules ./node_modules
COPY services/frontend/. ./

ENV NEXT_TELEMETRY_DISABLED=1

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

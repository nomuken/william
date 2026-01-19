FROM node:20-alpine AS deps

WORKDIR /app/services/admin-frontend

COPY services/admin-frontend/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

FROM node:20-alpine AS builder

WORKDIR /app/services/admin-frontend

COPY --from=deps /app/services/admin-frontend/node_modules ./node_modules
COPY services/admin-frontend/. ./

ENV NEXT_TELEMETRY_DISABLED=1

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

FROM node:20-alpine@sha256:fb4cd12c85ee03686f6af5362a0b0d56d50c58a04632e6c0fb8363f609372293 AS builder
WORKDIR /repo

COPY packages/console-shell/package.json packages/console-shell/package-lock.json ./packages/console-shell/
COPY packages/console-shell/src ./packages/console-shell/src
COPY packages/console-shell/assets ./packages/console-shell/assets
COPY packages/console-shell/tsconfig.json ./packages/console-shell/
RUN npm ci --legacy-peer-deps --prefix packages/console-shell

COPY apps/audit-explorer/package.json apps/audit-explorer/package-lock.json ./apps/audit-explorer/
ARG NEXT_PUBLIC_ALERT_CONSOLE_URL=http://localhost:3000
ARG NEXT_PUBLIC_AUDIT_EXPLORER_URL=http://localhost:3002
ARG NEXT_PUBLIC_GRAPH_EXPLORER_URL=http://localhost:3003
ARG NEXT_PUBLIC_SIMULATION_CONSOLE_URL=http://localhost:3004
ARG NEXT_PUBLIC_REPORT_CONSOLE_URL=http://localhost:3005
ENV NEXT_PUBLIC_ALERT_CONSOLE_URL=$NEXT_PUBLIC_ALERT_CONSOLE_URL \
    NEXT_PUBLIC_AUDIT_EXPLORER_URL=$NEXT_PUBLIC_AUDIT_EXPLORER_URL \
    NEXT_PUBLIC_GRAPH_EXPLORER_URL=$NEXT_PUBLIC_GRAPH_EXPLORER_URL \
    NEXT_PUBLIC_SIMULATION_CONSOLE_URL=$NEXT_PUBLIC_SIMULATION_CONSOLE_URL \
    NEXT_PUBLIC_REPORT_CONSOLE_URL=$NEXT_PUBLIC_REPORT_CONSOLE_URL

WORKDIR /repo/apps/audit-explorer
RUN npm ci
COPY apps/audit-explorer/ ./
RUN npm run build

FROM node:20-alpine@sha256:fb4cd12c85ee03686f6af5362a0b0d56d50c58a04632e6c0fb8363f609372293
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /repo/apps/audit-explorer/.next/standalone ./
COPY --from=builder /repo/apps/audit-explorer/.next/static ./apps/audit-explorer/.next/static
COPY --from=builder /repo/apps/audit-explorer/public ./apps/audit-explorer/public
EXPOSE 3000
ENV PORT=3000
CMD ["node", "apps/audit-explorer/server.js"]

FROM node:20-alpine AS build-env
ENV CI=true
COPY frontend/ /app/
WORKDIR /app
RUN corepack enable
RUN corepack prepare pnpm@latest --activate
RUN pnpm install
RUN pnpm run build

FROM node:20-alpine
ENV CI=true
COPY frontend/package.json /app/package.json
COPY frontend/pnpm-lock.yaml /app/pnpm-lock.yaml
RUN corepack enable
RUN corepack prepare pnpm@latest --activate
COPY --from=build-env /app/build /app/build
WORKDIR /app
RUN pnpm install -P
CMD ["pnpm", "start"]

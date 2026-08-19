# Dockerfile：生产构建（node 构建 → node 非 root 运行 nitro 服务）

ARG NODE_IMAGE=node:20-alpine
FROM ${NODE_IMAGE} AS build
ARG NPM_REGISTRY=https://registry.npmmirror.com
RUN npm config set registry ${NPM_REGISTRY}
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY . .
RUN npm run build

FROM ${NODE_IMAGE}
WORKDIR /app
ENV NODE_ENV=production \
    NITRO_HOST=0.0.0.0 \
    NITRO_PORT=3000 \
    TZ=Asia/Shanghai
COPY --from=build /app/.output ./.output
# node 镜像内置非 root 用户 node（uid 1000）
USER node
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]

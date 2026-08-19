# Dockerfile：生产构建（多阶段：golang 编译 → alpine 非 root 运行）

ARG GO_BUILD_IMAGE=golang:1.25-alpine
FROM ${GO_BUILD_IMAGE} AS build
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY} \
    CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/cloud-api ./cmd/server

ARG ALPINE_RUNTIME_IMAGE=alpine:3.20
FROM ${ALPINE_RUNTIME_IMAGE}
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 cloud \
    && adduser -S -u 10001 -G cloud -h /app cloud
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=build /out/cloud-api /usr/local/bin/cloud-api
RUN chown cloud:cloud /app
USER cloud:cloud
EXPOSE 8080
ENTRYPOINT ["cloud-api"]

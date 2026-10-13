FROM golang:1.22-alpine AS build
# 国内网络无法直连 proxy.golang.org，使用 goproxy.cn
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /src
# 先复制依赖清单以利用构建缓存
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/wordtile-server ./cmd/server && \
    go test ./internal/merge/...

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/wordtile-server /usr/local/bin/wordtile-server
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["wordtile-server"]

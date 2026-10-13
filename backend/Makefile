.PHONY: dev tidy test lint docker-build

dev: ## 本地启动（需要 PostgreSQL；默认连接 postgres://wordtile:wordtile@localhost:5432/wordtile）
	go run ./cmd/server

tidy: ## 解析依赖并生成 go.sum
	go mod tidy

test: ## 运行单元测试
	go test ./...

lint: ## golangci-lint（需已安装）
	golangci-lint run ./...

docker-build:
	docker build -t wordtile-backend ./backend

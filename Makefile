.PHONY: run test lint build race up down logs clean

run: ## roda o servidor localmente
	cd server && go run ./src

test: ## roda todos os testes
	cd server && go test ./...

test-race: ## roda todos os testes com race detector
	cd server && go test -race ./...

build: ## compila o binário
	cd server && go build ./...

lint: ## roda o golangci-lint
	cd server && golangci-lint run

fmt: ## formata o código
	cd server && gofmt -w .

up: ## sobe a stack de produção
	docker compose -f server/docker-compose.prod.yml up -d --build

down: ## derruba a stack de produção
	docker compose -f server/docker-compose.prod.yml down

logs: ## logs da stack de produção
	docker compose -f server/docker-compose.prod.yml logs -f

clean: ## limpa artefatos de build
	cd server && go clean -cache -testcache
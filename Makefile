PREFIX?=$(shell pwd)

SHELL:=/bin/bash

BRANCH?=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
SHA:=$(shell git rev-parse --short HEAD 2>/dev/null)

.PHONY: help
help: ## You are looking at it
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## Run locally
	go run ./cmd

.PHONY: logs
logs: ## Tails the logs for the services running in docker compose
	docker-compose logs -f

.PHONY: up
up: ## Run the binary and all backing-services in docker-compose Note: you will need to set environment variables. See docker-compose.yml
	docker-compose up -d --build

.PHONY: down
down: ## Tears down docker compose services, volumes, and orphaned containers
	docker-compose down --remove-orphans

.PHONY: redis
redis: ## Runs only the redis service in docker-compose
	@docker-compose up -d redis && \
	echo "RedisInsight:\thttp://localhost:8001"

.PHONY: redis-dump
redis-dump: redis ## Dump the current state of Redis to a local file
	@docker exec -it knbn-redis-1 sh -c "redis-cli SAVE" && \
	docker cp knbn-redis-1:/data/dump.rdb ./pkg/store/dump.rdb

.PHONY: redis-load
redis-load: redis ## Load our example Redis db into the running container
	@docker cp ./pkg/store/dump.rdb knbn-redis-1:/data/dump.rdb && \
	docker compose stop redis && \
	docker compose start redis


.PHONY: tools
tools: ## Installs development tools
	@go install github.com/a-h/templ/cmd/templ@latest

.PHONY: generate
generate: tools ## Generates templs
	templ generate
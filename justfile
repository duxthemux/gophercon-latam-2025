#!/usr/bin/env just --justfile

install-preqres:
    ollama pull gemma3
    ollama pull nomic-embed-text
    go mod tidy
    go get ./...

run:
    - ollama run gemma3 --keepalive=180m &
    go run ./cmd/api

up: ollama-up compose-up

down: ollama-down compose-down

ollama-up:
    - ollama run gemma3 --keepalive=180m &

ollama-down:
    - ollama stop gemma3

# Initiates docker compose
compose-up:
    #!/usr/bin/env sh
    cd deploy
    docker compose -f ./docker-compose.yaml -p gophercon up -d

# Removes docker stack
compose-down:
    #!/usr/bin/env sh
    cd deploy
    docker compose -f ./docker-compose.yaml -p gophercon down --remove-orphans
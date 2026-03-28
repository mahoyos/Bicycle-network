.PHONY: build run test coverage docker-up docker-down

build:
	go build -o bin/rental-service ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -v -cover

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down -v

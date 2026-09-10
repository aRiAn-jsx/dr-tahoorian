.PHONY: run dev build test migrate-up migrate-down clean

run:
	go run cmd/server/main.go

dev:
	air

build:
	go build -o bin/server cmd/server/main.go

test:
	go test ./...

migrate-up:
	goose -dir migrations sqlite3 tahoorian.db up

migrate-down:
	goose -dir migrations sqlite3 tahoorian.db down

clean:
	rm -rf bin/

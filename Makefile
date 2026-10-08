.PHONY: all build run test clean migrate-up

APP_NAME=bin/server
MAIN_SRC=cmd/server/main.go

all: build

build:
	@echo "Building Go binary..."
	@mkdir -p bin
	go build -o $(APP_NAME) $(MAIN_SRC)

run:
	@echo "Running Go server..."
	go run $(MAIN_SRC)

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning binaries..."
	rm -rf bin/

migrate-up:
	@echo "Database migrations are auto-applied on startup."

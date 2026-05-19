.PHONY: api test build run migrate

api:
	go run ./cmd/apigen/

test:
	go test ./...

build:
	go build -o artline-backend ./cmd/server/

run:
	go run ./cmd/server/

migrate:
	goose -dir db/migrations postgres "$(DATABASE_URL)" up

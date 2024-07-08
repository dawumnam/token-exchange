build: 
	@go build -o bin/token-trader cmd/main.go

test: 
	@go test -v ./...

run: build
	@bin/token-trader

migrate:
	@migrate create -ext sql -dir cmd/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down 
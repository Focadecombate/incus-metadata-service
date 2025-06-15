run-api:
	go run ./metadata-service/cmd/server

lint:
	golangci-lint run ./...

generate-db:
	sqlc generate
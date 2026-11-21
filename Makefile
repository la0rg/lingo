.PHONY: up run css css-watch air lint test test-integration cover mocks vuln

up:
	docker compose up -d minio dynamodb

run: up css
	go run main.go

css:
	npx @tailwindcss/cli -i web/src/css/input.css -o web/dist/styles.css --minify

css-watch:
	npx @tailwindcss/cli -i web/src/css/input.css -o web/dist/styles.css --watch

air: up
	air

lint:
	golangci-lint run ./...

test:
	go test -race -cover ./...

# Requires DynamoDB Local: `make up`
test-integration:
	go test -race -tags integration ./internal/repository/...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

mocks:
	mockery

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

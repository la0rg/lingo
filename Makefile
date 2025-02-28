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
	golangci-lint run
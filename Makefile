up:
	docker-compose up -d

run: up css
	go run main.go

css:
	npx tailwindcss -c web/tailwind.config.js -i web/src/css/input.css \
	-o web/dist/styles.css --minify

css-watch:
	npx tailwindcss -c web/tailwind.config.js -i web/src/css/input.css \
	-o web/dist/styles.css --minify --watch

air:
	air --build.cmd "go build -o .bin/lingo cmd/server/main.go" --build.bin ".bin/lingo"
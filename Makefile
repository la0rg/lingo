up:
	docker-compose up -d

run: up css
	go run main.go

css:
	tailwindcss -c web/tailwind.config.js -i web/src/css/input.css \
	-o web/dist/styles.css --minify
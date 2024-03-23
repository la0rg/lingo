run:
	docker-compose up -d

local: run
	AWS_REGION="us-east-1" \
	REGION="eu-west-1" \
	AWS_ACCESS_KEY_ID="minioadmin" \
	AWS_SECRET_ACCESS_KEY="minioadmin" go run main.go
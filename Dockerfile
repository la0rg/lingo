FROM golang:1.27.4 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-X 'github.com/la0rg/lingo/internal/config.buildDate=$(date)' \
    -X 'github.com/la0rg/lingo/internal/config.gitCommit=$(git rev-parse HEAD)' \
    -X 'github.com/la0rg/lingo/internal/config.gitVersion=$(git describe --tags --abbrev=0 | tr -d '\n')'" \
    -o lingo main.go

FROM alpine:3
RUN apk add --no-cache ca-certificates && adduser -D -H -u 10001 lingo
WORKDIR /app
COPY --from=build /app/lingo .
USER lingo
ENTRYPOINT ["/app/lingo"]

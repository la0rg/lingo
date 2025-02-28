FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-X 'github.com/la0rg/lingo/internal/config.buildDate=$(date)' \
    -X 'github.com/la0rg/lingo/internal/config.gitCommit=$(git rev-parse HEAD)' \
    -X 'github.com/la0rg/lingo/internal/config.gitVersion=$(git describe --tags --abbrev=0 | tr -d '\n')'" \
    -o lingo main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/lingo .
ENTRYPOINT ["/app/lingo"]
FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o lingo .

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/lingo .
ENTRYPOINT ["/app/lingo"]
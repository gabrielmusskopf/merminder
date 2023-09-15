FROM golang:1.20.3

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/merminder ./cmd/merminder.go

CMD ["./bin/merminder"]

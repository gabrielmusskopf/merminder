FROM golang:1.20.3

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN rm .merminder.yml || echo "not found"
RUN rm .merminder.yaml || echo "not found"

RUN CGO_ENABLED=0 GOOS=linux go build -o ./merminder

CMD ["./merminder"]

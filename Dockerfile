FROM golang:1.20.3

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN mkdir build
RUN CGO_ENABLED=0 GOOS=linux go build -o ./merminder

CMD ["./merminder"]

dev:
	go run ./cmd/merminder.go

build: clean deps
	CGO_ENABLED=0 go build -v -o ./bin/merminder ./cmd/merminder.go

deps:
	go mod tidy

clean:
	rm -rf ./bin

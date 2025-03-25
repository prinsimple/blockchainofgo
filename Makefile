build:
	go build -o ./bin/goblock


run: build
	./bin/goblock

test:
	go test -v ./...



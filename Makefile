envbin: main.go
	go build

envbin-linux: main.go
	GOOS=linux go build -o envbin-linux

run: envbin
	go run main.go

build-docker: envbin-linux
	docker build . --file Dockerfile.ubuntu -t envbin

run-docker: build-docker
	docker run --rm --name envbin -d -p 8080 envbin

clean:
	rm -f envbin envbin-linux

.PHONY: run build-docker run-docker build-linux clean

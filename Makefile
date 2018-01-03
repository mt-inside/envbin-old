envbin: main.go
	go build

run:
	go run main.go

build-docker:
	docker build . -t envbin

build-docker-ubuntu:
	docker build . --file Dockerfile.ubuntu -t envbin-ubuntu

run-docker: build-docker
	docker run --rm --name envbin -d -p 8080 envbin

run-docker-ubuntu: build-docker-ubuntu
	docker run --rm --name envbin -d -p 8080 envbin-ubuntu

clean:
	rm -f envbin

.PHONY: run build-docker build-docker-ubuntu run-docker run-docker-ubuntu clean

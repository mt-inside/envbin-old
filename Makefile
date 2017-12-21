envbin: main.go
	go build

run: envbin
	go run main.go

build-docker: envbin
	docker build . -t envbin

run-docker: build-docker
	docker run --rm --name envbin -d -p 8080 envbin

build-k8s:
	GOOS=linux go build


.PHONY: run build-docker run-docker build-k8s

FROM golang:1.9 as build
WORKDIR /go/src/envbin
COPY . .
RUN go-wrapper download # go get -d -v ./...
ENV CGO_ENABLED=0
RUN go-wrapper install # go install -v ./...

FROM scratch
COPY --from=build /go/bin/envbin /
COPY --from=build /go/src/envbin/main.html /

EXPOSE 8080
CMD ["/envbin"]

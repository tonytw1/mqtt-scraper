FROM golang:1.15

WORKDIR /go/src/app
COPY . .

RUN go get -dt -v ./...
RUN go install -v ./...
run go test -v
run go build -v

CMD ["app"]

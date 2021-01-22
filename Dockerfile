FROM golang:1.15

WORKDIR /go/src/app
COPY . .

RUN go get -dt -v ./...
RUN go install -v ./...
RUN go test -v
RUN go build -v

CMD ["app"]

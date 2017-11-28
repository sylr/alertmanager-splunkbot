FROM golang:1.9.2-alpine

ADD . $GOPATH/src/github.com/sylr/alertmanager-splunkbot
WORKDIR $GOPATH/src/github.com/sylr/alertmanager-splunkbot

RUN apk update && apk upgrade && apk add --no-cache git
RUN go get ./...
RUN go build
RUN go install

ENTRYPOINT ["/go/bin/alertmanager-splunkbot"]

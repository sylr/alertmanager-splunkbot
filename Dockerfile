FROM golang:1.9.2

ADD . $GOPATH/src/github.com/sylr/alertmanager-splunkbot

WORKDIR $GOPATH/src/github.com/sylr/alertmanager-splunkbot
RUN go get ./...
RUN go build
RUN go install

ENTRYPOINT ["/go/bin/alertmanager-splunkbot"]

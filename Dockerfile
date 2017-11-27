FROM golang:1.9.2

RUN go get github.com/sylr/alertmanager-splunkbot

ENTRYPOINT ["/go/bin/alertmanager-splunkbot"]

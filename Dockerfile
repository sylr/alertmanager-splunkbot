FROM golang:1.9.2-alpine3.6

ADD . $GOPATH/src/github.com/sylr/alertmanager-splunkbot
WORKDIR $GOPATH/src/github.com/sylr/alertmanager-splunkbot

RUN apk update && apk upgrade && apk add --no-cache git

RUN go get ./...
RUN go build
RUN go install

# -----------------------------------------------------------------------------

FROM alpine:3.6

WORKDIR /bin
RUN apk --no-cache add ca-certificates
COPY --from=0 "/go/bin/alertmanager-splunkbot" .

ENTRYPOINT ["/bin/alertmanager-splunkbot"]

FROM golang:1.10 AS builder
WORKDIR /go/src/github.com/pusher/bridge-troll
COPY . .
RUN go get -u github.com/golang/dep/cmd/dep \
 && dep ensure -v \
 && cd cmd \
 && env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bridgetroll

FROM scratch
COPY --from=builder /go/src/github.com/pusher/bridge-troll/cmd/bridgetroll /bin/bridgetroll

ENTRYPOINT ["/bin/bridgetroll"]

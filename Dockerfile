# Build the manager binary
FROM golang:1.12-alpine as builder

# Copy in the go src
WORKDIR /go/src/github.com/fbsb/pingdom-operator
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/fbsb/pingdom-operator/cmd/manager

# Copy the controller-manager into a thin image
FROM scratch

WORKDIR /

COPY --from=builder /go/src/github.com/fbsb/pingdom-operator/manager /bin/manager
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 1000

ENTRYPOINT ["/bin/manager"]

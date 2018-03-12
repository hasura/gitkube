# TODO: change to a slimmer and latest version
FROM golang:1.8.5-jessie as builder

# setup the working directory
WORKDIR /go/src/github.com/hasura/gitkube

# copy source code
COPY vendor vendor
COPY pkg pkg
COPY cmd cmd

# build the source
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/gitkube cmd/gitkube-controller/main.go

# use a minimal alpine image
FROM alpine:3.7
WORKDIR /bin
# copy the binary from builder
COPY --from=builder /bin/gitkube /bin/gitkube
# run the binary
CMD ["/bin/gitkube"]

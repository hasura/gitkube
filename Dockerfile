FROM golang:1.11.1-alpine as builder

# setup the working directory
WORKDIR /go/src/github.com/hasura/gitkube

# copy source code
COPY vendor vendor
COPY pkg pkg
COPY cmd cmd

# build the source
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/gitkube-controller cmd/gitkube-controller/main.go

# use a minimal alpine image
FROM alpine:3.7
WORKDIR /bin
# copy the binary from builder
COPY --from=builder /bin/gitkube-controller /bin/gitkube-controller
# run the binary
CMD ["/bin/gitkube-controller"]

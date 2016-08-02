FROM golang

# Copy the local package files to the container's workspace.

ADD ./go /go

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go install github.com/theorangechefco/cts/identity

ENTRYPOINT /go/bin/identity --port 50051 --redis_host 127.0.0.1:6379 --redis_pass abc

EXPOSE 50051

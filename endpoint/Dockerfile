FROM golang

# Copy the local package files to the container's workspace.

ADD ./go /go

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go install github.com/theorangechefco/cts/endpoint

ENTRYPOINT /go/bin/endpoint --port 50051 --recipe_server_addr 127.0.0.1:50052 --identity_server_addr 127.0.0.1:50053 --profile_server_addr 127.0.0.1:500534
EXPOSE 50051

FROM golang

# Copy the local package files to the container's workspace.

ADD ./go /go

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go install github.com/theorangechefco/cts/event

ENTRYPOINT /go/bin/event --port 50051  --identity_server_addr 127.0.0.1:50051 --cassandra_server_addr 127.0.0.1 --cassandra_user eventsrv --cassandra_pass abc

EXPOSE 50051

FROM golang

# Copy the local package files to the container's workspace.

ADD ./go /go
# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go install github.com/theorangechefco/cts/profile

ENTRYPOINT /go/bin/profile --port 50051 --db_host 127.0.0.1 --db_port 3306 --db_user admin --db_pass admin --db_name profile
EXPOSE 50051

FROM golang

# Copy the local package files to the container's workspace.

ADD ./go /go

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go install github.com/theorangechefco/cts/recipestore

ENTRYPOINT /go/bin/recipestore --port 50051 --mongo_host 127.0.0.1:27017 --mongo_user recipestoresrv --mongo_pass abc

EXPOSE 50051

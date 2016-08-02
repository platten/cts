#!/bin/bash

#
# go_script.sh
# Orange Chef Countertop Server Go development utility script
#
#  Created by Paul Pietkiewicz on 7/6/15.
# Copyright (c) 2015 The Orange Chef Company. All rights reserved.
#

# TODO:
# * Add Dockerfile automatic file generation


DESC="Orange Chef Countertop Server Go development utility script"
VERSION="0.3"
DATE="9/23/2015"


#
# Environment variables
#
GRPCPORT="50051"


#
# Functions
#

usage ()
{
  echo "${DESC}, version ${VERSION}"
  echo "Last updated on ${DATE}"
  echo "$0 [--help | --shell | --run | --deploy]"
}

env ()
{
  echo "Setting temp environment variables..."

  export PREFIX="$(pwd)"
  export ORIG_GOPATH="${PREFIX}/go"
  export GOBIN="${ORIG_GOPATH}/bin"
  export PATH="${PREFIX}/src/go/bin:${GOBIN}:/usr/bin:${PATH}"
  export SHARED_LIBS_HOME="$(dirname "${PREFIX}")/go-shared-libs"
  export GO_PROTOS="$(dirname "${PREFIX}")/go-protos"
  GO_SRC_PROJECT_PREFIX="github.com/theorangechefco/cts"
  export PROJECT_TARGET="${ORIG_GOPATH}/src/${GO_SRC_PROJECT_PREFIX}"
  export GO_SRC_SHARED_LIBS="${PROJECT_TARGET}/go-shared-libs"
  export GO_SRC_PROTOS="${PROJECT_TARGET}/go-protos"
  export APP_TARGET="${PROJECT_TARGET}/$(basename "${PREFIX}")"
  export GOPATH="${ORIG_GOPATH}:${GO_SRC_SHARED_LIBS}/thirdparty"
}


link ()
{
  echo "Linking directories..."
  mkdir -p "${ORIG_GOPATH}/src" "${ORIG_GOPATH}/bin" "${PROJECT_TARGET}"
  ln -snfv "${PREFIX}/src" "${APP_TARGET}"
  ln -snfv "${SHARED_LIBS_HOME}" "${GO_SRC_SHARED_LIBS}"
  ln -snfv "${GO_PROTOS}" "${GO_SRC_PROTOS}"
}

copy ()
{
  echo "Copying project source code prior to container build..."
  mkdir -p "${ORIG_GOPATH}/src" "${ORIG_GOPATH}/bin" "${APP_TARGET}" "${GO_SRC_PROTOS}" "${GO_SRC_SHARED_LIBS}"
  # mkdir -p "${GOPATH}/src" "${GOPATH}/bin" "${APP_TARGET}" "${GO_SRC_SHARED_LIBS}"
  cp -Rv "${PREFIX}"/src/* "${APP_TARGET}"
  cp -Rv "${SHARED_LIBS_HOME}"/* "${GO_SRC_SHARED_LIBS}"
  cp -Rv "${GO_PROTOS}"/* "${GO_SRC_PROTOS}"
}

cleanup ()
{
  rm -rf "${ORIG_GOPATH}"/*
}

shell ()
{
    env
    link
    echo "Starting temp shell..."
    cd "${APP_TARGET}"
    PS1="Temp Go shell \w$ " INSHELL="true" /bin/bash --noprofile --norc
    echo "Exited temp shell."
    cleanup
    exit 0
}

docker_restart ()
{
  if [ "$(boot2docker status)" == 'running' ]; then
    echo "Stopping boot2docker"
    boot2docker down
    if [ "$(boot2docker status)" == 'running' ]; then
      echo "Docker / boot2docker still running, exiting."
      exit 10
    fi
  fi

  echo "Checking if port $GRPCPORT is free"
  if [ "$(lsof -nP -iTCP:"$GRPCPORT" -sTCP:LISTEN)" ]; then
    echo "Port $GRPCPORT not available, exiting."
    exit 10
  fi

  echo "Bringing boot2docker back up..."
  boot2docker up

  echo "Checking if boot2docker is up"
  if [ "$(boot2docker status)" == 'running' ]; then
    echo "boot2docker running"
  else
    echo "boot2docker did not start"
    exit 12
  fi

  echo "Cleaning stale Docker containers..."
  docker_cleanup
}

docker_cleanup ()
{
  for container in $(docker ps -aq); do
    docker rm "$container"
  done
}

run ()
{
  # TODO: check if application is currently running
  if [ ! -e "app.yaml" ]; then
    echo "No app.yaml, cannot run."
    exit 6
  fi
  env
  cleanup
  copy

  docker_restart

  echo "Starting application..."
  gcloud --verbosity info preview app run ./app.yaml
  echo "Closed application."
  cleanup
  exit 0
}

manual_build ()
{
  # TODO: check if application is currently running
  if [ ! -e "app.yaml" ]; then
    echo "No app.yaml, cannot run."
    exit 6
  fi
  env
  cleanup
  copy

  #docker_restart
  PS1="Manual docker build shell \w$ " INSHELL="true" /bin/bash --noprofile --norc
  echo "Exited temp shell."
  cleanup
  exit 0
}

deploy ()
{
  # TODO: check if application is currently running
  if [ ! -e "app.yaml" ]; then
    echo "No app.yaml, cannot deploy."
    exit 6
  fi
  env
  cleanup
  docker_restart
  copy
  echo "Deploying application..."
  gcloud --verbosity info preview app deploy ./app.yaml  --server preview.appengine.google.com
  cleanup
  exit 0
}


# Check base condtions
if [ ! -e Dockerfile ]; then
  echo "Not in current valid module directory, Dockerfile not present!"
  exit 3
fi

if ! (( $(grep -c "FROM golang" Dockerfile) )); then
#if grep --quiet "FROM golang" Dockerfile; then
  echo "Dockerfile not using Go base image, not in Go module."
  exit 4
fi

case "$1" in
    --help)
      usage
      exit 0
    ;;
    --shell)
      shell
    ;;
    --run)
      run
    ;;
    --deploy)
      deploy
    ;;
    --manual)
      manual_build
    ;;
    *)
      usage
      exit 1
    ;;
esac

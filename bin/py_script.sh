#!/bin/bash

#
# py_script.sh
# Orange Chef Countertop Server Go development utility script
#
#  Created by Paul Pietkiewicz on 7/27/15.
# Copyright (c) 2015 The Orange Chef Company. All rights reserved.
#

# TODO:
# * Add Dockerfile automatic file generation


DESC="Orange Chef Countertop Server Python development utility script"
VERSION="0.1"
DATE="7/27/2015"


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
  export SHARED_LIBS_HOME="$(dirname "${PREFIX}")/py-shared-libs"
  export PY_PROTOS=""$(dirname "${PREFIX}")/py-protos"
  export PYTHONPATH="${PREFIX}/src:${SHARED_LIBS_HOME}:${PY_PROTOS}"
  export APP_TARGET="${PREFIX}/py"
}

copy ()
{
  echo "Copying project source code prior to container build..."
  cp -Rv "${PREFIX}"/src/* "${APP_TARGET}"
  cp -Rv "${SHARED_LIBS_HOME}"/* "${APP_TARGET}"
}

cleanup ()
{
  rm -rf "${APP_TARGET}/*"
}

shell ()
{
    env
    cd src
    PS1="Temp Python shell \w$ " INSHELL="true" /bin/bash --noprofile --norc
    echo "Exited temp shell."
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

if ! (( $(grep -c "PYTHONPATH" Dockerfile) )); then
  echo "Dockerfile not using Python base image, not in Python module."
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
    *)
      usage
      exit 1
    ;;
esac

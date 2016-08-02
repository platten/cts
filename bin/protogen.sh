#!/bin/bash

if [ ! -e .gitmodules ]; then
  echo "Not in cts project, .gitmodules not present!"
  exit 1
fi

if [ -d "/tmp/gogen" ]; then
  rm -rf /tmp/gogen/*
else
  mkdir /tmp/gogen
fi
if ! $(protoc -I protos/  protos/*.proto  --go_out=plugins=grpc:/tmp/gogen); then
    echo "Go Proto compilation failed, exiting!"
    exit 2
fi

if [ -d "/tmp/pygen" ]; then
  rm -rf /tmp/pygen/*
else
  mkdir /tmp/pygen
fi
if ! $(protoc -I protos/ --python_out=py-protos --grpc_out=/tmp/pygen --plugin=protoc-gen-grpc=`which grpc_python_plugin` protos/*.proto); then
    echo "Python Proto compilation failed, exiting!"
    exit 2
fi

if [ -d "/tmp/objgen" ]; then
  rm -rf /tmp/objgen/*
else
  mkdir /tmp/objgen
fi

if ! $(protoc -I protos --objc_out=objc-protos --objcgrpc_out=/tmp/objgen protos/*.proto); then
    echo "ObjectiveC Proto compilation failed, exiting!"
    exit 2
fi

echo "All protocol buffers compiled successfully"
rm -rf objc-protos/*
rm -f py-protos/*
rm -rf go-protos/*

mv /tmp/gogen/* go-protos/
mv /tmp/pygen/* py-protos/
mv /tmp/objgen/* objc-protos/

rmdir /tmp/gogen /tmp/pygen /tmp/objgen

echo "Done!"

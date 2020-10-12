#!/bin/bash
set -e
cd `dirname "$0"`

mkdir -p local

if ! command -v go &> /dev/null
then
  echo "Could not find go. Make sure it is installed and in your path."
  exit 1
fi

echo "updating bpak..."
go install ./cmd/bpak

if ! command -v bpak &> /dev/null
then
  echo "Could not find bpak after installing it. Make sure your \$GOPATH/bin is in your path."
  exit 1
fi

echo "generating..."
go generate

echo "testing..."
go test ./...

echo "building example..."
exename=example
if [[ "$OSTYPE" == "cygwin" ]]; then
  # POSIX compatibility layer and Linux environment emulation for Windows
  exename=example.exe
elif [[ "$OSTYPE" == "msys" ]]; then
  # Lightweight shell and GNU utilities compiled for Windows (part of MinGW)
  exename=example.exe
fi
go build -o local/$exename ./cmd/example

echo "success!"

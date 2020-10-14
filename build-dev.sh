#!/bin/bash
set -e
cd `dirname "$0"`

mkdir -p local

if ! command -v go &> /dev/null
then
  echo "could not find go. Make sure it is installed and in your path."
  exit 1
fi

echo "Updating bpak..."
go install ./cmd/bpak

if ! command -v bpak &> /dev/null
then
  echo "could not find bpak after installing it. Make sure your \$GOPATH/bin is in your path."
  exit 1
fi

echo "Generating..."
go generate

echo "Testing..."
go test ./...

echo "Building example..."
exename=example
if [[ "$OSTYPE" == "cygwin" ]]; then
  exename=example.exe
elif [[ "$OSTYPE" == "msys" ]]; then
  exename=example.exe
fi
go build -o local/$exename ./cmd/example

echo "Done"

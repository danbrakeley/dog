#!/bin/bash -e
cd `dirname "$0"`

mkdir -p local

if ! command -v go &> /dev/null
then
  echo "could not find go. Make sure it is installed and in your path."
  exit 1
fi

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
TARGET_EXE=example

if [[ "${GOOS}" == "windows" ]]; then
  TARGET_EXE=${TARGET_EXE}.exe
fi

echo "Testing..."
go test ./...

echo "Building example..."
go build -o local/$TARGET_EXE ./cmd/example

echo "Done"

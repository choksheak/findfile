#/bin/bash

if [ ! `which go` ]; then
  echo Please install the go executable.
  exit 1
fi

echo Running go fmt
go fmt ff/...

if [ `which golint` ]; then
  echo Running golint
  golint github.com/choksheak/findfile/ff
elif [ `which golangci-lint` ]; then
  echo Running golangci-lint
  golangci-lint run ff
else
  echo "Skipping golint (install: Linux: go install golang.org/x/lint/golint@latest OR Mac: brew install golangci-lint)"
fi

echo Building ff executable
go install github.com/choksheak/findfile/ff@latest

cd ff; go build

echo "Compiled ff executable in ./ff directory."

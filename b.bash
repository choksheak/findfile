#/bin/bash

# Exit when any command fails.
set -e

echo Running go fmt
go fmt github.com/choksheak/findfile/ff

if [ `which golint1` ]; then
  echo Running golint
  golint github.com/choksheak/findfile/ff
else
  echo "Skipping golint (install: go get -u github.com/golang/lint/golint)"
fi

echo Building ff executable
go install github.com/choksheak/findfile/ff@latest

cd ff; go build

echo "Compiled ff executable in ./ff directory."

#/bin/bash

echo Running go fmt
go fmt github.com\choksheak\findfile\ff

if [ !`which golint` ]; then
  go get -u github.com/golang/lint/golint
fi

echo Running golint
golint github.com\choksheak\findfile\ff

echo Building ff executable
go install github.com\choksheak\findfile\ff

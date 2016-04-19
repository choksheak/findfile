#/bin/bash

echo Running go fmt
go fmt

if [ !`which golint` ]; then
  go get -u github.com/golang/lint/golint
fi

echo Running golint
golint github.com\choksheak\findfile
golint github.com\choksheak\findfile\ff

echo Running go build
go build github.com\choksheak\findfile

echo Creating ff.exe
go install github.com\choksheak\findfile\ff

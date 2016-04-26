@echo off

echo Running go fmt
go fmt github.com\choksheak\findfile\ff

where golint >nul 2>&1
if %ERRORLEVEL% EQU 0 goto do_golint
    echo Skipping golint (install: go get -u github.com/golang/lint/golint)
goto after_golint
:do_golint
    echo Running golint
    golint github.com\choksheak\findfile\ff
:after_golint

echo Building ff executable
go install github.com\choksheak\findfile\ff

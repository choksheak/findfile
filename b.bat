@echo Running go fmt
@go fmt

@REM Install golint
@REM go get -u github.com/golang/lint/golint

@echo Running golint
@golint github.com\choksheak\findfile
@golint github.com\choksheak\findfile\ff

@echo Running go build
@go build github.com\choksheak\findfile

@echo Creating ff.exe
@go install github.com\choksheak\findfile\ff

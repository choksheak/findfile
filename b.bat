@echo Running go fmt
@go fmt github.com\choksheak\findfile\ff

@REM Install golint
@REM go get -u github.com/golang/lint/golint

@echo Running golint
@golint github.com\choksheak\findfile\ff

@echo Building ff executable
@go install github.com\choksheak\findfile\ff

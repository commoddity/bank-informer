build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/main.exe main.go
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/main main.go
build-mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/main main.go




# This target install pre-commit to the repo and should be run only once, after cloning the repo for the first time.
init-pre-commit:
	wget https://github.com/pre-commit/pre-commit/releases/download/v2.20.0/pre-commit-2.20.0.pyz;
	python3 pre-commit-2.20.0.pyz install;
	go install golang.org/x/tools/cmd/goimports@v0.6.0;
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.0;
	go install -v github.com/go-critic/go-critic/cmd/gocritic@v0.6.5;
	python3 pre-commit-2.20.0.pyz run --all-files;
	rm pre-commit-2.20.0.pyz.*;

version := $(shell git describe --tags --always)
build_time := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	CGO_ENABLED=0 go build -ldflags="-s -w"

build-use-env:
	CGO_ENABLED=0 go build -ldflags="-s -w -X 'main.Version=$(version)' -X 'main.BuildTime=$(build_time)'" -o build/schema-diff-${GOOS}-${GOARCH}

all:
	make GOARCH=amd64 GOOS=linux build-use-env
	make GOARCH=arm64 GOOS=linux build-use-env
	make GOARCH=amd64 GOOS=darwin build-use-env
	make GOARCH=arm64 GOOS=darwin build-use-env
	make GOARCH=amd64 GOOS=windows build-use-env
	make GOARCH=arm64 GOOS=windows build-use-env

clean:
	rm -rf build
	rm -rf ./schema-diff
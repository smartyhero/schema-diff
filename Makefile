default:
	CGO_ENABLED=0 go build -ldflags="-s -w"
linux-amd64:
	CGO_ENABLED=0 GO_ARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o build/schem-diff-linux-amd64
linux-arm64:
	CGO_ENABLED=0 GO_ARCH=arm64 GOOS=linux go build -ldflags="-s -w" -o build/schem-diff-linux-arm64
darwin-amd64:
	CGO_ENABLED=0 GO_ARCH=amd64 GOOS=darwin go build -ldflags="-s -w" -o build/schem-diff-darwin-amd64
darwin-arm64:
	CGO_ENABLED=0 GO_ARCH=arm64 GOOS=darwin go build -ldflags="-s -w" -o build/schem-diff-darwin-arm64
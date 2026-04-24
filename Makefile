APP     = gitlab-copy
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
OUTDIR  = dist

.PHONY: all clean mac-arm mac-intel linux linux-arm windows

all: mac-arm mac-intel linux linux-arm windows

mac-arm:
	@mkdir -p $(OUTDIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(OUTDIR)/$(APP)-darwin-arm64 ./cmd/

mac-intel:
	@mkdir -p $(OUTDIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(OUTDIR)/$(APP)-darwin-amd64 ./cmd/

linux:
	@mkdir -p $(OUTDIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(OUTDIR)/$(APP)-linux-amd64 ./cmd/

linux-arm:
	@mkdir -p $(OUTDIR)
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(OUTDIR)/$(APP)-linux-arm64 ./cmd/

windows:
	@mkdir -p $(OUTDIR)
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(OUTDIR)/$(APP)-windows-amd64.exe ./cmd/

clean:
	rm -rf $(OUTDIR)

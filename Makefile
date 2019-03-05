linux:export GOOS=linux
linux:export GOARCH=amd64

build: clean
	@echo "Compiling source"
	@rm -rf build
	@mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w" -o build/loRaWAN_Decoder cmd/main.go

linux: build

clean:
	@echo "Cleaning up workspace"
	@rm -rf build
	@rm -rf dist
	@rm -rf docs/public
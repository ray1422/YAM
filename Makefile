all: arm64_linux

arm64_linux:
	mkdir -p bin
	env GOOS=linux GOARCH=arm64 go build -tags netgo -ldflags '-s -w' -o bin/app-linux_arm64
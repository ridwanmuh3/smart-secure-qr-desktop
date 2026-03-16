APP_NAME = smartqrcode
BASE_FLAGS = -clean -obfuscated

.PHONY: build-windows build-darwin build-linux build-all

build-windows:
	wails build $(BASE_FLAGS) -upx -platform windows/amd64 -o $(APP_NAME).exe

build-darwin:
	wails build $(BASE_FLAGS) -platform darwin/universal -o $(APP_NAME)

build-linux:
	wails build $(BASE_FLAGS) -upx -platform linux/amd64 -o $(APP_NAME)

build-all: build-windows build-darwin build-linux

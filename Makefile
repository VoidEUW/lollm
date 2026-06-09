BINARY := lollm
BUILD_DIR := build
PREFIX ?= $(HOME)/.local

.PHONY: build run install clean fmt vet

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./src

run: build
	./$(BUILD_DIR)/$(BINARY)

install: build
	install -d $(PREFIX)/bin
	install -m 0755 $(BUILD_DIR)/$(BINARY) $(PREFIX)/bin/$(BINARY)

fmt:
	gofmt -w ./src

vet:
	go vet ./src

clean:
	rm -rf $(BUILD_DIR)

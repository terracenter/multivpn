BINARY_NAME=multivpn
VERSION=$(shell cat VERSION 2>/dev/null || echo "1.0.0")
GIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "release")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.GitHash=$(GIT_HASH) -s -w"

TAILWIND_BIN ?= $(shell which tailwindcss 2>/dev/null || echo "$$HOME/Workspace/Desarrollo/Sge/Sge-Go/bin/tailwindcss")

.PHONY: all build run test clean build-windows css install uninstall

all: build

css:
	@if [ -x "$(TAILWIND_BIN)" ]; then \
		echo '@import "tailwindcss"; @source "../templates/**/*.html";' > internal/web/static/input.css && \
		$(TAILWIND_BIN) -i internal/web/static/input.css -o internal/web/static/tailwind.css --minify && \
		rm -f internal/web/static/input.css && \
		echo "Tailwind CSS compilado exitosamente."; \
	else \
		echo "Aviso: tailwindcss no encontrado; usando CSS precompilado en el repositorio."; \
	fi

build: css
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

install: build
	@mkdir -p $(HOME)/.local/bin
	@cp $(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)
	@chmod 755 $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "MultiVPN instalado con éxito en $(HOME)/.local/bin/$(BINARY_NAME)"

uninstall:
	@rm -f $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "MultiVPN desinstalado de $(HOME)/.local/bin/$(BINARY_NAME)"

run: build
	./$(BINARY_NAME) web

test:
	go test -v ./...

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME).exe main.go

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe


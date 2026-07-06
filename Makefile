# files-signer — build & packaging
# Leé cada objetivo y corré:  make <objetivo>
# Objetivo por defecto: "make" muestra esta ayuda.

VERSION ?= 0.1.0
MAGICK  ?= magick
NFPM    ?= $(shell go env GOPATH)/bin/nfpm

.DEFAULT_GOAL := help

## help: muestra esta ayuda
help:
	@echo "files-signer — objetivos disponibles:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

## icons: genera los PNG desde assets/icon.svg (necesita ImageMagick)
icons:
	$(MAGICK) -background none assets/icon.svg -resize 128x128 assets/icon-128.png
	$(MAGICK) -background none assets/icon.svg -resize 256x256 assets/icon-256.png
	$(MAGICK) -background none assets/icon.svg -resize 512x512 assets/icon-512.png
	cp assets/icon-512.png assets/icon.png
	cp assets/icon-512.png internal/gui/icon.png
	@echo "OK: iconos generados"

## build: compila los dos binarios (CLI y GUI)
build:
	go build -o files-signer ./cmd/files-signer
	go build -o files-signer-gui ./cmd/files-signer-gui
	@echo "OK: files-signer y files-signer-gui"

## test: corre los tests
test:
	go test ./...

## nfpm: instala nfpm si no está (herramienta Go, sin sudo)
nfpm:
	@test -x "$(NFPM)" || go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
	@echo "OK: $(NFPM)"

## deb: genera el paquete .deb (Ubuntu) en dist/
deb: build nfpm
	@mkdir -p dist
	$(NFPM) pkg --config packaging/nfpm.yaml --packager deb --target dist/

## rpm: genera el paquete .rpm (Fedora) en dist/
rpm: build nfpm
	@mkdir -p dist
	$(NFPM) pkg --config packaging/nfpm.yaml --packager rpm --target dist/

## packages: genera .deb y .rpm
packages: deb rpm
	@echo "OK: paquetes en dist/"
	@ls -1 dist/

## all: iconos + build + tests + paquetes
all: icons build test packages

## clean: borra binarios y paquetes generados
clean:
	rm -f files-signer files-signer-gui
	rm -rf dist

.PHONY: help icons build test nfpm deb rpm packages all clean

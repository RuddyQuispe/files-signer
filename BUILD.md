# Guía de build y empaquetado

Cómo construir files-sign y generar los paquetes `.deb` (Ubuntu) y `.rpm` (Fedora),
paso a paso, para hacerlo vos a mano. Todo está automatizado en el `Makefile`; acá se
explica qué hace cada paso.

## Requisitos (una sola vez)

**Para compilar la CLI:** solo Go. Nada más.

**Para compilar la GUI** (necesita librerías gráficas de desarrollo). En Fedora:

```sh
sudo dnf install -y golang gcc \
  libXxf86vm-devel libX11-devel mesa-libGL-devel \
  libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel
```

**Para el ícono** (SVG → PNG): ImageMagick.

```sh
sudo dnf install -y ImageMagick
```

**Para los paquetes** (`.deb`/`.rpm`): nfpm, que es una herramienta Go (no necesita sudo).
El `Makefile` la instala solo, o a mano:

```sh
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
```

> nfpm queda en `$(go env GOPATH)/bin`. Si `nfpm` no se encuentra, agregá esa carpeta al PATH:
> `export PATH="$PATH:$(go env GOPATH)/bin"`

## Uso rápido (con make)

```sh
make            # muestra la ayuda con todos los objetivos
make icons      # genera los PNG desde assets/icon.svg
make build      # compila files-sign y files-sign-gui
make test       # corre los tests
make packages   # genera dist/*.deb y dist/*.rpm
make all        # todo lo anterior
make clean      # borra binarios y dist/
```

## Qué hace cada paso (a mano, sin make)

### 1. Ícono: SVG → PNG

Genera los PNG en varios tamaños y copia el que la app embebe:

```sh
magick -background none assets/icon.svg -resize 128x128 assets/icon-128.png
magick -background none assets/icon.svg -resize 256x256 assets/icon-256.png
magick -background none assets/icon.svg -resize 512x512 assets/icon-512.png
cp assets/icon-512.png assets/icon.png
cp assets/icon-512.png internal/gui/icon.png    # este lo embebe la GUI (go:embed)
```

### 2. Compilar los binarios

```sh
go build -o files-sign     ./cmd/files-sign      # CLI
go build -o files-sign-gui ./cmd/files-sign-gui  # app de escritorio
```

### 3. Generar los paquetes

```sh
mkdir -p dist
nfpm pkg --config packaging/nfpm.yaml --packager deb --target dist/   # Ubuntu
nfpm pkg --config packaging/nfpm.yaml --packager rpm --target dist/   # Fedora
```

Quedan en `dist/`:
- `files-sign_<version>_amd64.deb`
- `files-sign-<version>-1.x86_64.rpm`

Instalan: binarios en `/usr/bin`, `.desktop` en `/usr/share/applications`, íconos en
`/usr/share/icons/hicolor/...`, metainfo en `/usr/share/metainfo`. Con eso la app aparece
en el menú de aplicaciones con su ícono.

## Instalar y probar el paquete localmente

Fedora:

```sh
sudo dnf install ./dist/files-sign-*.x86_64.rpm
```

Ubuntu:

```sh
sudo apt install ./dist/files-sign_*_amd64.deb
```

## Ver el contenido de un paquete (sin instalar)

```sh
rpm -qlp dist/files-sign-*.rpm     # Fedora
dpkg -c  dist/files-sign_*.deb     # Ubuntu
```

## COPR (repo de Fedora)

Para publicar en COPR se usa `packaging/files-sign.spec`. COPR rebuildea desde el código en
un entorno limpio (mock, sin red), así que conviene versionar las dependencias con
`go mod vendor` para que compile offline. La app ID es `com.rquispe.filessign`.

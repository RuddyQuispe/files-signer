# Guía de build y empaquetado

Cómo construir files-signer y generar los paquetes `.deb` (Ubuntu) y `.rpm` (Fedora),
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
make build      # compila files-signer y files-signer-gui
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
go build -o files-signer     ./cmd/files-signer      # CLI
go build -o files-signer-gui ./cmd/files-signer-gui  # app de escritorio
```

### 3. Generar los paquetes

```sh
mkdir -p dist
nfpm pkg --config packaging/nfpm.yaml --packager deb --target dist/   # Ubuntu
nfpm pkg --config packaging/nfpm.yaml --packager rpm --target dist/   # Fedora
```

Quedan en `dist/`:
- `files-signer_<version>_amd64.deb`
- `files-signer-<version>-1.x86_64.rpm`

Instalan: binarios en `/usr/bin`, `.desktop` en `/usr/share/applications`, íconos en
`/usr/share/icons/hicolor/...`, metainfo en `/usr/share/metainfo`. Con eso la app aparece
en el menú de aplicaciones con su ícono.

## Instalar y probar el paquete localmente

Fedora:

```sh
sudo dnf install ./dist/files-signer-*.x86_64.rpm
```

Ubuntu:

```sh
sudo apt install ./dist/files-signer_*_amd64.deb
```

## Ver el contenido de un paquete (sin instalar)

```sh
rpm -qlp dist/files-signer-*.rpm     # Fedora
dpkg -c  dist/files-signer_*.deb     # Ubuntu
```

## COPR (repo de Fedora)

Para publicar en COPR se usa `packaging/files-signer.spec`. COPR rebuildea desde el código en
un entorno limpio (mock, sin red), así que conviene versionar las dependencias con
`go mod vendor` para que compile offline. La app ID es `com.rquispe.filessigner`.

## Repositorio APT (Ubuntu/Debian) por GitHub Pages

Los usuarios instalan con `apt` desde un repo estático hospedado en GitHub Pages. Lo
construye y firma el workflow `.github/workflows/apt-repo.yml` (corre en un runner Ubuntu:
compila los binarios, arma el `.deb` con nfpm, genera el índice con `apt-ftparchive` y lo
firma con GPG). La lógica del repo está en `packaging/apt/build-repo.sh`.

Estructura publicada en la rama `gh-pages`:

```
KEY.gpg                                          # clave pública para verificar
pool/main/f/files-signer/*.deb                   # paquetes (se acumulan versiones)
dists/stable/main/binary-amd64/Packages[.gz]     # índice
dists/stable/{Release,Release.gpg,InRelease}     # metadatos firmados
```

### Alta (una sola vez)

1. Crear la clave GPG de firma **sin passphrase** (el CI la usa en modo batch):

   ```sh
   gpg --batch --quick-generate-key \
     "files-signer repo <development.rquispe@gmail.com>" default default never
   KEYID=$(gpg --list-secret-keys --with-colons | awk -F: '/^sec/{print $5; exit}')
   gpg --armor --export-secret-keys "$KEYID"    # copiar esta salida al secret
   ```

2. En GitHub → Settings → Secrets and variables → Actions: crear el secret
   **`APT_GPG_PRIVATE_KEY`** con la clave privada exportada arriba.
3. Settings → Pages: source = rama `gh-pages` (la crea el primer run del workflow).
4. El repo debe ser **público** (Pages gratis).

### Publicar una versión

Crear un tag `vX.Y.Z` (o disparar el workflow a mano con el número de versión):

```sh
git tag v0.1.0 && git push origin v0.1.0
```

El workflow compila, firma y publica. Después, en una Ubuntu, las instrucciones de
instalación del [README](README.md#instalar-en-ubuntudebian-apt) deben funcionar y
`apt upgrade` traer las nuevas versiones.

### Probar el script localmente

`build-repo.sh` necesita herramientas Debian (`apt-utils`), así que en Fedora se corre
mejor dentro de un contenedor Ubuntu. Chequeo de sintaxis sin ejecutarlo:

```sh
bash -n packaging/apt/build-repo.sh
```

# files-signer

Firmador y verificador de archivos multiplataforma basado en **PKCS#7 / CMS**.
Firma y verifica **cualquier** tipo de archivo (PDF, YAML, JAR, ZIP, sin extensión, etc.)
usando un certificado y clave en PEM. Alternativa a XolidoSign que corre en **Windows, Linux y Mac**,
con **app de escritorio** (Fyne) y **CLI**.

> ¿Solo querés usarlo? Leé el **[Manual de usuario](MANUAL.md)**.

## Características

- **App de escritorio** (ventana) con guía interactiva tipo spotlight, y **CLI** para automatizar.
- Firma **adjunta** (`.p7m`, con el contenido embebido) y **separada** (`.p7s`, solo firma),
  seleccionables. Extensiones estándar S/MIME; ambas mantienen el nombre completo
  (`documento.pdf.p7m` / `documento.pdf.p7s`).
- Verificación de integridad + identidad del firmante.
- Validación opcional de la cadena de confianza (`--trust`).
- Claves PEM protegidas por contraseña (encriptación legacy y PKCS#8).
- Firma en lote (varios archivos por comando).
- Selector de archivos **nativo del sistema** en la GUI (con fallback).
- Salida compatible con OpenSSL y cualquier verificador CMS estándar.

## Instalar en Ubuntu/Debian (apt)

Repositorio APT nativo (sin snap/flatpak), con actualizaciones vía `apt upgrade`:

```sh
curl -fsSL https://ruddyquispe.github.io/files-signer/KEY.gpg \
  | sudo gpg --dearmor -o /usr/share/keyrings/files-signer.gpg
echo "deb [signed-by=/usr/share/keyrings/files-signer.gpg] \
  https://ruddyquispe.github.io/files-signer stable main" \
  | sudo tee /etc/apt/sources.list.d/files-signer.list
sudo apt update && sudo apt install files-signer
```

Instala el comando `files-signer` (CLI) y la app de escritorio (aparece en el menú).
Cómo se publica y mantiene ese repo está en **[BUILD.md](BUILD.md)**.

## Instalación / compilación

CLI (Go puro, binario estático, sin dependencias):

```sh
go build -o files-signer ./cmd/files-signer

# Cross-compile
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o dist/files-signer-linux      ./cmd/files-signer
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/files-signer.exe         ./cmd/files-signer
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o dist/files-signer-macos-arm64 ./cmd/files-signer
```

App de escritorio (Fyne, requiere toolchain C + libs gráficas para compilar):

```sh
go build -o files-signer-gui ./cmd/files-signer-gui
```

> En Linux la compilación de la GUI necesita las libs de desarrollo de OpenGL/X11
> (en Fedora: `libXxf86vm-devel` y afines). El binario resultante NO requiere instalar nada
> en la máquina del usuario final. Para empaquetar releases se recomienda
> [`fyne-cross`](https://github.com/fyne-io/fyne-cross).

## Uso rápido (CLI)

```sh
# Firmar (genera .p7m y .p7s)
files-signer sign --pem cert.pem --password CLAVE documento.pdf

# Verificar una firma adjunta
files-signer verify documento.pdf.p7m

# Verificar una firma separada
files-signer verify documento.pdf --sig documento.pdf.p7s

# Recuperar el archivo original desde una firma adjunta (.p7m)
files-signer extract documento.pdf.p7m -o documento.pdf
```

Referencia completa en el **[Manual de usuario](MANUAL.md)**.

## Arquitectura

Diseño hexagonal: el dominio (firma/verificación) es independiente de la interfaz y de cómo
se carga la clave. Las interfaces (CLI y GUI) envuelven el mismo dominio.

```
cmd/
  files-signer/         Entry point CLI
  files-signer-gui/     Entry point app de escritorio
internal/
  signing/            DOMINIO: firma y verificación sobre bytes (PKCS#7/CMS)
    sign.go             Signer.Sign (attached / detached)
    verify.go           Verify (+ validación de cadena opcional)
    output.go           OutputMode + SignatureFilenames (.p7m / .p7s)
  keystore/           PUERTO del material de firma
    keystore.go         interface Loader (extensible)
    pem.go              carga PEM con contraseña
  cli/                Interfaz de terminal
    root.go, sign_cmd.go, verify_cmd.go, args.go
  gui/                Interfaz de escritorio (Fyne)
    app.go              ventana + pestañas + tour
    sign_tab.go, verify_tab.go
    filedialog.go       selector nativo (zenity) con fallback a Fyne
    spotlight.go, tour.go   guía interactiva tipo Driver.js
    widgets.go
```

**Regla de oro:** si cambia CÓMO se carga la clave (PFX/P12, keychain) o CÓMO se invoca
(CLI, GUI), el dominio `signing/` no se toca.

## Detalles técnicos

- **Hash:** SHA-256. **Salida:** DER.
- **Dependencias:**
  - [`github.com/digitorus/pkcs7`](https://github.com/digitorus/pkcs7) — operaciones CMS.
  - [`go.step.sm/crypto/pemutil`](https://pkg.go.dev/go.step.sm/crypto/pemutil) — claves PEM
    encriptadas (la stdlib de Go no lee PKCS#8 encriptado).
  - [`fyne.io/fyne/v2`](https://fyne.io) — GUI de escritorio.
  - [`github.com/ncruces/zenity`](https://github.com/ncruces/zenity) — diálogos nativos.
  - `crypto/x509` (stdlib) — validación de cadena de confianza.

## Empaquetado (Linux)

Assets en `packaging/` + `assets/` (icono, `.desktop`, AppStream `metainfo.xml`).
Todo el flujo está en el `Makefile` (`make help`) y explicado paso a paso en **[BUILD.md](BUILD.md)**:

```sh
make icons      # SVG → PNG
make build      # compila CLI + GUI
make packages   # genera dist/*.deb y dist/*.rpm
```

**`.deb` (Ubuntu) y `.rpm` (Fedora) con [nfpm](https://nfpm.goreleaser.com):**

```sh
go build -o files-signer ./cmd/files-signer
go build -o files-signer-gui ./cmd/files-signer-gui
nfpm pkg --config packaging/nfpm.yaml --packager deb --target dist/
nfpm pkg --config packaging/nfpm.yaml --packager rpm --target dist/
```

Instala binarios en `/usr/bin`, el `.desktop` en `/usr/share/applications`, iconos en
`hicolor` y el metainfo en `/usr/share/metainfo` → la app aparece en el menú con su icono.

**COPR (Fedora):** `packaging/files-signer.spec`. Para builds offline de mock, versioná las
dependencias con `go mod vendor`. La app ID es `com.rquispe.filessigner`; para Flathub más
adelante hará falta un ID basado en un dominio/GitHub propio.

## Desarrollo

```sh
go test ./...
go build ./...
```

¿Venís de otro lenguaje y no conocés el ecosistema Go? La
**[Guía de desarrollo](DEVELOPMENT.md)** explica cada comando (módulos, `go run` vs
`go build`, `./...`, cómo funcionan los tests) aplicado a este repo.

## Roadmap

- Formatos de clave `.pfx` / `.p12` (nueva implementación de `keystore.Loader`).
- Sellado de tiempo RFC 3161 (TSA).
- Revocación CRL / OCSP en la validación de confianza.
- Empaquetado para tiendas (Flatpak/Flathub, `.deb`/`.rpm`) y drag-and-drop en la GUI.
```

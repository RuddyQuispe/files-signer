# Guía de desarrollo — files-signer (para quien viene de otro lenguaje)

Para desarrolladores que saben programar pero **no conocen Go**. Cada sección explica
primero el concepto y después el comando de este repo.

> ¿Solo usar el programa? → [Manual de usuario](MANUAL.md).
> ¿Empaquetar `.deb`/`.rpm`? → [BUILD.md](BUILD.md).

---

## 1. Instalar Go

Go es compilado: no hay intérprete. Este proyecto pide Go **1.26.4+** (lo dice `go.mod`).

```sh
go version   # go version go1.26.4 linux/amd64
```

Si no lo tenés: https://go.dev/dl o `sudo dnf install golang` (Fedora).

---

## 2. Cómo se organiza el proyecto

Dos palabras que se confunden:

- **Paquete:** una carpeta con archivos `.go` que se compilan juntos. Es la unidad de
  arquitectura.
- **Módulo:** el proyecto entero. Es la unidad de distribución y versionado.

### `go.mod` y `go.sum` (archivos, no comandos)

Equivalen a lo que ya conocés:

```
go.mod  ≈  package.json / pom.xml   (nombre, versión de Go, dependencias)
go.sum  ≈  package-lock.json         (hash exacto de cada dep; se genera solo, no lo tocás)
```

La primera línea de `go.mod` es `module files-signer`. Ese nombre es la raíz de los
imports internos: por eso el código escribe `import "files-signer/internal/signing"`.

Los gestiona el comando `go mod` (con espacio): `go mod tidy` sincroniza los archivos
con los imports reales del código.

### ¿Un módulo o varios (multi-módulo)?

En Java/Maven separás la arquitectura con varios módulos. **En Go no.** Acá la
arquitectura la dan los **paquetes**; el módulo solo sirve para versionar y publicar por
separado. Este repo es **una sola app que se libera junta** → **un solo módulo**.
Partirlo resolvería un problema que no tenemos (versionar cada capa aparte) y sumaría
trabajo. La separación hexagonal ya está hecha con los paquetes de `internal/`.

### Las carpetas

| Carpeta      | Qué es | ¿Regla o convención? |
|--------------|--------|----------------------|
| `cmd/`       | Puntos de entrada (`func main`). Una subcarpeta = un ejecutable (CLI y GUI). | Convención (Go no la exige) |
| `internal/`  | Toda la lógica. **Solo el propio módulo puede importarla.** | Regla del compilador |
| `assets/`    | Recursos: `icon.svg` (fuente) y los PNG generados. | — |
| `packaging/` | Recetas para armar los instaladores Linux (nfpm, `.desktop`, metainfo). | — |
| `dist/`      | Salida final: `.deb` y `.rpm` ya construidos. Se genera, no se edita. | — |

`internal/` es lo importante: el compilador **prohíbe** importarla desde afuera del
módulo. Es encapsulamiento forzado por el lenguaje. Arquitectura completa en el
[README](README.md#arquitectura).

---

## 3. Ejecutar sin compilar: `go run`

Compila a un binario temporal, lo ejecuta y lo descarta. Ideal para probar mientras
desarrollás.

```sh
go run ./cmd/files-signer sign --pem cert.pem --password CLAVE documento.pdf
go run ./cmd/files-signer verify documento.pdf.p7m
```

`./cmd/files-signer` **no es un archivo, es una ruta a un paquete** (la carpeta con el
`func main`). Lo que va después son argumentos del programa.

---

## 4. Compilar: `go build`

Igual que `go run`, pero **deja el binario en disco**.

```sh
go build -o files-signer ./cmd/files-signer   # deja ./files-signer
./files-signer verify documento.pdf.p7m
```

### ¿Dónde quedan los binarios?

Go **no tiene carpeta de salida por defecto** (no hay `target/` ni `dist/`). Deja el
binario donde estás parado; con `-o` elegís nombre y ruta. Este repo usa `-o` para
dejarlos en la **raíz**: `files-signer` (CLI) y `files-signer-gui` (GUI). Están en
`.gitignore`: no se commitean, se regeneran.

La GUI necesita librerías gráficas para **compilarse** (no para correr). Requisitos en
[BUILD.md](BUILD.md).

```sh
go build -o files-signer-gui ./cmd/files-signer-gui   # requiere libs gráficas (ver BUILD.md)
```

---

## 5. El patrón `./...`

Significa "este directorio y todos los paquetes debajo".

```sh
go build ./...   # compila TODO (detecta errores de compilación)
go test  ./...   # corre TODOS los tests
```

`./cmd/files-signer` = un paquete. `./...` = todos.

---

## 6. Dependencias

No se guardan en el repo. Go las baja a un caché global la primera vez que compilás
(por eso no hay `node_modules`).

```sh
go mod download   # baja las deps declaradas (build/test ya lo hacen solos)
go mod tidy       # sincroniza go.mod/go.sum con los imports del código; corrélo si tocaste imports
```

---

## 7. Tests

Go trae testing integrado, sin frameworks. Reglas:

- Los tests viven **al lado del código**, en archivos `_test.go`.
- Cada test es `func TestXxx(t *testing.T)`. Falla con `t.Fatalf(...)`; si termina sin
  fallar, pasó.

Ejemplo real: [`internal/signing/signing_test.go`](internal/signing/signing_test.go)
(firmar→verificar, tampering, extract, confianza).

```sh
go test ./...                                    # todos
go test -v ./internal/signing                    # -v: muestra cada test
go test -cover ./...                             # agrega % de cobertura
go test -v -run TestAttachedRoundTrip ./internal/signing   # uno solo por nombre
```

Cómo leerlo:

```
ok      files-signer/internal/signing   coverage: 76.7% of statements   → pasó
?       files-signer/cmd/files-signer     [no test files]                 → sin tests (normal)
FAIL    ...                                                            → algo falló
```

---

## 8. Atajos con `make`

`make` da nombres cortos a comandos largos. Es documentación ejecutable: abrís el
`Makefile` y ves qué corre cada uno.

| Comando      | Qué corre por detrás          |
|--------------|-------------------------------|
| `make build` | `go build` de CLI y GUI       |
| `make test`  | `go test ./...`               |
| `make icons` | genera los PNG desde el SVG   |
| `make all`   | iconos + build + test + paquetes |
| `make`       | lista los objetivos            |

Empaquetado (`make packages`, `deb`, `rpm`) en [BUILD.md](BUILD.md).

---

## 9. Limpiar y reconstruir

`make clean` borra binarios y `dist/`. **No toca las dependencias** (viven en el caché
global, no en el repo).

```sh
make clean            # borra files-signer, files-signer-gui y dist/
go clean -cache       # borra el caché de compilación (recompila de cero)
go clean -modcache    # ⚠️ borra TODAS las deps descargadas (global, afecta a todos tus proyectos)
```

De cero: `make clean && go clean -cache && make all`.

---

## 10. Cambiar el logo

Hay **una sola fuente**: `assets/icon.svg`. Todo lo demás se genera.

```sh
# 1. Reemplazá assets/icon.svg
# 2. Regenerá los PNG (incluye el que embebe la GUI):
make icons
# 3. Recompilá para que el binario tome el ícono nuevo:
make build
```

¿Por qué hay una copia en `internal/gui/icon.png`? Porque la GUI **embebe** su ícono con
`//go:embed`, y `go:embed` **no puede leer archivos fuera de su propia carpeta** (no
admite `../`). Por eso `make icons` copia el ícono a `internal/gui/` para que el código
lo alcance. Los `assets/*.png` los usa aparte el empaquetado (van al menú del sistema).

---

## 11. Flujo típico

1. Cambiás código.
2. `go build ./...` — compila todo; si algo no compila, lo ves acá.
3. `go test ./...` — que quede verde.
4. `go run ./cmd/files-signer ...` — probás a mano.

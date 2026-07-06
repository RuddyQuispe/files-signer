# Manual de usuario — files-signer

Guía simple para firmar y verificar archivos. No necesitás saber de criptografía.

---

## ¿Qué hace este programa?

Toma **cualquier** archivo (PDF, YAML, JAR, ZIP, un Dockerfile, un texto sin extensión, lo que sea)
y le pone tu **firma digital** usando tu certificado. Después, cualquiera puede **verificar** que:

1. El archivo lo firmaste vos (con tu certificado).
2. El archivo no fue modificado después de firmarlo.

Es lo mismo que hacías con XolidoSign, pero funciona en Windows, Linux y Mac — con una **app de
escritorio** (ventana) y también por **terminal** para automatizar.

---

## Los dos tipos de firma

Cuando firmás, se pueden generar dos archivos. Son la misma firma, empaquetada de dos formas.
Usamos las extensiones estándar (S/MIME):

| Firma | Qué es | Tamaño | Cómo se llama |
|-------|--------|--------|---------------|
| **Adjunta** (`.p7m`) | La firma **con tu archivo adentro** | Parecido al original | `documento.pdf.p7m` |
| **Separada** (`.p7s`) | **Solo** la firma | Muy chico (unos KB) | `documento.pdf.p7s` |

Podés generar los dos, o solo el que necesites. La extensión no cambia la validez de la firma
(se verifica por contenido, no por nombre); usamos `.p7m`/`.p7s` porque es el estándar.

---

## Antes de empezar: qué necesitás

1. El **programa** `files-signer` (el binario para tu sistema operativo).
2. Tu **certificado** y tu **clave privada** en formato PEM (`.pem`).
3. La **contraseña** de tu clave (si tiene).

> Podés tener el certificado y la clave en un solo archivo `.pem`, o en dos archivos separados.

---

## Usar la app de escritorio (ventana)

Abrí `files-signer-gui`. La primera vez, una **guía interactiva** te resalta paso a paso qué hacer
(podés repetirla con el botón «¿Cómo se usa?»).

- **Pestaña Firmar**: elegí tu certificado con «Examinar» (se abre el selector de archivos nativo
  del sistema), escribí la contraseña, agregá archivos con «Agregar archivo(s)», elegí qué generar
  y tocá **Firmar**.
- **Pestaña Verificar**: elegí el archivo de firma; si es separada (`.p7s`), agregá también el
  original; marcá «Validar confianza» si querés, y tocá **Verificar**. Si la firma es adjunta
  (`.p7m`), podés tocar **Extraer original** para recuperar el archivo embebido.

El resto de este manual explica la versión de **terminal** (útil para automatizar).

---

## Firmar un archivo (terminal)

### Caso simple: un archivo, los dos tipos de firma

```
files-signer sign --pem certificado.pem --password TU_CLAVE documento.pdf
```

Esto crea `documento.pdf.p7m` (adjunta) y `documento.pdf.p7s` (separada) al lado del original.

### Elegir qué firma generar

```
# Solo la firma separada (la chica)
files-signer sign --pem certificado.pem --password TU_CLAVE --out detached documento.pdf

# Solo la firma adjunta (la que contiene el archivo)
files-signer sign --pem certificado.pem --password TU_CLAVE --out attached documento.pdf
```

Opciones de `--out`: `both` (las dos, por defecto), `attached`, `detached`.

### Firmar varios archivos de una vez

```
files-signer sign --pem certificado.pem --password TU_CLAVE app.yaml Dockerfile release.zip
```

### Si el certificado y la clave están en archivos separados

```
files-signer sign --pem certificado.pem --key clave.pem --password TU_CLAVE documento.pdf
```

### Guardar las firmas en otra carpeta

```
files-signer sign --pem certificado.pem --password TU_CLAVE --outdir firmas/ documento.pdf
```

### No querés escribir la contraseña en el comando

Usá `--password-stdin` y el programa la pide por entrada estándar (no queda en el historial):

```
echo "TU_CLAVE" | files-signer sign --pem certificado.pem --password-stdin documento.pdf
```

---

## Verificar un archivo

### Verificar una firma adjunta

El archivo ya está adentro, así que solo pasás la firma:

```
files-signer verify documento.pdf.p7m
```

### Verificar una firma separada

Necesitás el archivo original **y** la firma:

```
files-signer verify documento.pdf --sig documento.pdf.p7s
```

### Recuperar el archivo original desde una firma adjunta (`.p7m`)

Una firma adjunta lleva el archivo adentro. Podés recuperarlo intacto (además de verificarlo):

```
files-signer extract documento.pdf.p7m -o documento.pdf
```

Por defecto, si no ponés `-o`, guarda quitando el `.p7m` del nombre. No sobrescribe un archivo
existente salvo que agregues `-f`. (Sobre una firma separada `.p7s` da error: no contiene el archivo.)

### Verificar además que el certificado es de confianza

Por defecto el programa comprueba que la firma es válida y quién firmó.
Si además querés validar que el certificado viene de una autoridad de confianza,
agregá `--trust` con el archivo de la CA:

```
files-signer verify documento.pdf --sig documento.pdf.p7s --trust --ca autoridad.pem
```

---

## Cómo se lee el resultado

**Firma válida:**

```
VALID: signature is intact
  signer: CN=demo firmante
  trust:  not checked (use --trust --ca ca.pem)
```

**Firma inválida** (archivo modificado, firma corrupta, etc.):

```
INVALID: signature verification failed: pkcs7: Message digest mismatch
```

El programa devuelve código de salida `0` si es válida y distinto de `0` si falla,
así lo podés usar en scripts.

---

## Preguntas frecuentes

**¿Funciona con cualquier tipo de archivo?**
Sí. La firma trabaja sobre los bytes del archivo, no le importa el formato ni la extensión.

**¿Necesito instalar algo más (OpenSSL, Java, etc.)?**
No. Es un solo programa que ya trae todo adentro.

**Me dice "wrong password".**
La contraseña de la clave privada es incorrecta. Revisala.

**Perdí el archivo original y tengo solo la firma separada (`documento.pdf.p7s`).**
La firma separada NO contiene el archivo. Si querés un archivo autocontenido, usá `--out attached` (la firma adjunta, que sí lleva el contenido adentro).

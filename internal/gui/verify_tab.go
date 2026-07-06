package gui

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"files-sign/internal/signing"
)

func newVerifyTab(win fyne.Window) (fyne.CanvasObject, []highlight) {
	sigRow, sigEntry := filePicker(win, "Elegí el archivo de firma (.p7m o .p7s)")
	origRow, origEntry := filePicker(win, "Solo si es una firma separada")
	caRow, caEntry := filePicker(win, "Certificado(s) de confianza .pem")

	trust := widget.NewCheck("Validar que el certificado sea de confianza", nil)

	form := widget.NewForm(
		widget.NewFormItem("Firma (.p7m / .p7s)", sigRow),
		widget.NewFormItem("Archivo original", origRow),
		widget.NewFormItem("", trust),
		widget.NewFormItem("CA de confianza", caRow),
	)

	help := widget.NewLabel(
		"Firma adjunta (.p7m): elegí solo el archivo de firma. Podés extraer el original.\n" +
			"Firma separada (.p7s): elegí la firma y también el archivo original.",
	)
	help.Wrapping = fyne.TextWrapWord

	verifyBtn := widget.NewButtonWithIcon("Verificar", theme.ConfirmIcon(), func() {
		runVerify(win, sigEntry.Text, origEntry.Text, trust.Checked, caEntry.Text)
	})
	verifyBtn.Importance = widget.HighImportance

	extractBtn := widget.NewButtonWithIcon("Extraer original (.p7m)", theme.DownloadIcon(), func() {
		runExtract(win, sigEntry.Text, trust.Checked, caEntry.Text)
	})

	actions := container.NewGridWithColumns(2, verifyBtn, extractBtn)

	content := container.NewBorder(
		container.NewPadded(form),
		container.NewPadded(actions),
		nil, nil,
		container.NewPadded(help),
	)

	highlights := []highlight{
		{sigRow, "1 · Archivo de firma", "Elegí el archivo de firma: .p7m (adjunta, trae el contenido) o .p7s (separada)."},
		{trust, "2 · Confianza (opcional)", "Marcá esto si además querés validar que el certificado sea de una autoridad de confianza."},
		{verifyBtn, "3 · Verificar", "Tocá acá para comprobar que la firma es válida y que el archivo no fue modificado."},
		{extractBtn, "4 · Extraer original", "Si la firma es adjunta (.p7m), recuperá el archivo original intacto que lleva embebido."},
	}
	return content, highlights
}

// guiTrustPool builds the trust pool for verify/extract from the GUI inputs.
// Returns ok=false (after showing a dialog) when the user asked for trust but
// the inputs are invalid.
func guiTrustPool(win fyne.Window, trust bool, caPath string) (*x509.CertPool, bool) {
	if !trust {
		return nil, true
	}
	if caPath == "" {
		dialog.ShowInformation("Falta la CA", "Para validar la confianza, elegí el archivo de la CA.", win)
		return nil, false
	}
	caPEM, err := os.ReadFile(caPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("No se pudo leer la CA:\n\n%w", err), win)
		return nil, false
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		dialog.ShowError(fmt.Errorf("El archivo de CA no contiene certificados válidos"), win)
		return nil, false
	}
	return pool, true
}

func runVerify(win fyne.Window, sigPath, origPath string, trust bool, caPath string) {
	if sigPath == "" {
		dialog.ShowInformation("Falta la firma", "Elegí el archivo de firma (.p7m o .p7s).", win)
		return
	}
	signature, err := os.ReadFile(sigPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("No se pudo leer la firma:\n\n%w", err), win)
		return
	}

	var content []byte
	if origPath != "" {
		content, err = os.ReadFile(origPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("No se pudo leer el archivo original:\n\n%w", err), win)
			return
		}
	}

	pool, ok := guiTrustPool(win, trust, caPath)
	if !ok {
		return
	}

	res, err := signing.Verify(signature, content, pool)
	if err != nil {
		dialog.ShowError(fmt.Errorf("FIRMA INVÁLIDA — el archivo fue modificado o la firma no es correcta.\n\n%w", err), win)
		return
	}

	msg := "La firma es válida: el archivo no fue modificado."
	if res.Signer != nil {
		msg += fmt.Sprintf("\n\nFirmante: %s", res.Signer.Subject)
	}
	if res.TrustChecked {
		msg += "\nConfianza: cadena de certificados validada."
	} else {
		msg += "\nConfianza: no verificada."
	}
	dialog.ShowInformation("Firma válida ✓", msg, win)
}

func runExtract(win fyne.Window, sigPath string, trust bool, caPath string) {
	if sigPath == "" {
		dialog.ShowInformation("Falta la firma", "Elegí un archivo de firma adjunta (.p7m).", win)
		return
	}
	signature, err := os.ReadFile(sigPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("No se pudo leer la firma:\n\n%w", err), win)
		return
	}

	pool, ok := guiTrustPool(win, trust, caPath)
	if !ok {
		return
	}

	content, res, err := signing.Extract(signature, pool)
	if err != nil {
		dialog.ShowError(fmt.Errorf("No se pudo extraer el original:\n\n%w", err), win)
		return
	}

	defaultName := strings.TrimSuffix(filepath.Base(sigPath), ".p7m")
	pickSave(win, "Guardar archivo original", defaultName, func(dest string) {
		if err := os.WriteFile(dest, content, 0o644); err != nil {
			dialog.ShowError(fmt.Errorf("No se pudo guardar:\n\n%w", err), win)
			return
		}
		msg := "Archivo original recuperado y guardado."
		if res.Signer != nil {
			msg += fmt.Sprintf("\n\nFirmante: %s", res.Signer.Subject)
		}
		dialog.ShowInformation("Extraído ✓", msg, win)
	})
}

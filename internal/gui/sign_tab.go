package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"files-signer/internal/keystore"
	"files-signer/internal/signing"
)

const (
	optBoth     = "Ambas (adjunta + separada)"
	optAttached = "Solo adjunta (incluye el archivo)"
	optDetached = "Solo separada (la firma sola)"
)

type signTab struct {
	win      fyne.Window
	pem      *widget.Entry
	key      *widget.Entry
	pass     *widget.Entry
	mode     *widget.RadioGroup
	files    []string
	list     *widget.List
	selected int
}

func newSignTab(win fyne.Window) (fyne.CanvasObject, []highlight) {
	s := &signTab{win: win, selected: -1}

	certRow, pem := filePicker(win, "Elegí tu certificado .pem")
	keyRow, key := filePicker(win, "Solo si la clave está en otro archivo")
	s.pem, s.key = pem, key

	s.pass = widget.NewPasswordEntry()
	s.pass.SetPlaceHolder("Contraseña de la clave (si tiene)")

	s.mode = widget.NewRadioGroup([]string{optBoth, optAttached, optDetached}, nil)
	s.mode.SetSelected(optBoth)

	form := widget.NewForm(
		widget.NewFormItem("Certificado", certRow),
		widget.NewFormItem("Clave (opcional)", keyRow),
		widget.NewFormItem("Contraseña", s.pass),
		widget.NewFormItem("Generar", s.mode),
	)

	s.list = widget.NewList(
		func() int { return len(s.files) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(s.files[i])
		},
	)
	s.list.OnSelected = func(id widget.ListItemID) { s.selected = id }
	s.list.OnUnselected = func(id widget.ListItemID) {
		if s.selected == id {
			s.selected = -1
		}
	}

	addBtn := widget.NewButtonWithIcon("Agregar archivo(s)", theme.ContentAddIcon(), func() {
		pickFiles(win, "Elegí archivos a firmar", func(paths []string) {
			s.files = append(s.files, paths...)
			s.list.Refresh()
		})
	})
	removeBtn := widget.NewButtonWithIcon("Quitar", theme.DeleteIcon(), func() {
		if s.selected >= 0 && s.selected < len(s.files) {
			s.files = append(s.files[:s.selected], s.files[s.selected+1:]...)
			s.selected = -1
			s.list.UnselectAll()
			s.list.Refresh()
		}
	})

	filesHeader := container.NewVBox(
		widget.NewLabelWithStyle("Archivos a firmar", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(addBtn, removeBtn),
	)
	filesSection := container.NewBorder(filesHeader, nil, nil, nil, s.list)

	signBtn := widget.NewButtonWithIcon("Firmar", theme.DocumentCreateIcon(), s.run)
	signBtn.Importance = widget.HighImportance

	content := container.NewBorder(
		container.NewPadded(form),
		container.NewPadded(signBtn),
		nil, nil,
		container.NewPadded(filesSection),
	)

	highlights := []highlight{
		{certRow, "1 · Tu certificado", "Elegí tu archivo .pem (certificado). Si la clave está en otro archivo, cargala en «Clave», y escribí la contraseña si tu clave la tiene."},
		{filesSection, "2 · Archivos a firmar", "Agregá uno o varios archivos y elegí qué generar: ambas firmas, solo la adjunta (incluye el archivo) o solo la separada."},
		{signBtn, "3 · Firmar", "Cuando esté todo listo, tocá acá. Se crean los archivos .p7m y/o .p7s junto a cada archivo original."},
	}
	return content, highlights
}

func (s *signTab) run() {
	if s.pem.Text == "" {
		dialog.ShowInformation("Falta el certificado", "Elegí el archivo .pem de tu certificado.", s.win)
		return
	}
	if len(s.files) == 0 {
		dialog.ShowInformation("Sin archivos", "Agregá al menos un archivo para firmar.", s.win)
		return
	}

	mode := signing.Both
	switch s.mode.Selected {
	case optAttached:
		mode = signing.AttachedOnly
	case optDetached:
		mode = signing.DetachedOnly
	}

	material, err := keystore.PEMLoader{
		CertFile: s.pem.Text,
		KeyFile:  s.key.Text,
		Password: s.pass.Text,
	}.Load()
	if err != nil {
		dialog.ShowError(fmt.Errorf("No se pudo cargar el certificado o la clave (¿contraseña correcta?):\n\n%w", err), s.win)
		return
	}
	signer := signing.Signer{Material: material}

	var report strings.Builder
	failures := 0
	for _, f := range s.files {
		if err := signFile(signer, f, mode, &report); err != nil {
			fmt.Fprintf(&report, "✗ %s: %v\n", filepath.Base(f), err)
			failures++
		}
	}

	if failures > 0 {
		dialog.ShowError(fmt.Errorf("Terminó con %d error(es):\n\n%s", failures, report.String()), s.win)
		return
	}
	dialog.ShowInformation("Firma completada", report.String(), s.win)
}

func signFile(signer signing.Signer, path string, mode signing.OutputMode, report *strings.Builder) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	res, err := signer.Sign(content, mode)
	if err != nil {
		return err
	}

	attached, detached := signing.SignatureFilenames(path, mode)
	if res.Attached != nil {
		if err := os.WriteFile(attached, res.Attached, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(report, "✓ %s\n", filepath.Base(attached))
	}
	if res.Detached != nil {
		if err := os.WriteFile(detached, res.Detached, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(report, "✓ %s\n", filepath.Base(detached))
	}
	return nil
}

// Package gui is the desktop interface (Fyne). It wraps the same signing domain
// that the CLI uses; no cryptographic logic lives here.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// filePicker builds an entry with a Browse button that opens the native OS file
// chooser and fills the entry with the chosen path. Returns the row to place in
// a form and the entry to read from.
func filePicker(win fyne.Window, placeholder string) (fyne.CanvasObject, *widget.Entry) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)

	browse := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		pickFile(win, "Elegí un archivo", func(path string) {
			entry.SetText(path)
		})
	})

	return container.NewBorder(nil, nil, nil, browse, entry), entry
}

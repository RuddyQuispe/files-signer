package gui

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/ncruces/zenity"
)

// pickFile opens the native OS file chooser and calls onPick with the selected
// path. If the native dialog is unavailable, it falls back to Fyne's own dialog.
func pickFile(win fyne.Window, title string, onPick func(string)) {
	go func() {
		path, err := zenity.SelectFile(zenity.Title(title))
		switch {
		case err == nil:
			fyne.Do(func() { onPick(path) })
		case errors.Is(err, zenity.ErrCanceled):
			// user cancelled — do nothing
		default:
			fyne.Do(func() { fyneOpenFallback(win, onPick) })
		}
	}()
}

// pickFiles opens the native OS chooser allowing multiple selection at once.
func pickFiles(win fyne.Window, title string, onPick func([]string)) {
	go func() {
		paths, err := zenity.SelectFileMultiple(zenity.Title(title))
		switch {
		case err == nil:
			fyne.Do(func() { onPick(paths) })
		case errors.Is(err, zenity.ErrCanceled):
			// user cancelled — do nothing
		default:
			fyne.Do(func() {
				fyneOpenFallback(win, func(p string) { onPick([]string{p}) })
			})
		}
	}()
}

// fyneOpenFallback shows Fyne's built-in file dialog, enlarged, when the native
// chooser is not available.
func fyneOpenFallback(win fyne.Window, onPick func(string)) {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil || r == nil {
			return
		}
		defer r.Close()
		onPick(r.URI().Path())
	}, win)
	d.Resize(fyne.NewSize(900, 600))
	d.Show()
}

// pickSave opens the native OS "save file" chooser and calls onPick with the
// chosen destination path. Falls back to Fyne's dialog if the native one fails.
func pickSave(win fyne.Window, title, defaultName string, onPick func(string)) {
	go func() {
		path, err := zenity.SelectFileSave(
			zenity.Title(title),
			zenity.Filename(defaultName),
			zenity.ConfirmOverwrite(),
		)
		switch {
		case err == nil:
			fyne.Do(func() { onPick(path) })
		case errors.Is(err, zenity.ErrCanceled):
			// user cancelled — do nothing
		default:
			fyne.Do(func() { fyneSaveFallback(win, defaultName, onPick) })
		}
	}()
}

// fyneSaveFallback shows Fyne's built-in save dialog when the native one is not
// available. It only resolves the destination path; the caller writes the bytes.
func fyneSaveFallback(win fyne.Window, defaultName string, onPick func(string)) {
	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil || w == nil {
			return
		}
		path := w.URI().Path()
		w.Close()
		onPick(path)
	}, win)
	d.SetFileName(defaultName)
	d.Resize(fyne.NewSize(900, 600))
	d.Show()
}

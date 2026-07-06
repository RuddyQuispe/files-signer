package gui

import (
	_ "embed"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed icon.png
var iconPNG []byte

const prefTourCompleted = "tour_completed"

// Run starts the desktop application.
func Run() {
	a := app.NewWithID("com.rquispe.filessign")
	a.SetIcon(fyne.NewStaticResource("icon.png", iconPNG))
	w := a.NewWindow("files-sign — Firmador y verificador de archivos")
	w.Resize(fyne.NewSize(700, 640))

	signContent, signHi := newSignTab(w)
	verifyContent, verifyHi := newVerifyTab(w)

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Firmar", theme.DocumentCreateIcon(), signContent),
		container.NewTabItemWithIcon("Verificar", theme.ConfirmIcon(), verifyContent),
	)

	targets := buildTourTargets(signHi, verifyHi)

	help := widget.NewButtonWithIcon("¿Cómo se usa?", theme.HelpIcon(), func() {
		startTour(w, tabs, targets, nil)
	})
	title := widget.NewLabelWithStyle("files-sign", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	header := container.NewBorder(nil, nil, title, help)

	w.SetContent(container.NewBorder(container.NewPadded(header), nil, nil, nil, tabs))

	// Guided tour on first run; the "¿Cómo se usa?" button replays it anytime.
	if !a.Preferences().Bool(prefTourCompleted) {
		go func() {
			time.Sleep(600 * time.Millisecond)
			fyne.Do(func() {
				startTour(w, tabs, targets, func() {
					a.Preferences().SetBool(prefTourCompleted, true)
				})
			})
		}()
	}

	w.ShowAndRun()
}

func buildTourTargets(signHi, verifyHi []highlight) []tourTarget {
	targets := []tourTarget{
		{tabIndex: -1, highlight: highlight{
			title: "Bienvenido a files-sign 👋",
			body:  "Te muestro en pocos pasos cómo firmar y verificar archivos. Podés repetir esta guía cuando quieras con el botón «¿Cómo se usa?».",
		}},
	}
	for _, h := range signHi {
		targets = append(targets, tourTarget{tabIndex: 0, highlight: h})
	}
	for _, h := range verifyHi {
		targets = append(targets, tourTarget{tabIndex: 1, highlight: h})
	}
	return targets
}

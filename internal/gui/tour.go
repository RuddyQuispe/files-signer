package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// highlight is a widget the tour points at, with its explanation.
type highlight struct {
	obj   fyne.CanvasObject
	title string
	body  string
}

// tourTarget is a highlight plus the tab it lives on (-1 = welcome / no tab).
type tourTarget struct {
	tabIndex int
	highlight
}

// startTour runs a Driver.js-style guided walkthrough over the real widgets.
// onFinish (may be nil) runs when the tour ends, whether finished or skipped.
func startTour(win fyne.Window, tabs *container.AppTabs, targets []tourTarget, onFinish func()) {
	sp := &spotlight{win: win}

	done := func() {
		sp.hide()
		if onFinish != nil {
			onFinish()
		}
	}

	var show func(i int)
	show = func(i int) {
		t := targets[i]
		if t.tabIndex >= 0 {
			tabs.SelectIndex(t.tabIndex)
		}
		// Let the (possibly just-switched) tab lay out before measuring positions.
		go func() {
			time.Sleep(150 * time.Millisecond)
			fyne.Do(func() {
				var pos fyne.Position
				var size fyne.Size
				if t.obj != nil {
					pos = fyne.CurrentApp().Driver().AbsolutePositionForObject(t.obj)
					size = t.obj.Size()
				}
				sp.show(pos, size, buildCard(i, len(targets), t, sp, show, done))
			})
		}()
	}

	show(0)
}

func buildCard(i, total int, t tourTarget, sp *spotlight, show func(int), done func()) fyne.CanvasObject {
	body := widget.NewLabel(t.body)
	body.Wrapping = fyne.TextWrapWord

	progress := widget.NewLabelWithStyle(
		fmt.Sprintf("Paso %d de %d", i+1, total),
		fyne.TextAlignLeading, fyne.TextStyle{Italic: true},
	)

	skip := widget.NewButton("Saltar", func() { done() })
	back := widget.NewButton("Anterior", func() { sp.hide(); show(i - 1) })
	if i == 0 {
		back.Disable()
	}

	var next *widget.Button
	if i == total-1 {
		next = widget.NewButton("¡Entendido!", func() { done() })
	} else {
		next = widget.NewButton("Siguiente", func() { sp.hide(); show(i + 1) })
	}
	next.Importance = widget.HighImportance

	buttons := container.NewBorder(nil, nil, skip, container.NewHBox(back, next), progress)
	content := container.NewVBox(body, widget.NewSeparator(), buttons)
	return widget.NewCard(t.title, "", content)
}

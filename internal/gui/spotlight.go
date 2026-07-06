package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

// spotlight draws a Driver.js-style overlay: it dims the whole window except for
// a rectangular "hole" around a target widget, outlines that target, and shows a
// card (the tour step) next to it. The dim is built from four rectangles around
// the target, so no alpha compositing is needed.
type spotlight struct {
	win     fyne.Window
	current fyne.CanvasObject
}

func (s *spotlight) hide() {
	if s.current != nil {
		s.win.Canvas().Overlays().Remove(s.current)
		s.current = nil
	}
}

// show renders the overlay around the target rect (pos/size). A zero size dims
// the whole canvas and centers the card (used for the welcome step).
func (s *spotlight) show(pos fyne.Position, size fyne.Size, card fyne.CanvasObject) {
	s.hide()

	canvasSize := s.win.Canvas().Size()
	dim := color.NRGBA{R: 0, G: 0, B: 0, A: 175}
	newDim := func() *canvas.Rectangle { return canvas.NewRectangle(dim) }

	var objs []fyne.CanvasObject

	if size.IsZero() {
		full := newDim()
		full.Move(fyne.NewPos(0, 0))
		full.Resize(canvasSize)
		objs = append(objs, full)
	} else {
		pad := float32(6)
		tx, ty := pos.X-pad, pos.Y-pad
		tw, th := size.Width+2*pad, size.Height+2*pad

		top := newDim()
		top.Move(fyne.NewPos(0, 0))
		top.Resize(fyne.NewSize(canvasSize.Width, max(0, ty)))

		bottom := newDim()
		bottom.Move(fyne.NewPos(0, ty+th))
		bottom.Resize(fyne.NewSize(canvasSize.Width, max(0, canvasSize.Height-(ty+th))))

		left := newDim()
		left.Move(fyne.NewPos(0, ty))
		left.Resize(fyne.NewSize(max(0, tx), th))

		right := newDim()
		right.Move(fyne.NewPos(tx+tw, ty))
		right.Resize(fyne.NewSize(max(0, canvasSize.Width-(tx+tw)), th))

		border := canvas.NewRectangle(color.Transparent)
		border.StrokeColor = theme.Color(theme.ColorNamePrimary)
		border.StrokeWidth = 3
		border.CornerRadius = 6
		border.Move(fyne.NewPos(tx, ty))
		border.Resize(fyne.NewSize(tw, th))

		objs = append(objs, top, bottom, left, right, border)
	}

	// Place the card: below the target, or above/centered if it doesn't fit.
	cardW := float32(380)
	cardH := card.MinSize().Height
	var cx, cy float32
	if size.IsZero() {
		cx = (canvasSize.Width - cardW) / 2
		cy = (canvasSize.Height - cardH) / 2
	} else {
		cx = pos.X
		if cx+cardW > canvasSize.Width {
			cx = canvasSize.Width - cardW - 8
		}
		cx = max(8, cx)

		cy = pos.Y + size.Height + 12
		if cy+cardH > canvasSize.Height {
			cy = pos.Y - cardH - 12 // not enough room below → above
		}
		cy = max(8, cy)
	}
	card.Move(fyne.NewPos(cx, cy))
	card.Resize(fyne.NewSize(cardW, cardH))
	objs = append(objs, card)

	overlay := container.NewWithoutLayout(objs...)
	overlay.Resize(canvasSize)
	s.current = overlay
	s.win.Canvas().Overlays().Add(overlay)
}

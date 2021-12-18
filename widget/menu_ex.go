package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/internal/widget"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
)

type menuRendererEx struct {
	menuRenderer
	height float32
}

func (r *menuRendererEx) Layout(s fyne.Size) {
	minSize := r.MinSize()
	var boxSize fyne.Size
	if r.m.customSized {
		boxSize = minSize.Max(s)
	} else {
		boxSize = minSize
	}
	scrollSize := boxSize
	if c := fyne.CurrentApp().Driver().CanvasForObject(r.m.super()); c != nil {
		ap := fyne.CurrentApp().Driver().AbsolutePositionForObject(r.m.super())
		pos, size := c.InteractiveArea()
		bottomPad := c.Size().Height - pos.Y - size.Height
		if ah := c.Size().Height - bottomPad - ap.Y; ah < boxSize.Height {
			scrollSize = fyne.NewSize(boxSize.Width, ah)
		}
	}
	if scrollSize != r.m.Size() {
		r.m.Resize(scrollSize)
		return
	}

	r.LayoutShadow(scrollSize, fyne.NewPos(0, -(r.height+2*theme.Padding())))
	r.scroll.Resize(scrollSize)
	r.box.Resize(boxSize)
	r.layoutActiveChild()
}

// ----------------

type searchBox struct {
	BaseWidget
	items  []fyne.CanvasObject
	input  *Entry
	height float32
}

var _ fyne.Widget = (*searchBox)(nil)

func (b *searchBox) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(theme.BackgroundColor())
	cont := fyne.NewContainerWithLayout(layout.NewVBoxLayout(), b.items...)
	return &searchBoxRenderer{
		BaseRenderer: widget.NewBaseRenderer([]fyne.CanvasObject{background, cont}),
		b:            b,
		background:   background,
		cont:         cont,
	}
}

func newSearchBox(height float32) *searchBox {
	search := NewEntry()
	search.Validator = nil
	b := &searchBox{
		items:  []fyne.CanvasObject{search},
		input:  search,
		height: height,
	}
	b.Resize(fyne.NewSize(1, height))
	b.ExtendBaseWidget(b)
	return b
}

type searchBoxRenderer struct {
	widget.BaseRenderer
	b          *searchBox
	background *canvas.Rectangle
	cont       *fyne.Container
}

var _ fyne.WidgetRenderer = (*searchBoxRenderer)(nil)

func (r *searchBoxRenderer) Layout(size fyne.Size) {
	s := fyne.NewSize(size.Width, size.Height+2*theme.Padding())
	r.background.Resize(s)
	r.cont.Resize(s)
	r.cont.Move(fyne.NewPos(0, theme.Padding()))
}

func (r *searchBoxRenderer) MinSize() fyne.Size {
	return r.cont.MinSize().Add(fyne.NewSize(0, 2*theme.Padding()))
}

func (r *searchBoxRenderer) Refresh() {
	r.background.FillColor = theme.BackgroundColor()
	r.background.Refresh()
	canvas.Refresh(r.b)
}

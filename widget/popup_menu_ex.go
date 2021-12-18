package widget

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/widget"
	"fyne.io/fyne/v2/theme"
)

var _ fyne.Widget = (*PopUpMenuEx)(nil)
var _ fyne.Focusable = (*PopUpMenuEx)(nil)

// PopUpMenuEx is a Menu which displays itself in an OverlayContainer.
type PopUpMenuEx struct {
	*Menu
	canvas  fyne.Canvas
	overlay *widget.OverlayContainer

	menuScroll   *widget.Scroll
	search       *searchBox
	searchOn     bool
	searchHeight float32
	searchFunc   func(s string)
}

// NewPopUpMenuEx creates a new, reusable popup menu. You can show it using ShowAtPosition.
//
// Since: 2.0
func NewPopUpMenuEx(menu *fyne.Menu, c fyne.Canvas, searchHeight float32) *PopUpMenuEx {
	m := &Menu{}
	m.setMenu(menu)
	p := &PopUpMenuEx{Menu: m, canvas: c, searchOn: false}
	if searchHeight >= 20 {
		p.searchHeight = searchHeight
		p.searchOn = true
		p.searchFunc = func(s string) {
			find := strings.ToLower(s)
			for _, item := range p.Items {
				if mItem, ok := item.(*menuItem); ok && strings.Contains(strings.ToLower(mItem.Item.Label), find) {
					p.menuScroll.Offset.Y = mItem.Position().Y
					p.menuScroll.Refresh()
					p.Menu.activateItem(mItem)
					return
				}
			}
		}
	}
	p.ExtendBaseWidget(p)
	p.Menu.Resize(p.Menu.MinSize())
	p.Menu.customSized = true
	o := widget.NewOverlayContainer(p, c, p.Dismiss)
	o.Resize(o.MinSize())
	p.overlay = o
	p.OnDismiss = func() {
		p.Hide()
	}
	return p
}

// FocusGained is triggered when the object gained focus. For the pop-up menu it does nothing.
//
// Implements: fyne.Focusable
func (p *PopUpMenuEx) FocusGained() {}

// FocusLost is triggered when the object lost focus. For the pop-up menu it does nothing.
//
// Implements: fyne.Focusable
func (p *PopUpMenuEx) FocusLost() {}

// Hide hides the pop-up menu.
//
// Implements: fyne.Widget
func (p *PopUpMenuEx) Hide() {
	p.overlay.Hide()
	p.Menu.Hide()
}

// Move moves the pop-up menu.
// The position is absolute because pop-up menus are shown in an overlay which covers the whole canvas.
//
// Implements: fyne.Widget
func (p *PopUpMenuEx) Move(pos fyne.Position) {
	y := pos.Y
	if p.search != nil {
		y = y + p.search.height + 2*theme.Padding()
	}
	p.BaseWidget.Move(fyne.NewPos(pos.X, y))
}

// Resize changes the size of the pop-up menu.
//
// Implements: fyne.Widget
func (p *PopUpMenuEx) Resize(size fyne.Size) {
	p.BaseWidget.Move(p.Position())
	p.Menu.Resize(size)

	if p.search != nil {
		p.search.Move(fyne.NewPos(0, -(p.search.Size().Height + 2*theme.Padding())))
		p.search.Resize(fyne.NewSize(p.Menu.size.Width, p.search.Size().Height))
	}
}

// Show makes the pop-up menu visible.
//
// Implements: fyne.Widget
func (p *PopUpMenuEx) Show() {
	p.Menu.alignment = p.alignment
	p.Menu.Refresh()

	p.overlay.Show()
	p.Menu.Show()
	if !fyne.CurrentDevice().IsMobile() {
		p.canvas.Focus(p)
	}
}

// ShowAtPosition shows the pop-up menu at the specified position.
func (p *PopUpMenuEx) ShowAtPosition(pos fyne.Position) {
	p.Move(pos)
	p.Show()
}

// TypedKey handles key events. It allows keyboard control of the pop-up menu.
//
// Implements: fyne.Focusable
func (p *PopUpMenuEx) TypedKey(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyDown:
		p.ActivateNext()
	case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		p.TriggerLast()
	case fyne.KeyEscape:
		p.Dismiss()
	case fyne.KeyLeft:
		p.DeactivateLastSubmenu()
	case fyne.KeyRight:
		p.ActivateLastSubmenu()
	case fyne.KeyUp:
		p.ActivatePrevious()
	}
}

func (p *PopUpMenuEx) CreateRenderer() fyne.WidgetRenderer {
	if !p.searchOn {
		return p.Menu.CreateRenderer()
	}

	p.search = newSearchBox(p.searchHeight)
	p.search.input.OnChanged = p.searchFunc
	p.Menu.ExtendBaseWidget(p.Menu)
	box := newMenuBox(p.Menu.Items)
	p.menuScroll = widget.NewVScroll(box)
	p.menuScroll.SetMinSize(box.MinSize())
	objects := []fyne.CanvasObject{p.menuScroll, p.search}
	for _, i := range p.Menu.Items {
		if item, ok := i.(*menuItem); ok && item.Child() != nil {
			objects = append(objects, item.Child())
		}
	}
	return &menuRendererEx{
		menuRenderer: menuRenderer{
			ShadowingRenderer: widget.NewShadowingRenderer(objects, widget.MenuLevel),
			box:               box,
			m:                 p.Menu,
			scroll:            p.menuScroll,
		},
		height: p.search.height,
	}
}

// TypedRune handles text events. For pop-up menus this does nothing.
//
// Implements: fyne.Focusable
func (p *PopUpMenuEx) TypedRune(rune) {}

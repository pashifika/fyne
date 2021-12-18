package widget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

// SelectEx widget has a list of options, with the current one shown, and triggers an event func when clicked
type SelectEx struct {
	DisableableWidget

	// Alignment sets the text alignment of the select and its list of options.
	//
	// Since: 2.1
	Alignment   fyne.TextAlign
	Selected    SelectOption
	Options     []SelectOption
	PlaceHolder string
	OnChanged   func(opt SelectOption) `json:"-"`

	focused bool
	hovered bool
	popUp   *PopUpMenuEx
	tapAnim *fyne.Animation
	binder  binding.String

	searchBoxHeight float32
}

type SelectOption interface {
	Label() string
	Value() string
}

type _unSelectOption struct {
}

func (u _unSelectOption) Label() string {
	return ""
}

func (u _unSelectOption) Value() string {
	return ""
}

var _ fyne.Widget = (*SelectEx)(nil)
var _ desktop.Hoverable = (*SelectEx)(nil)
var _ fyne.Tappable = (*SelectEx)(nil)
var _ fyne.Focusable = (*SelectEx)(nil)
var _ fyne.Disableable = (*SelectEx)(nil)

// NewSelectEx creates a new select widget with the set list of options and changes handler
func NewSelectEx(options []SelectOption, placeHolder string, binder binding.String,
	searchBoxHeight float32, changed func(opt SelectOption)) *SelectEx {
	if placeHolder == "" {
		placeHolder = defaultPlaceHolder
	}
	s := &SelectEx{
		OnChanged:   changed,
		Options:     options,
		PlaceHolder: placeHolder,
		binder:      binder,

		searchBoxHeight: searchBoxHeight,
	}
	s.ExtendBaseWidget(s)
	return s
}

// ClearSelected clears the current option of the select widget.  After
// clearing the current option, the Select widget's PlaceHolder will
// be displayed.
func (s *SelectEx) ClearSelected() {
	s.updateSelected(_unSelectOption{})
}

// SyncSelected sync binder to default selected item
func (s *SelectEx) SyncSelected() *SelectEx {
	if s.binder != nil {
		if val, err := s.binder.Get(); err == nil {
			for _, opt := range s.Options {
				if opt.Value() == val {
					s.updateSelected(opt)
					break
				}
			}
		}
	}
	return s
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (s *SelectEx) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)
	s.propertyLock.RLock()
	icon := NewIcon(theme.MenuDropDownIcon())
	if s.PlaceHolder == "" {
		s.PlaceHolder = defaultPlaceHolder
	}
	if s.Selected == nil {
		s.Selected = _unSelectOption{}
	}
	txtProv := NewRichTextWithText(s.Selected.Label())
	txtProv.inset = fyne.NewSize(theme.Padding(), theme.Padding())
	txtProv.ExtendBaseWidget(txtProv)
	txtProv.Wrapping = fyne.TextTruncate
	if s.disabled {
		txtProv.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameDisabled
	}

	background := &canvas.Rectangle{}
	line := canvas.NewRectangle(theme.ShadowColor())
	tapBG := canvas.NewRectangle(color.Transparent)
	s.tapAnim = newButtonTapAnimation(tapBG, s)
	s.tapAnim.Curve = fyne.AnimationEaseOut
	objects := []fyne.CanvasObject{background, line, tapBG, txtProv, icon}
	r := &selectRendererEx{icon, txtProv, background, line, objects, s}
	background.FillColor, line.FillColor = r.bgLineColor()
	r.updateIcon()
	s.propertyLock.RUnlock() // updateLabel and some text handling isn't quite right, resolve in text refactor for 2.0
	r.updateLabel()
	return r
}

// FocusGained is called after this Select has gained focus.
//
// Implements: fyne.Focusable
func (s *SelectEx) FocusGained() {
	s.focused = true
	s.Refresh()
}

// FocusLost is called after this Select has lost focus.
//
// Implements: fyne.Focusable
func (s *SelectEx) FocusLost() {
	s.focused = false
	s.Refresh()
}

// Hide hides the select.
//
// Implements: fyne.Widget
func (s *SelectEx) Hide() {
	if s.popUp != nil {
		s.popUp.Hide()
		s.popUp = nil
	}
	s.BaseWidget.Hide()
}

// MinSize returns the size that this widget should not shrink below
func (s *SelectEx) MinSize() fyne.Size {
	s.ExtendBaseWidget(s)
	return s.BaseWidget.MinSize()
}

// MouseIn is called when a desktop pointer enters the widget
func (s *SelectEx) MouseIn(*desktop.MouseEvent) {
	s.hovered = true
	s.Refresh()
}

// MouseMoved is called when a desktop pointer hovers over the widget
func (s *SelectEx) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut is called when a desktop pointer exits the widget
func (s *SelectEx) MouseOut() {
	s.hovered = false
	s.Refresh()
}

// Move changes the relative position of the select.
//
// Implements: fyne.Widget
func (s *SelectEx) Move(pos fyne.Position) {
	s.BaseWidget.Move(pos)

	if s.popUp != nil {
		s.popUp.Move(s.popUpPos())
	}
}

// Resize sets a new size for a widget.
// Note this should not be used if the widget is being managed by a Layout within a Container.
func (s *SelectEx) Resize(size fyne.Size) {
	s.BaseWidget.Resize(size)

	if s.popUp != nil {
		s.popUp.Resize(fyne.NewSize(size.Width, s.popUp.MinSize().Height))
	}
}

// SelectedIndex returns the index value of the currently selected item in Options list.
// It will return -1 if there is no selection.
func (s *SelectEx) SelectedIndex() int {
	for i, option := range s.Options {
		if s.Selected == option {
			return i
		}
	}
	return -1 // not selected/found
}

// SetSelected sets the current option of the select widget
func (s *SelectEx) SetSelected(label string) {
	for _, option := range s.Options {
		if label == option.Label() {
			s.updateSelected(option)
		}
	}
}

// SetSelectedIndex will set the Selected option from the value in Options list at index position.
func (s *SelectEx) SetSelectedIndex(index int) {
	if index < 0 || index >= len(s.Options) {
		return
	}

	s.updateSelected(s.Options[index])
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler
func (s *SelectEx) Tapped(*fyne.PointEvent) {
	if s.Disabled() {
		return
	}

	s.tapAnimation()
	s.Refresh()

	s.showPopUp()
}

// TypedKey is called if a key event happens while this Select is focused.
//
// Implements: fyne.Focusable
func (s *SelectEx) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeySpace, fyne.KeyUp, fyne.KeyDown:
		s.showPopUp()
	case fyne.KeyRight:
		i := s.SelectedIndex() + 1
		if i >= len(s.Options) {
			i = 0
		}
		s.SetSelectedIndex(i)
	case fyne.KeyLeft:
		i := s.SelectedIndex() - 1
		if i < 0 {
			i = len(s.Options) - 1
		}
		s.SetSelectedIndex(i)
	}
}

// TypedRune is called if a text event happens while this Select is focused.
//
// Implements: fyne.Focusable
func (s *SelectEx) TypedRune(_ rune) {
	// intentionally left blank
}

func (s *SelectEx) popUpPos() fyne.Position {
	buttonPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(s.super())
	return buttonPos.Add(fyne.NewPos(0, s.Size().Height-theme.InputBorderSize()))
}

func (s *SelectEx) showPopUp() {
	items := make([]*fyne.MenuItem, len(s.Options))
	for i := range s.Options {
		opt := s.Options[i] // capture
		items[i] = fyne.NewMenuItem(opt.Label(), func() {
			s.updateSelected(opt)
			s.popUp = nil
		})
	}

	c := fyne.CurrentApp().Driver().CanvasForObject(s.super())
	s.popUp = NewPopUpMenuEx(fyne.NewMenu("", items...), c, s.searchBoxHeight)
	s.popUp.alignment = s.Alignment
	s.popUp.ShowAtPosition(s.popUpPos())
	s.popUp.Resize(fyne.NewSize(s.Size().Width, s.popUp.MinSize().Height))
}

func (s *SelectEx) tapAnimation() {
	if s.tapAnim == nil {
		return
	}
	s.tapAnim.Stop()
	s.tapAnim.Start()
}

func (s *SelectEx) updateSelected(opt SelectOption) {
	if s.binder == nil {
		return
	}
	err := s.binder.Set(opt.Value())
	if err != nil {
		fyne.LogError("Error setting current data value", err)
		return
	}
	s.Selected = opt

	if s.OnChanged != nil {
		s.OnChanged(s.Selected)
	}

	s.Refresh()
}

type selectRendererEx struct {
	icon             *Icon
	label            *RichText
	background, line *canvas.Rectangle

	objects []fyne.CanvasObject
	combo   *SelectEx
}

func (s *selectRendererEx) Objects() []fyne.CanvasObject {
	return s.objects
}

func (s *selectRendererEx) Destroy() {}

// Layout the components of the button widget
func (s *selectRendererEx) Layout(size fyne.Size) {
	s.line.Resize(fyne.NewSize(size.Width, theme.InputBorderSize()))
	s.line.Move(fyne.NewPos(0, size.Height-theme.InputBorderSize()))
	s.background.Resize(fyne.NewSize(size.Width, size.Height-theme.InputBorderSize()*2))
	s.background.Move(fyne.NewPos(0, theme.InputBorderSize()))
	s.label.inset = fyne.NewSize(theme.Padding(), theme.Padding())

	iconPos := fyne.NewPos(size.Width-theme.IconInlineSize()-theme.Padding()*2, (size.Height-theme.IconInlineSize())/2)
	labelSize := fyne.NewSize(iconPos.X-theme.Padding(), s.label.MinSize().Height)

	s.label.Resize(labelSize)
	s.label.Move(fyne.NewPos(theme.Padding(), (size.Height-labelSize.Height)/2))

	s.icon.Resize(fyne.NewSize(theme.IconInlineSize(), theme.IconInlineSize()))
	s.icon.Move(iconPos)
}

// MinSize calculates the minimum size of a select button.
// This is based on the selected text, the drop icon and a standard amount of padding added.
func (s *selectRendererEx) MinSize() fyne.Size {
	s.combo.propertyLock.RLock()
	defer s.combo.propertyLock.RUnlock()

	minPlaceholderWidth := fyne.MeasureText(s.combo.PlaceHolder, theme.TextSize(), fyne.TextStyle{}).Width
	min := s.label.MinSize()
	min.Width = minPlaceholderWidth
	min = min.Add(fyne.NewSize(theme.Padding()*6, theme.Padding()*2))
	return min.Add(fyne.NewSize(theme.IconInlineSize()+theme.Padding()*2, 0))
}

func (s *selectRendererEx) Refresh() {
	s.combo.propertyLock.RLock()
	s.updateLabel()
	s.updateIcon()
	s.background.FillColor, s.line.FillColor = s.bgLineColor()
	s.combo.propertyLock.RUnlock()

	s.Layout(s.combo.Size())
	if s.combo.popUp != nil {
		s.combo.popUp.alignment = s.combo.Alignment
		s.combo.popUp.Move(s.combo.popUpPos())
		s.combo.popUp.Resize(fyne.NewSize(s.combo.size.Width, s.combo.popUp.MinSize().Height))
		s.combo.popUp.Refresh()
	}
	s.background.Refresh()
	canvas.Refresh(s.combo.super())
}

func (s *selectRendererEx) bgLineColor() (bg color.Color, line color.Color) {
	if s.combo.Disabled() {
		return theme.InputBackgroundColor(), theme.DisabledColor()
	}
	if s.combo.focused {
		return theme.FocusColor(), theme.PrimaryColor()
	}
	if s.combo.hovered {
		return theme.HoverColor(), theme.ShadowColor()
	}
	return theme.InputBackgroundColor(), theme.ShadowColor()
}

func (s *selectRendererEx) updateIcon() {
	if s.combo.Disabled() {
		s.icon.Resource = theme.NewDisabledResource(theme.MenuDropDownIcon())
	} else {
		s.icon.Resource = theme.MenuDropDownIcon()
	}
	s.icon.Refresh()
}

func (s *selectRendererEx) updateLabel() {
	if s.combo.PlaceHolder == "" {
		s.combo.PlaceHolder = defaultPlaceHolder
	}

	s.label.Segments[0].(*TextSegment).Style.Alignment = s.combo.Alignment
	if s.combo.disabled {
		s.label.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameDisabled
	} else {
		s.label.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameForeground
	}
	if s.combo.Selected.Label() == "" {
		s.label.Segments[0].(*TextSegment).Text = s.combo.PlaceHolder
	} else {
		s.label.Segments[0].(*TextSegment).Text = s.combo.Selected.Label()
	}
	s.label.Refresh()
}

package muiebitengine

import (
	"image/color"
	_ "image/png"
	"sync"

	"github.com/gabstv/microui-go/microui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

var (
	fontmap    = make(map[microui.Font]Font)
	lastFontID int
	fontMutex  sync.RWMutex
)

type FontFaceWrapper struct {
	FontBase
	font.Face
	OffsetX       int
	OffsetY       int
	GetTextWidth  func(text string) int32
	GetTextHeight func() int32
}

func (w *FontFaceWrapper) Draw(dst *ebiten.Image, tx string, x, y int, clr color.Color) {
	text.Draw(dst, tx, w.Face, x+w.OffsetX, y+w.OffsetY, clr)
}

func (w *FontFaceWrapper) TextWidth(str string) int32 {
	if w.GetTextWidth == nil {
		r := text.BoundString(w.Face, str)
		return int32(r.Dx())
	}
	return w.GetTextWidth(str)
}

func (w *FontFaceWrapper) TextHeight() int32 {
	if w.GetTextHeight == nil {
		return int32(w.Face.Metrics().Height.Ceil())
	}
	return w.GetTextHeight()
}

type Font interface {
	Draw(dst *ebiten.Image, text string, x, y int, clr color.Color)
	TextWidth(text string) int32
	TextHeight() int32
	id() microui.Font
	setID(id microui.Font)
}

type FontBase struct {
	fid microui.Font
}

func (f *FontBase) id() microui.Font {
	return f.fid
}

func (f *FontBase) setID(id microui.Font) {
	f.fid = id
}

// UIFont returns a microui.Font from a Font interface.
// If the font is not registered, it will be registered.
// The font can be removed with RemoveFont.
func UIFont(f Font) microui.Font {
	if f == nil {
		panic("nil font")
	}
	fontMutex.Lock()
	defer fontMutex.Unlock()
	// check if exists
	if f.id() != 0 {
		if _, ok := fontmap[f.id()]; ok {
			return f.id()
		}
	}
	lastFontID++
	id := microui.Font(lastFontID)
	f.setID(id)
	fontmap[id] = f
	return id
}

func RemoveFont(f Font) {
	fontMutex.Lock()
	defer fontMutex.Unlock()
	id := f.id()
	if _, ok := fontmap[id]; ok {
		delete(fontmap, id)
	}
}

func RemoveAllFonts() {
	fontMutex.Lock()
	defer fontMutex.Unlock()
	fontmap = make(map[microui.Font]Font)
	lastFontID = 0
}

func GetFont(id microui.Font) Font {
	fontMutex.RLock()
	defer fontMutex.RUnlock()
	return fontmap[id]
}

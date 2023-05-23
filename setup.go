package muiebitengine

import (
	_ "embed"
	_ "image/png"

	"github.com/gabstv/microui-go/microui"
	ebtext "github.com/hajimehoshi/ebiten/v2/text"
)

//go:embed default_atlas.png
var defaultAtlasPNG []byte

var (
	iconClose     = microui.NewRect(0, 0, 16, 16)
	iconResize    = microui.NewRect(24, 24, 6, 6)
	iconCheck     = microui.NewRect(16, 0, 16, 16)
	iconCollapsed = microui.NewRect(32, 0, 16, 16)
	iconExpanded  = microui.NewRect(48, 0, 16, 16)
	atlasWhite    = microui.NewRect(2, 18, 3, 3)

	DefaultAtlasRects = []microui.Rect{
		{},
		iconClose,
		iconResize,
		iconCheck,
		iconCollapsed,
		iconExpanded,
		atlasWhite,
	}
)

func init() {
	const (
		cw = 6
		ch = 16
	)
	microui.DefaultGetTextWidth = func(font microui.Font, text string) int32 {
		if font == 0 {
			return cw * int32(len([]rune(text)))
		}
		ff := Font(font)
		if ff.GetTextWidth != nil {
			return ff.GetTextWidth(text)
		}
		r := ebtext.BoundString(ff, text)
		//TODO: test
		return int32(r.Dx())
	}
	microui.DefaultGetTextHeight = func(font microui.Font) int32 {
		if uintptr(font) == 0 {
			return ch
		}
		ff := Font(font)
		if ff.GetTextHeight != nil {
			return ff.GetTextHeight()
		}
		r := ebtext.BoundString(ff, "A")
		//TODO: test
		return int32(r.Dy())
	}
}

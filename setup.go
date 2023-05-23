package muiebitengine

import (
	_ "embed"
	_ "image/png"
	"sync"

	"github.com/gabstv/microui-go/microui"
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

var currentDriver Driver
var currentDriverLock sync.Mutex

func getCurrentDriver() Driver {
	currentDriverLock.Lock()
	defer currentDriverLock.Unlock()
	return currentDriver
}

func setCurrentDriver(d Driver) {
	currentDriverLock.Lock()
	defer currentDriverLock.Unlock()
	currentDriver = d
}

func init() {
	const (
		cw = 6
		ch = 16
	)
	microui.DefaultGetTextWidth = func(font microui.Font, text string) int32 {
		if font == 0 {
			d := getCurrentDriver()
			if d == nil || d.DefaultFont() == nil {
				return cw * int32(len([]rune(text)))
			}
			return d.DefaultFont().TextWidth(text)
		}
		f := GetFont(font)
		return f.TextWidth(text)
	}
	microui.DefaultGetTextHeight = func(font microui.Font) int32 {
		if font == 0 {
			d := getCurrentDriver()
			if d == nil || d.DefaultFont() == nil {
				return ch
			}
			return d.DefaultFont().TextHeight()
		}
		f := GetFont(font)
		return f.TextHeight()
	}
}

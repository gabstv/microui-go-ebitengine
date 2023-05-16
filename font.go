package muiebitengine

import (
	_ "image/png"
	"sync"

	"github.com/gabstv/microui-go/microui"
	"golang.org/x/image/font"
)

var (
	fontmap        = make(map[microui.Font]FontWrapper)
	inversefontmap = make(map[font.Face]microui.Font)
	lastFontID     int
	fontMutex      sync.RWMutex
)

type FontWrapper struct {
	font.Face
	OffsetX int
	OffsetY int
}

func FontID(ff font.Face) microui.Font {
	fontMutex.Lock()
	defer fontMutex.Unlock()
	// check if exists
	if id, ok := inversefontmap[ff]; ok {
		return id
	}
	lastFontID++
	id := microui.Font(lastFontID)
	fontmap[id] = FontWrapper{
		Face: ff,
	}
	inversefontmap[ff] = id
	return id
}

func SetFontOffsets(ff font.Face, x, y int) {
	id := FontID(ff)
	fontMutex.Lock()
	defer fontMutex.Unlock()
	fw := fontmap[id]
	fw.OffsetX = x
	fw.OffsetY = y
	fontmap[id] = fw
}

func RemoveFont(ff font.Face) {
	fontMutex.Lock()
	defer fontMutex.Unlock()
	if id, ok := inversefontmap[ff]; ok {
		delete(fontmap, id)
		delete(inversefontmap, ff)
	}
}

func RemoveAllFonts() {
	fontMutex.Lock()
	defer fontMutex.Unlock()
	fontmap = make(map[microui.Font]FontWrapper)
	inversefontmap = make(map[font.Face]microui.Font)
	lastFontID = 0
}

func Font(id microui.Font) FontWrapper {
	fontMutex.RLock()
	defer fontMutex.RUnlock()
	return fontmap[id]
}

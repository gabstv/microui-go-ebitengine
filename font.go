package muiebitengine

import (
	_ "image/png"
	"sync"

	"github.com/gabstv/microui-go/microui"
	"golang.org/x/image/font"
)

var (
	fontmap        = make(map[microui.Font]font.Face)
	inversefontmap = make(map[font.Face]microui.Font)
	lastFontID     int
	fontMutex      sync.RWMutex
)

func RegisterFont(ff font.Face) microui.Font {
	fontMutex.Lock()
	defer fontMutex.Unlock()
	// check if exists
	if id, ok := inversefontmap[ff]; ok {
		return id
	}
	lastFontID++
	id := microui.Font(lastFontID)
	fontmap[id] = ff
	inversefontmap[ff] = id
	return id
}
func DeregisterFont(ff font.Face) {
	if id, ok := inversefontmap[ff]; ok {
		delete(fontmap, id)
		delete(inversefontmap, ff)
	}
}

func Font(id microui.Font) font.Face {
	fontMutex.RLock()
	defer fontMutex.RUnlock()
	return fontmap[id]
}

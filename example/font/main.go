package main

import (
	_ "embed"
	"fmt"
	"image/color"

	ebitenui "github.com/gabstv/microui-go-ebitengine"
	"github.com/gabstv/microui-go/demo"
	mu "github.com/gabstv/microui-go/microui"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// font from (CC0):
// https://managore.itch.io/m5x7

//go:embed m5x7.ttf
var customTTF []byte

var (
	normalFont font.Face
)

func main() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("microui-go + ebitengine")
	ctx := mu.NewContext()
	g := &game{
		uidriver: ebitenui.New(ctx, ebitenui.WithDefaultFont(normalFont), ebitenui.WithDefaultFontOffset(0, 12)),
		ctx:      ctx,
	}
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}

type game struct {
	uidriver ebitenui.Driver
	ctx      *mu.Context
}

func (g *game) Update() error {
	g.uidriver.UpdateInputs()
	g.ctx.Begin()
	demo.DemoWindow(g.ctx)
	demo.LogWindow(g.ctx)
	demo.StyleWindow(g.ctx)
	// to use a custom font apart from the default, you can do:
	// g.ctx.PushStyleFont(ebitenui.FontID(fontFace))
	// g.ctx.PopStyle()
	g.ctx.End()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	bgc := demo.BackgroundColor()
	cc := color.RGBA{
		R: bgc.R,
		G: bgc.G,
		B: bgc.B,
		A: bgc.A,
	}
	screen.Fill(cc)
	g.uidriver.Draw(screen)
}

func (g *game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func init() {
	tt, err := opentype.Parse(customTTF)
	if err != nil {
		panic(fmt.Errorf("failed to parse custom font: %w", err))
	}

	const dpi = 72
	normalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create custom font face: %w", err))
	}
}

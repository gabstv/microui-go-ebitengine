package main

import (
	"image/color"

	ebitenui "github.com/gabstv/microui-go-ebitengine"
	"github.com/gabstv/microui-go/demo"
	mu "github.com/gabstv/microui-go/microui"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("microui-go + ebitengine")
	ctx := mu.NewContext()
	g := &game{
		uidriver: ebitenui.New(ctx),
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

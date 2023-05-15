package muiebitengine

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"strings"

	"github.com/gabstv/microui-go/microui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Driver interface {
	UpdateInputs()
	Draw(screen *ebiten.Image)
}

type renderer struct {
	options  driverOptions
	ctx      *microui.Context
	commands []drawCommand
	charbuf  []rune
}

type DriverOption func(*driverOptions)

type driverOptions struct {
	AtlasTexture      *ebiten.Image
	AtlasRects        []microui.Rect
	ScrollMultiplierX float64
	ScrollMultiplierY float64
}

func WithAtlasTexture(tex *ebiten.Image) DriverOption {
	return func(o *driverOptions) {
		o.AtlasTexture = tex
	}
}

func WithScrollMultiplier(x, y float64) DriverOption {
	return func(o *driverOptions) {
		o.ScrollMultiplierX = x
		o.ScrollMultiplierY = y
	}
}

func New(ctx *microui.Context, options ...DriverOption) Driver {
	opts := &driverOptions{
		AtlasRects:        DefaultAtlasRects,
		ScrollMultiplierX: 1.0,
		ScrollMultiplierY: -30.0,
	}
	for _, o := range options {
		o(opts)
	}
	r := &renderer{
		options:  *opts,
		ctx:      ctx,
		commands: make([]drawCommand, 0, 4096),
		charbuf:  make([]rune, 1024),
	}
	r.atlasSetup()
	ctx.SetRenderCommand(r.render)
	ctx.SetBeginRender(r.beginFrame)
	ctx.SetEndRender(func() {})
	// ctx.SetEndRender(func() {
	// 	rl.EndScissorMode()
	// })
	return r
}

func (r *renderer) UpdateInputs() {
	x, y := ebiten.CursorPosition()
	r.ctx.InputMouseMove(int32(x), int32(y))

	swx, swy := ebiten.Wheel()
	if swx != 0 || swy != 0 {
		r.ctx.InputScroll(int32(swx*r.options.ScrollMultiplierX), int32(swy*r.options.ScrollMultiplierY))
	}
	var mbtns microui.MouseButton
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mbtns |= microui.MouseLeft
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		mbtns |= microui.MouseMiddle
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		mbtns |= microui.MouseRight
	}
	if mbtns != 0 {
		x, y := ebiten.CursorPosition()
		r.ctx.InputMouseDown(int32(x), int32(y), mbtns)
	}
	var mbtnsUp microui.MouseButton
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		mbtnsUp |= microui.MouseLeft
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) {
		mbtnsUp |= microui.MouseMiddle
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		mbtnsUp |= microui.MouseRight
	}
	if mbtnsUp != 0 {
		x, y := ebiten.CursorPosition()
		r.ctx.InputMouseUp(int32(x), int32(y), mbtnsUp)
	}
	r.charbuf = ebiten.AppendInputChars(r.charbuf)
	if len(r.charbuf) > 0 {
		r.ctx.InputText(string(r.charbuf))
		r.charbuf = r.charbuf[:0]
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyControl) {
		r.ctx.InputKeyDown(microui.KeyCtrl)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyShift) {
		r.ctx.InputKeyDown(microui.KeyShift)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyAlt) {
		r.ctx.InputKeyDown(microui.KeyAlt)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		r.ctx.InputKeyDown(microui.KeyBackspace)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		r.ctx.InputKeyDown(microui.KeyReturn)
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyControl) {
		r.ctx.InputKeyUp(microui.KeyCtrl)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyShift) {
		r.ctx.InputKeyUp(microui.KeyShift)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyAlt) {
		r.ctx.InputKeyUp(microui.KeyAlt)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyBackspace) {
		r.ctx.InputKeyUp(microui.KeyBackspace)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyEnter) {
		r.ctx.InputKeyUp(microui.KeyReturn)
	}
}

func (r *renderer) Draw(screen *ebiten.Image) {
	r.ctx.Render()
	currentClip := screen
	for _, cmd := range r.commands {
		switch cmd.Type {
		case commandTypeText:
			r.drawText(currentClip, cmd.Text)
		case commandTypeRect:
			r.drawRect(currentClip, cmd.Rect)
		case commandTypeIcon:
			r.drawIcon(currentClip, cmd.Icon)
		case commandTypeClip:
			currentClip = r.drawClip(screen, cmd.Clip)
		}
	}
	r.commands = r.commands[:0]
}

func (r *renderer) drawText(clip *ebiten.Image, cmd textCommand) {
	if cmd.Font == nil {
		// draw using default ebiten font
		ebitenutil.DebugPrintAt(clip, cmd.Text, int(cmd.Pos.X), int(cmd.Pos.Y))
		return
	}
	//TODO: draw using github.com/hajimehoshi/ebiten/v2/text
}

func (r *renderer) drawRect(clip *ebiten.Image, cmd rectCommand) {
	rect := convertRect(r.options.AtlasRects[6])
	subim := r.options.AtlasTexture.SubImage(rect).(*ebiten.Image)
	geom := ebiten.GeoM{}
	geom.Scale(float64(cmd.Rect.W)/float64(rect.Dx()), float64(cmd.Rect.H)/float64(rect.Dy()))
	geom.Translate(float64(cmd.Rect.X), float64(cmd.Rect.Y))
	cscale := ebiten.ColorScale{}
	cscale.ScaleWithColor(convertColor(cmd.Color))
	clip.DrawImage(subim, &ebiten.DrawImageOptions{
		GeoM:       geom,
		ColorScale: cscale,
	})
}

func (r *renderer) drawIcon(clip *ebiten.Image, cmd iconCommand) {
	iconimg := r.options.AtlasTexture.SubImage(convertRect(r.options.AtlasRects[cmd.ID])).(*ebiten.Image)
	geom := ebiten.GeoM{}
	geom.Translate(float64(cmd.Rect.X), float64(cmd.Rect.Y))
	cscale := ebiten.ColorScale{}
	cscale.ScaleWithColor(convertColor(cmd.Color))
	clip.DrawImage(iconimg, &ebiten.DrawImageOptions{
		GeoM:       geom,
		ColorScale: cscale,
	})
}

func (r *renderer) drawClip(screen *ebiten.Image, cmd clipCommand) *ebiten.Image {
	if cmd.Rect.W == 0 || cmd.Rect.H == 0 {
		return nil
	}
	return screen.SubImage(convertRect(cmd.Rect)).(*ebiten.Image)
}

func (r *renderer) atlasSetup() {
	if r.options.AtlasTexture != nil {
		return
	}
	img, _, err := image.Decode(bytes.NewReader(defaultAtlasPNG))
	if err != nil {
		panic(fmt.Errorf("failed to decode default atlas: %w", err))
	}
	r.options.AtlasTexture = ebiten.NewImageFromImage(img)
}

func (r *renderer) beginFrame() {
	if len(r.commands) > 0 {
		r.commands = r.commands[:0]
	}
}

func (r *renderer) render(cmd *microui.Command) {
	switch cmd.Type() {
	case microui.CommandText:
		r.putText(cmd.Text())
	case microui.CommandRect:
		r.putRect(cmd.Rect())
	case microui.CommandIcon:
		r.putIcon(cmd.Icon())
	case microui.CommandClip:
		r.putClip(cmd.Clip())
	}
}

func (r *renderer) putText(cmd microui.TextCommand) {
	r.commands = append(r.commands, drawCommand{
		Type: commandTypeText,
		Text: textCommand{
			Font:  cmd.Font(),
			Pos:   cmd.Pos(),
			Color: cmd.Color(),
			Text:  ztstr(cmd.Text()),
		},
	})
}

func (r *renderer) putRect(cmd microui.RectCommand) {
	r.commands = append(r.commands, drawCommand{
		Type: commandTypeRect,
		Rect: rectCommand{
			Rect:  cmd.Rect(),
			Color: cmd.Color(),
		},
	})
}

func (r *renderer) putIcon(cmd microui.IconCommand) {
	r.commands = append(r.commands, drawCommand{
		Type: commandTypeIcon,
		Icon: iconCommand{
			ID:    cmd.ID(),
			Rect:  cmd.Rect(),
			Color: cmd.Color(),
		},
	})
}

func (r *renderer) putClip(cmd microui.ClipCommand) {
	r.commands = append(r.commands, drawCommand{
		Type: commandTypeClip,
		Clip: clipCommand{
			Rect: cmd.Rect(),
		},
	})
}

type drawCommand struct {
	Type commandType
	Text textCommand
	Rect rectCommand
	Icon iconCommand
	Clip clipCommand
}

type commandType int

const (
	commandTypeText commandType = iota
	commandTypeRect
	commandTypeIcon
	commandTypeClip
)

type textCommand struct {
	Font  microui.Font
	Pos   microui.Vec2
	Color microui.Color
	Text  string
}

type rectCommand struct {
	Rect  microui.Rect
	Color microui.Color
}

type iconCommand struct {
	ID    int32
	Rect  microui.Rect
	Color microui.Color
}

type clipCommand struct {
	Rect microui.Rect
}

func convertRect(r microui.Rect) image.Rectangle {
	return image.Rect(int(r.X), int(r.Y), int(r.X+r.W), int(r.Y+r.H))
}

func convertColor(c microui.Color) color.Color {
	return color.RGBA{
		R: c.R,
		G: c.G,
		B: c.B,
		A: c.A,
	}
}

func ztstr(v string) string {
	p := strings.IndexByte(v, 0)
	if p == -1 {
		return v
	}
	return v[:p]
}

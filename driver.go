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
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type Driver interface {
	UpdateInputs()
	Draw(screen *ebiten.Image)
}

type driver struct {
	options  driverOptions
	ctx      *microui.Context
	commands []drawCommand
	charbuf  []rune
}

type DriverOption func(*driverOptions)

type driverOptions struct {
	AtlasTexture       *ebiten.Image
	AtlasRects         []microui.Rect
	ScrollMultiplierX  float64
	ScrollMultiplierY  float64
	DefaultFont        font.Face
	DefaultFontOffsetX int
	DefaultFontOffsetY int
}

func WithAtlasTexture(tex *ebiten.Image) DriverOption {
	return func(o *driverOptions) {
		o.AtlasTexture = tex
	}
}

func WithDefaultFont(ff font.Face) DriverOption {
	return func(o *driverOptions) {
		o.DefaultFont = ff
	}
}

func WithDefaultFontOffset(x, y int) DriverOption {
	return func(o *driverOptions) {
		o.DefaultFontOffsetX = x
		o.DefaultFontOffsetY = y
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
	r := &driver{
		options:  *opts,
		ctx:      ctx,
		commands: make([]drawCommand, 0, 4096),
		charbuf:  make([]rune, 1024),
	}
	r.atlasSetup()
	ctx.SetRenderCommand(r.render)
	ctx.SetBeginRender(r.beginFrame)
	ctx.SetEndRender(func() {})
	return r
}

func (d *driver) UpdateInputs() {
	x, y := ebiten.CursorPosition()
	d.ctx.InputMouseMove(int32(x), int32(y))

	swx, swy := ebiten.Wheel()
	if swx != 0 || swy != 0 {
		d.ctx.InputScroll(int32(swx*d.options.ScrollMultiplierX), int32(swy*d.options.ScrollMultiplierY))
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
		d.ctx.InputMouseDown(int32(x), int32(y), mbtns)
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
		d.ctx.InputMouseUp(int32(x), int32(y), mbtnsUp)
	}
	d.charbuf = ebiten.AppendInputChars(d.charbuf)
	if len(d.charbuf) > 0 {
		d.ctx.InputText(string(d.charbuf))
		d.charbuf = d.charbuf[:0]
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyControl) {
		d.ctx.InputKeyDown(microui.KeyCtrl)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyShift) {
		d.ctx.InputKeyDown(microui.KeyShift)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyAlt) {
		d.ctx.InputKeyDown(microui.KeyAlt)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		d.ctx.InputKeyDown(microui.KeyBackspace)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		d.ctx.InputKeyDown(microui.KeyReturn)
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyControl) {
		d.ctx.InputKeyUp(microui.KeyCtrl)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyShift) {
		d.ctx.InputKeyUp(microui.KeyShift)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyAlt) {
		d.ctx.InputKeyUp(microui.KeyAlt)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyBackspace) {
		d.ctx.InputKeyUp(microui.KeyBackspace)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyEnter) {
		d.ctx.InputKeyUp(microui.KeyReturn)
	}
}

func (d *driver) Draw(screen *ebiten.Image) {
	d.ctx.Render()
	currentClip := screen
	for _, cmd := range d.commands {
		switch cmd.Type {
		case commandTypeText:
			d.drawText(currentClip, cmd.Text)
		case commandTypeRect:
			d.drawRect(currentClip, cmd.Rect)
		case commandTypeIcon:
			d.drawIcon(currentClip, cmd.Icon)
		case commandTypeClip:
			currentClip = d.drawClip(screen, cmd.Clip)
		}
	}
	d.commands = d.commands[:0]
}

func (d *driver) drawText(clip *ebiten.Image, cmd textCommand) {
	if clip == nil {
		return
	}
	if cmd.Font == 0 && d.options.DefaultFont == nil {
		// draw using default ebiten font
		// the drawback is that the color cannot be changed
		ebitenutil.DebugPrintAt(clip, cmd.Text, int(cmd.Pos.X), int(cmd.Pos.Y))
		return
	}
	var ff font.Face
	var offx, offy int
	if cmd.Font == 0 {
		ff = d.options.DefaultFont
		offx = d.options.DefaultFontOffsetX
		offy = d.options.DefaultFontOffsetY
	} else {
		w := Font(cmd.Font)
		ff = w.Face
		offx = w.OffsetX
		offy = w.OffsetY
	}
	// b := text.BoundString(ff, cmd.Text)
	c := convertColor(cmd.Color)
	text.Draw(clip, cmd.Text, ff, int(cmd.Pos.X)+offx, int(cmd.Pos.Y)+offy, c)
}

func (d *driver) drawRect(clip *ebiten.Image, cmd rectCommand) {
	if clip == nil {
		return
	}
	rect := convertRect(d.options.AtlasRects[6])
	subim := d.options.AtlasTexture.SubImage(rect).(*ebiten.Image)
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

func (d *driver) drawIcon(clip *ebiten.Image, cmd iconCommand) {
	if clip == nil {
		return
	}
	iconimg := d.options.AtlasTexture.SubImage(convertRect(d.options.AtlasRects[cmd.ID])).(*ebiten.Image)
	geom := ebiten.GeoM{}
	geom.Translate(float64(cmd.Rect.X), float64(cmd.Rect.Y))
	cscale := ebiten.ColorScale{}
	cscale.ScaleWithColor(convertColor(cmd.Color))
	clip.DrawImage(iconimg, &ebiten.DrawImageOptions{
		GeoM:       geom,
		ColorScale: cscale,
	})
}

func (d *driver) drawClip(screen *ebiten.Image, cmd clipCommand) *ebiten.Image {
	if cmd.Rect.W == 0 || cmd.Rect.H == 0 {
		return nil
	}
	return screen.SubImage(convertRect(cmd.Rect)).(*ebiten.Image)
}

func (d *driver) atlasSetup() {
	if d.options.AtlasTexture != nil {
		return
	}
	img, _, err := image.Decode(bytes.NewReader(defaultAtlasPNG))
	if err != nil {
		panic(fmt.Errorf("failed to decode default atlas: %w", err))
	}
	d.options.AtlasTexture = ebiten.NewImageFromImage(img)
}

func (d *driver) beginFrame() {
	if len(d.commands) > 0 {
		d.commands = d.commands[:0]
	}
}

func (d *driver) render(cmd *microui.Command) {
	switch cmd.Type() {
	case microui.CommandText:
		d.putText(cmd.Text())
	case microui.CommandRect:
		d.putRect(cmd.Rect())
	case microui.CommandIcon:
		d.putIcon(cmd.Icon())
	case microui.CommandClip:
		d.putClip(cmd.Clip())
	}
}

func (d *driver) putText(cmd microui.TextCommand) {
	d.commands = append(d.commands, drawCommand{
		Type: commandTypeText,
		Text: textCommand{
			Font:  cmd.Font(),
			Pos:   cmd.Pos(),
			Color: cmd.Color(),
			Text:  ztstr(cmd.Text()),
		},
	})
}

func (d *driver) putRect(cmd microui.RectCommand) {
	d.commands = append(d.commands, drawCommand{
		Type: commandTypeRect,
		Rect: rectCommand{
			Rect:  cmd.Rect(),
			Color: cmd.Color(),
		},
	})
}

func (d *driver) putIcon(cmd microui.IconCommand) {
	d.commands = append(d.commands, drawCommand{
		Type: commandTypeIcon,
		Icon: iconCommand{
			ID:    cmd.ID(),
			Rect:  cmd.Rect(),
			Color: cmd.Color(),
		},
	})
}

func (d *driver) putClip(cmd microui.ClipCommand) {
	d.commands = append(d.commands, drawCommand{
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

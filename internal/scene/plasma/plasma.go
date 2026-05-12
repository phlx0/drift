package plasma

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

type Plasma struct {
	w, h  int
	theme scene.Theme
	time  float64

	cfgSpeed float64
	cfgScale float64
}

func New(cfg config.PlasmaConfig) *Plasma {
	return &Plasma{
		cfgSpeed: cfg.Speed,
		cfgScale: cfg.Scale,
	}
}

func (p *Plasma) Name() string { return "plasma" }

func (p *Plasma) Init(w, h int, t scene.Theme) {
	p.w, p.h = w, h
	p.theme = t
	p.time = 0
}

func (p *Plasma) Resize(w, h int) {
	p.w, p.h = w, h
}

func (p *Plasma) Update(dt float64) {
	p.time += dt * p.cfgSpeed
}

func (p *Plasma) value(cx, cy int) float64 {
	x := float64(cx) / float64(p.w) * 12.0 * p.cfgScale
	y := float64(cy) / float64(p.h) * 6.0 * p.cfgScale

	t := p.time

	mx := 6.0 * p.cfgScale * (0.5 + 0.35*math.Sin(t*0.37))
	my := 3.0 * p.cfgScale * (0.5 + 0.35*math.Cos(t*0.29))
	dx := x - mx
	dy := y - my

	v1 := math.Sin(x + t)
	v2 := math.Sin(y + t*1.1)
	v3 := math.Sin((x+y)*0.6 + t*0.8)
	v4 := math.Sin(math.Sqrt(dx*dx+dy*dy)*1.5 + t*1.2)

	return (v1 + v2 + v3 + v4 + 4.0) / 8.0 // [0, 1]
}

func (p *Plasma) color(v float64) scene.RGBColor {
	n := len(p.theme.Palette)
	scaled := v * float64(n)
	i := int(scaled) % n
	frac := scaled - math.Floor(scaled)
	col := scene.Lerp(p.theme.Palette[i], p.theme.Palette[(i+1)%n], frac)
	if v > 0.85 {
		col = scene.Lerp(col, p.theme.Bright, (v-0.85)/0.15)
	} else if v < 0.15 {
		col = scene.Lerp(p.theme.Dim[i%len(p.theme.Dim)], col, v/0.15)
	}
	return col
}

func densityChar(v float64) rune {
	switch {
	case v < 0.20:
		return ' '
	case v < 0.40:
		return '░'
	case v < 0.60:
		return '▒'
	case v < 0.80:
		return '▓'
	default:
		return '█'
	}
}

func (p *Plasma) Draw(screen tcell.Screen) {
	if p.w < 10 || p.h < 3 {
		return
	}
	for cy := 0; cy < p.h; cy++ {
		for cx := 0; cx < p.w; cx++ {
			v := p.value(cx, cy)
			screen.SetContent(cx, cy, densityChar(v), nil, p.color(v).Style())
		}
	}
}

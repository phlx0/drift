package starfield

import (
	"math"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

var sfGlyphs = []rune{'·', '∘', '*', '✦'}

type Starfield struct {
	w, h  int
	theme scene.Theme
	stars []sfStar
	rng   *rand.Rand

	cfgCount int
	cfgSpeed float64
}

type sfStar struct {
	x, y    float64
	z       float64 // 1 = far, approaches 0 = at viewer
	prevPX  int
	prevPY  int
	hasPrev bool
	palIdx  int
}

func New(cfg config.StarfieldConfig) *Starfield {
	return &Starfield{
		cfgCount: cfg.Count,
		cfgSpeed: cfg.Speed,
	}
}

func (s *Starfield) Name() string { return "starfield" }

func (s *Starfield) Init(w, h int, t scene.Theme) {
	s.w, s.h = w, h
	s.theme = t
	s.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	s.stars = make([]sfStar, s.cfgCount)
	for i := range s.stars {
		s.stars[i] = s.spawnStar(true)
	}
}

func (s *Starfield) Resize(w, h int) {
	s.w, s.h = w, h
	for i := range s.stars {
		s.stars[i] = s.spawnStar(true)
	}
}

func (s *Starfield) spawnStar(scattered bool) sfStar {
	x := s.rng.Float64()*2 - 1
	y := s.rng.Float64()*2 - 1
	z := 0.0
	if scattered {
		z = 0.02 + s.rng.Float64()*0.98
	} else {
		z = 0.75 + s.rng.Float64()*0.25
	}
	return sfStar{
		x:      x,
		y:      y,
		z:      z,
		palIdx: s.rng.Intn(len(s.theme.Palette)),
	}
}

func (s *Starfield) sfProject(st sfStar) (px, py int, ok bool) {
	fx := float64(s.w)*0.5 + st.x/st.z*float64(s.w)*0.5
	fy := float64(s.h)*0.5 + st.y/st.z*float64(s.h)*0.5
	px = int(fx + 0.5)
	py = int(fy + 0.5)
	ok = px >= 0 && px < s.w && py >= 0 && py < s.h
	return
}

func (s *Starfield) Update(dt float64) {
	speed := s.cfgSpeed * 0.3

	for i := range s.stars {
		st := &s.stars[i]

		if px, py, ok := s.sfProject(*st); ok {
			st.prevPX, st.prevPY, st.hasPrev = px, py, true
		} else {
			st.hasPrev = false
		}

		st.z -= speed * dt

		if st.z <= 0.01 {
			s.stars[i] = s.spawnStar(false)
			continue
		}
		if _, _, ok := s.sfProject(*st); !ok {
			s.stars[i] = s.spawnStar(false)
		}
	}
}

func (s *Starfield) Draw(screen tcell.Screen) {
	for _, st := range s.stars {
		px, py, ok := s.sfProject(st)
		if !ok {
			continue
		}

		brightness := math.Pow(1.0-st.z, 1.5)

		palColor := s.theme.Palette[st.palIdx%len(s.theme.Palette)]
		dimColor := s.theme.Dim[st.palIdx%len(s.theme.Dim)]

		if st.hasPrev && (st.prevPX != px || st.prevPY != py) {
			if st.prevPX >= 0 && st.prevPX < s.w && st.prevPY >= 0 && st.prevPY < s.h {
				trailColor := scene.Lerp(dimColor, palColor, brightness*0.35)
				screen.SetContent(st.prevPX, st.prevPY, '·', nil, trailColor.Style())
			}
		}

		glyphIdx := int(brightness * float64(len(sfGlyphs)))
		if glyphIdx >= len(sfGlyphs) {
			glyphIdx = len(sfGlyphs) - 1
		}

		var color scene.RGBColor
		if brightness > 0.85 {
			color = scene.Lerp(palColor, s.theme.Bright, (brightness-0.85)/0.15)
		} else {
			color = scene.Lerp(dimColor, palColor, brightness/0.85)
		}

		screen.SetContent(px, py, sfGlyphs[glyphIdx], nil, color.Style())
	}
}

package bonsai

import (
	"math"
	"math/rand"
	"time"

	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
)

type bonsaiState int

const (
	bonsaiGrowing bonsaiState = iota
	bonsaiHolding
	bonsaiFading
)

const (
	angleUp   = -math.Pi / 2
	aimSpread = 1.15
	maxDepth  = 4
	maxTips   = 6
	stepX     = 1.4
	stepY     = 0.85
	maxPetals = 22
)

type segment struct {
	x, y  int
	ch    rune
	color scene.RGBColor
}

type tip struct {
	x, y     float64
	angle    float64
	aim      float64
	lean     float64
	life     int
	baseLife int
	depth    int
}

type petal struct {
	x, y  float64
	fall  float64
	sway  float64
	phase float64
	ch    rune
	color scene.RGBColor
}

type Bonsai struct {
	w, h    int
	groundY int
	theme   scene.Theme
	rng     *rand.Rand

	segs   []segment
	pot    []segment
	tips   []tip
	leaves []segment
	petals []petal

	fadeList []segment
	visible  int

	maxSegs      int
	branchesLeft int
	branchesMade int
	state        bonsaiState
	growTimer    float64
	petalTimer   float64
	stateTimer   float64

	cfgPauseSeconds float64
	cfgFadeSeconds  float64
	cfgSpeed        float64
}

func New(cfg config.BonsaiConfig) *Bonsai {
	return &Bonsai{
		cfgPauseSeconds: cfg.PauseSeconds,
		cfgFadeSeconds:  cfg.FadeSeconds,
		cfgSpeed:        cfg.Speed,
	}
}

func (b *Bonsai) Name() string { return "bonsai" }

func (b *Bonsai) Init(w, h int, t scene.Theme) {
	b.w, b.h = w, h
	b.theme = t
	b.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	b.reset()
}

func (b *Bonsai) Resize(w, h int) {
	b.w, b.h = w, h
	b.reset()
}

func (b *Bonsai) reset() {
	b.segs = nil
	b.pot = nil
	b.tips = nil
	b.leaves = nil
	b.petals = nil
	b.fadeList = nil
	b.visible = 0
	b.state = bonsaiGrowing
	b.growTimer = 0
	b.petalTimer = 0
	b.stateTimer = 0
	b.maxSegs = b.w * b.h / 3
	b.branchesLeft = 9 + b.w*b.h/300
	b.branchesMade = 0

	if b.w < 16 || b.h < 8 {
		b.groundY = b.h - 1
		return
	}

	b.groundY = b.buildPot()

	life := int(float64(b.h)*0.42) + 3
	lean := 0.28 + b.rng.Float64()*0.30
	if b.rng.Intn(2) == 0 {
		lean = -lean
	}
	startX := b.w/2 + b.rng.Intn(b.w/8+1) - b.w/16

	t := tip{
		x:        float64(startX),
		y:        float64(b.groundY),
		lean:     lean,
		life:     life,
		baseLife: life,
	}
	t.aim = b.trunkAim(&t)
	t.angle = t.aim
	b.tips = []tip{t}

	b.stampTrunkBase(startX, b.groundY)
}

func (b *Bonsai) trunkAim(t *tip) float64 {
	prog := 1 - float64(t.life)/float64(t.baseLife)
	return clampAim(angleUp + t.lean*math.Cos(prog*math.Pi*1.4))
}

func (b *Bonsai) Update(dt float64) {
	switch b.state {
	case bonsaiGrowing:
		b.growTimer += dt
		stepDur := 1.0 / (16.0 * b.cfgSpeed)
		for b.growTimer >= stepDur && b.state == bonsaiGrowing {
			b.growTimer -= stepDur
			b.step()
		}

	case bonsaiHolding:
		b.stateTimer += dt
		b.updatePetals(dt)
		if b.stateTimer >= b.cfgPauseSeconds {
			b.startFade()
		}

	case bonsaiFading:
		b.stateTimer += dt
		b.updatePetals(dt)
		if len(b.fadeList) == 0 || b.cfgFadeSeconds <= 0 {
			b.reset()
			return
		}
		left := 1.0 - b.stateTimer/b.cfgFadeSeconds
		b.visible = int(scene.Clamp64(left, 0, 1) * float64(len(b.fadeList)))
		if b.visible <= 0 {
			b.reset()
		}
	}
}

func (b *Bonsai) startFade() {
	b.fadeList = make([]segment, 0, len(b.segs)+len(b.pot))
	b.fadeList = append(b.fadeList, b.segs...)
	b.fadeList = append(b.fadeList, b.pot...)
	b.rng.Shuffle(len(b.fadeList), func(i, j int) {
		b.fadeList[i], b.fadeList[j] = b.fadeList[j], b.fadeList[i]
	})
	b.visible = len(b.fadeList)
	b.state = bonsaiFading
	b.stateTimer = 0
}

func (b *Bonsai) step() {
	if len(b.tips) == 0 {
		b.state = bonsaiHolding
		b.stateTimer = 0
		return
	}

	next := make([]tip, 0, len(b.tips)+2)
	for i := range b.tips {
		t := b.tips[i]
		if !b.advance(&t) {
			continue
		}
		projected := len(next) + len(b.tips) - i
		if child, ok := b.maybeBranch(&t, projected); ok {
			next = append(next, child)
		}
		next = append(next, t)
	}
	b.tips = next
}

func (b *Bonsai) advance(t *tip) bool {
	px, py := int(math.Round(t.x)), int(math.Round(t.y))

	if t.depth == 0 {
		t.aim = b.trunkAim(t)
		t.angle += (b.rng.Float64()*2 - 1) * 0.14
	} else {
		t.angle += (b.rng.Float64()*2 - 1) * 0.30
	}
	t.angle += (t.aim - t.angle) * 0.3

	reach := stepX
	rise := stepY
	if t.depth > 0 {
		reach *= 1.25
		rise *= 0.9
	}
	t.x += math.Cos(t.angle) * reach
	t.y += math.Sin(t.angle) * rise

	cx, cy := int(math.Round(t.x)), int(math.Round(t.y))
	b.stamp(px, py, cx, cy, t)

	t.life--
	if t.life <= 0 {
		b.growPad(cx, cy, t.depth)
		return false
	}
	if cx < 0 || cx >= b.w || cy < 0 || cy > b.groundY {
		return false
	}
	return len(b.segs) < b.maxSegs
}

func (b *Bonsai) maybeBranch(t *tip, liveTips int) (tip, bool) {
	if t.depth >= maxDepth || len(b.segs) >= b.maxSegs {
		return tip{}, false
	}
	if b.branchesLeft <= 0 || liveTips >= maxTips {
		return tip{}, false
	}
	chance := 0.34 - float64(t.depth)*0.06
	if t.depth == 0 && b.branchesMade == 0 && float64(t.life) <= float64(t.baseLife)*0.6 {
		chance = 1
	}
	if t.depth == 0 && t.life > t.baseLife-2 {
		chance = 0
	}
	if b.rng.Float64() >= chance {
		return tip{}, false
	}

	life := int(float64(t.life)*0.75) + 2
	if t.depth == 0 {
		life = int(float64(t.baseLife)*0.55) + 2
	}
	if life < 3 {
		return tip{}, false
	}

	spread := 0.75 + b.rng.Float64()*0.5
	if t.depth > 0 {
		spread = 0.35 + b.rng.Float64()*0.45
	}
	if b.rng.Intn(2) == 0 {
		spread = -spread
	}
	aim := clampAim(t.aim + spread)
	b.branchesLeft--
	b.branchesMade++

	return tip{
		x:        t.x,
		y:        t.y,
		angle:    aim,
		aim:      aim,
		life:     life,
		baseLife: life,
		depth:    t.depth + 1,
	}, true
}

func clampAim(a float64) float64 {
	if a < angleUp-aimSpread {
		return angleUp - aimSpread
	}
	if a > angleUp+aimSpread {
		return angleUp + aimSpread
	}
	return a
}

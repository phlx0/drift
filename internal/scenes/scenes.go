package scenes

import (
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/scene"
	"github.com/phlx0/drift/internal/scene/clock"
	"github.com/phlx0/drift/internal/scene/constellation"
	"github.com/phlx0/drift/internal/scene/life"
	"github.com/phlx0/drift/internal/scene/maze"
	"github.com/phlx0/drift/internal/scene/orrery"
	"github.com/phlx0/drift/internal/scene/particles"
	"github.com/phlx0/drift/internal/scene/pipes"
	"github.com/phlx0/drift/internal/scene/rain"
	"github.com/phlx0/drift/internal/scene/starfield"
	"github.com/phlx0/drift/internal/scene/waveform"
)

func All(cfg config.SceneConfig) []scene.Scene {
	return []scene.Scene{
		constellation.New(cfg.Constellation),
		rain.New(cfg.Rain),
		particles.New(cfg.Particles),
		waveform.New(cfg.Waveform),
		orrery.New(cfg.Orrery),
		pipes.New(cfg.Pipes),
		maze.New(cfg.Maze),
		life.New(cfg.Life),
		clock.New(cfg.Clock),
		starfield.New(cfg.Starfield),
	}
}

func ByName(name string, cfg config.SceneConfig) scene.Scene {
	for _, s := range All(cfg) {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

func Names() []string {
	all := All(config.Default().Scene)
	names := make([]string, len(all))
	for i, s := range all {
		names[i] = s.Name()
	}
	return names
}

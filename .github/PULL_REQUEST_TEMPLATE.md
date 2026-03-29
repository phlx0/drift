## Summary

<!-- What does this PR do? One clear paragraph. -->

## Type of change

- [ ] Bug fix
- [ ] New scene
- [ ] New theme
- [ ] Configuration / CLI change
- [ ] Documentation
- [ ] Other

## Checklist

- [ ] Branch name follows `<type>/<short-description>` convention (e.g. `feat/aurora-scene`)
- [ ] Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/) format
- [ ] `CHANGELOG.md` updated under `## [Unreleased]`
- [ ] `make test` passes
- [ ] `go vet ./...` passes

**For new scenes:**
- [ ] Scene lives in `internal/scene/myscene/` as `package myscene` with a `New(cfg)` constructor
- [ ] Registered in `internal/scenes/scenes.go`
- [ ] Implements the full `scene.Scene` interface (`Name`, `Init`, `Update`, `Draw`, `Resize`)
- [ ] Uses `dt` for all time-based motion — no frame-counting
- [ ] Respects the theme — uses `t.Palette`, `t.Dim`, `t.Bright`; no hardcoded colors
- [ ] Never calls `screen.Show()` inside `Draw`
- [ ] Added config struct to `config.go` with sensible defaults
- [ ] Tested visually across at least two themes
- [ ] Demo GIF added to `demo/`

**For new themes:**
- [ ] Added to `Themes` map in `internal/scene/scene.go`
- [ ] Theme name added to comment in `internal/config/config.go`
- [ ] README theme list updated
- [ ] Tested visually with at least two scenes

## Screenshots / recording *(required for new scenes and themes)*

<!-- Attach a terminal recording or screenshot. -->
<!-- Tip: use vhs github.com/charmbracelet/vhs ;) -->
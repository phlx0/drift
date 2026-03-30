# Changelog

All notable changes to drift will be documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
drift uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

- **config validation** — `Load()` now rejects out-of-range values for all numeric config fields with a clear multi-line error listing every invalid field
- **shell snippet tests** — unit tests covering `shellSnippet()` for all supported shells and unsupported-shell error paths

### Fixed

- **orrery** scene — terminal resize no longer resets the RNG to the same seed used at init, preventing repetitive asteroid and UFO spawn patterns after resize events

---

## [0.10.0] — 2026-03-29

### Added

- **starfield** scene — classic 3-D star warp; stars spawn near the centre, accelerate toward the viewer and fan out to the edges; close stars flash bright and leave a one-cell trail; configurable via `[scene.starfield]`: `count`, `speed`

### Changed

- **scene package structure** — each scene now lives in its own subdirectory and Go package (`internal/scene/rain`, `internal/scene/orrery`, etc.); the orrery is split across `orrery.go`, `bodies.go`, `effects.go`, and `render.go`; scene registration moved to a new `internal/scenes` package to avoid circular imports
- **contributing guide and PR template** updated to reflect the new scene package structure and registration steps

---

## [0.9.0] — 2026-03-29

### Added

- **orrery** scene — a stylized solar system with a fixed sun, eight planets, and concentric orbit rings; configurable via `[scene.orrery]`: `bodies`, `trail_decay`
- **showcase quit keys** — `q` and `ctrl+c` now exit showcase mode in addition to `esc`, fixing terminals (e.g. Ghostty) that intercept the escape key (closes #34)

---

## [0.8.2] — 2026-03-27

### Fixed

- `drift config` now clearly reports when the config file is missing (with a hint to run `drift config --init`) and when it fails to parse (previously both cases silently used defaults with no indication)

---

## [0.8.1] — 2026-03-27

### Fixed

- `drift version` now shows the correct version, commit, and date when installed via `go install` — previously always showed `dev / none / unknown` because the binary had no ldflags applied

---

## [0.8.0] — 2026-03-27

### Added

- **showcase mode** — `drift --showcase` runs drift continuously without exiting on input; navigate scenes with `↑`/`↓` or `w`/`s`, cycle themes with `←`/`→` or `a`/`d`, press `esc` to quit. A two-row HUD overlay shows the current scene and theme name for 3 seconds after any navigation keypress, then fades out automatically.

---

## [0.7.0] — 2026-03-26

### Added

- **clock** scene — current time rendered as large braille digits (5×7 pixel font, 3×2 char cells per digit) centred on screen, styled in the active theme; shows the date below in dim color; configurable via `[scene.clock]`: `show_date`
- **AUR packaging** — `drift-bin` (pre-built binary) and `drift-git` (builds from HEAD) packages published to the Arch User Repository; install with `yay -S drift-bin` or `yay -S drift-git`
- **AUR release automation** — GitHub Actions workflow (`.github/workflows/aur.yml`) automatically updates `drift-bin` PKGBUILD and `.SRCINFO` on every release, computing new checksums and pushing to AUR
- **AUR badge** added to README
- **AUR install option** added to README installation section

---

## [0.6.1] — 2026-03-22

### Added

- **tmux status toggle** — hides the tmux status bar while drift is active and restores it on exit; opt-in via `hide_tmux_status = true` in `[engine]`

---

## [0.6.0] — 2026-03-21

### Added
- **rosepine** theme — Rosé Pine colorscheme; palette uses Love, Iris, Foam, and Pine accents with the base background layers as dim variants
- **automatic PR labeling** — labels applied automatically on PRs based on changed file paths and branch name prefix via `actions/labeler`
- **platform and PRs Welcome badges** added to README
- **scene and theme issue templates** added for structured contributions

### Changed

- CONTRIBUTING.md requires a `CHANGELOG.md` entry for every user-visible change; `make fmt` and `make lint` documented as the standard commands; project structure updated to include all scenes
- GitHub PR template expanded with per-type checklists for scenes and themes
- `make fmt` target added to Makefile (`gofmt -s -w ./...`); demo targets updated to include all scenes
- `gofmt` and `misspell` linters added to `.golangci.yml`
- README logo centered correctly using `<pre>` tag instead of fenced code block

### Fixed

- `gofmt -s` formatting applied across all source files

### Refactored

- `wallChar` in maze scene replaced 16-case switch with a bitmask lookup table
- `Draw` in waveform scene extracted `buildBrailleCell` helper to reduce complexity
- `Run` in engine extracted `handleEvent` and `handleTick` helpers to reduce complexity

---

## [0.5.0] — 2026-03-21

### Added

- **life** scene — Conway's Game of Life on a toroidal grid; newborn cells flash bright, age through the theme palette, then fade to dim variants for visual depth. Auto-resets after `reset_seconds` or after 3 seconds of stagnation. Configurable via `[scene.life]`: `density`, `speed`, `reset_seconds`
- **OLED pixel shift** — the engine nudges the entire rendered image by one cell every 10 seconds, cycling through a 3×3 grid (90-second full cycle). Reduces burn-in risk on OLED displays; all scenes benefit automatically with no per-scene changes. Closes #14

---

## [0.4.1] — 2026-03-21

### Added

- **pipes config**: `[scene.pipes]` now configurable — `heads`, `turn_chance`, `speed`, `reset_seconds`
- **maze config**: `[scene.maze]` now configurable — `pause_seconds`, `fade_seconds`, `speed`

### Fixed

- `drift config --init` template was missing `[scene.pipes]` and `[scene.maze]` sections
- `wildberries` theme was missing from the theme comment in the default config

---

## [0.4.0] — 2026-03-21

### Fixed

- **Scene config**: settings under `[scene.waveform]`, `[scene.rain]`, `[scene.constellation]`, and `[scene.particles]` in `config.toml` were parsed but never applied — scene constructors now receive and use their config values

---

## [0.3.1] — 2026-03-20

### Changed

- **CLI**: `--duration` flag now uses cobra's `Changed()` API instead of a `-1` sentinel to detect explicit user input; flag default changed from `-1` to `0`

---

## [0.3.0] — 2026-03-20

### Added

- **pipes** scene — box-drawing pipes snake across the screen in theme colors, wrapping at edges and resetting every 45 seconds
- **maze** scene — a perfect maze builds itself via depth-first search, rendered as thin box-drawing lines; holds on completion, then dissolves and regenerates

---

## [0.2.0] — 2026-03-20

### Fixed

- **bash shell integration**: terminal no longer freezes when drift activates.
  Background subshells spawned with job control enabled are placed in their own
  process group; bash would then send `SIGTTOU` to the process when drift called
  `tcsetattr`, suspending it and locking the terminal. The timer subshell is now
  spawned with `set +m` so it stays in the shell's process group. A foreground
  guard (`ps` stat check) was also added to skip activation if a command is
  running when the timer fires.

### Added

- `CODE_OF_CONDUCT.md` — Contributor Covenant 2.1
- `SECURITY.md` — private vulnerability reporting via GitHub Security Advisories
- `.github/CODEOWNERS` — all PRs auto-assigned to @phlx0
- `.github/dependabot.yml` — weekly automated updates for GitHub Actions and Go modules
- CI: `go mod tidy` check to catch uncommitted go.sum drift
- Release: `go test -race` step runs before GoReleaser to gate broken releases

### Changed

- `CONTRIBUTING.md` — expanded with Conventional Commits format, branch naming conventions, and type reference table
- CI: push trigger narrowed to `main` only (PRs remain covered by `pull_request`)
- CI: lint job now derives Go version from `go.mod` instead of a hardcoded value
- CI: `golangci-lint` pinned to `v1.64.0` instead of floating `latest`
- CI: coverage upload targets Go `1.24` (latest in matrix)
- Release: `goreleaser-action` pinned to `v6.3.0`

---

## [0.1.0] — 2026-03-19

First public release.

### Scenes

- **constellation** — stars drift slowly across the screen, connecting nearby neighbours with dotted lines; brightness twinkles per star
- **rain** — columns of half-width katakana characters and digits fall at varying speeds with bright heads and fading trails
- **particles** — a sinusoidal flow field drives 120 glyphs across the screen, leaving ghost trails as they move
- **waveform** — three layered sine waves rendered with Unicode braille characters for sub-character precision; amplitudes breathe in and out independently

### Themes

Seven built-in color themes matched to popular terminal colorschemes: `cosmic`, `nord`, `dracula`, `catppuccin`, `gruvbox`, `forest`, `mono`

### Shell integration

Idle detection via native shell mechanisms — no background daemons:

- **zsh** — TMOUT + TRAPALRM
- **bash** — PROMPT_COMMAND with a background timer
- **fish** — `fish_prompt` / `fish_preexec` event hooks

Activate with `eval "$(drift shell-init zsh)"` (or bash/fish).

### CLI

- `drift --scene <name>` — lock to a specific scene
- `drift --theme <name>` — override the color theme
- `drift --duration <n>` — seconds per scene when cycling, 0 = no cycling
- `drift list scenes` — list available scenes
- `drift list themes` — list themes with live color swatches
- `drift config --init` — write default config to `~/.config/drift/config.toml`
- `drift shell-init zsh|bash|fish` — print shell integration snippet

### Distribution

- Single static binary, no CGO, no runtime dependencies
- Pre-built releases for macOS and Linux (amd64 + arm64)
- goreleaser pipeline with SHA-256 checksums

[0.10.0]: https://github.com/phlx0/drift/releases/tag/v0.10.0
[0.9.0]: https://github.com/phlx0/drift/releases/tag/v0.9.0
[0.8.2]: https://github.com/phlx0/drift/releases/tag/v0.8.2
[0.8.1]: https://github.com/phlx0/drift/releases/tag/v0.8.1
[0.8.0]: https://github.com/phlx0/drift/releases/tag/v0.8.0
[0.7.0]: https://github.com/phlx0/drift/releases/tag/v0.7.0
[0.6.1]: https://github.com/phlx0/drift/releases/tag/v0.6.1
[0.6.0]: https://github.com/phlx0/drift/releases/tag/v0.6.0
[0.5.0]: https://github.com/phlx0/drift/releases/tag/v0.5.0
[0.4.1]: https://github.com/phlx0/drift/releases/tag/v0.4.1
[0.4.0]: https://github.com/phlx0/drift/releases/tag/v0.4.0
[0.3.1]: https://github.com/phlx0/drift/releases/tag/v0.3.1
[0.3.0]: https://github.com/phlx0/drift/releases/tag/v0.3.0
[0.2.0]: https://github.com/phlx0/drift/releases/tag/v0.2.0
[0.1.0]: https://github.com/phlx0/drift/releases/tag/v0.1.0

# Contributing to drift

Thank you for considering a contribution to drift. Here's everything you need to get started.

---

## Table of contents

- [Development setup](#development-setup)
- [Project structure](#project-structure)
- [Branches and commits](#branches-and-commits)
- [Adding a new scene](#adding-a-new-scene)
- [Adding a new theme](#adding-a-new-theme)
- [Code style](#code-style)
- [Testing](#testing)
- [Submitting a pull request](#submitting-a-pull-request)
- [Reporting bugs](#reporting-bugs)

---

## Development setup

```bash
# Clone and enter the repo
git clone https://github.com/phlx0/drift
cd drift

# Install dependencies and tidy go.sum
make setup

# Build
make build

# Run
./drift

# Run a specific scene for quick iteration
./drift --scene rain --theme dracula
```

You will need **Go 1.23 or later**.
The only external runtime dependency is `tcell/v2` — no C library required.

---

## Project structure

```
drift/
├── main.go                         Entry point, ldflags injection
├── cmd/drift/
│   ├── root.go                     CLI commands (cobra)
│   └── shell_snippets.go           Shell integration strings
├── internal/
│   ├── config/config.go            TOML config loading
│   ├── engine/engine.go            Render loop, scene lifecycle
│   ├── scene/
│   │   ├── scene.go                Scene interface, Theme, shared types and helpers
│   │   ├── constellation/          Drifting stars with connection lines
│   │   ├── rain/                   Falling character rain
│   │   ├── particles/              Flow-field particle system
│   │   ├── waveform/               Braille sine wave layers
│   │   ├── pipes/                  Box-drawing pipes
│   │   ├── maze/                   Recursive backtracker maze
│   │   ├── life/                   Conway's Game of Life
│   │   ├── clock/                  Braille large-digit clock
│   │   ├── starfield/              3-D star warp
│   │   └── orrery/                 Solar system (split across orrery.go, bodies.go, effects.go, render.go)
│   └── scenes/
│       └── scenes.go               Scene registry — All(), ByName(), Names()
└── .github/workflows/              CI and release automation
```

---

## Branches and commits

This project follows [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

### Commit messages

Format: `<type>(<scope>): <description>`

The `scope` is optional but encouraged. Keep the description short and in the imperative mood ("add", not "added" or "adds").

**Types:**

| Type | When to use |
|------|-------------|
| `feat` | A new feature or scene or theme |
| `fix` | A bug fix |
| `docs` | Documentation only changes |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test` | Adding or updating tests |
| `chore` | Build process, CI, dependency updates |
| `perf` | Performance improvements |

**Examples:**

```
feat(scene): add aurora scene
feat(theme): add tokyo-night theme
fix(rain): prevent panic on zero-width terminal
docs: improve contributing guide
refactor(engine): simplify render loop timing
test(scene): add theme palette length assertions
chore(ci): add windows to test matrix
```

For breaking changes, append `!` after the type and add a `BREAKING CHANGE:` footer:

```
feat!: remove --fps flag in favour of config file

BREAKING CHANGE: --fps CLI flag is no longer supported, set engine.fps in the config file instead.
```

### Branch names

Use the format `<type>/<short-description>`, matching the commit type:

```
feat/aurora-scene
feat/tokyo-night-theme
fix/rain-zero-width-panic
docs/contributing-conventions
chore/ci-windows-matrix
```

Branch off `main`. Do not commit directly to `main`.

---

## Adding a new scene

1. Create a directory `internal/scene/myscene/` and add `myscene.go` with `package myscene`.

2. Implement the `scene.Scene` interface:

   ```go
   import "github.com/phlx0/drift/internal/scene"

   type MyScene struct { ... }

   func New(cfg config.MySceneConfig) *MyScene { ... }

   func (s *MyScene) Name() string                      { return "myscene" }
   func (s *MyScene) Init(w, h int, t scene.Theme)      { ... }
   func (s *MyScene) Update(dt float64)                 { ... }
   func (s *MyScene) Draw(screen tcell.Screen)          { ... }
   func (s *MyScene) Resize(w, h int)                   { ... }
   ```

3. Register it in `All()` inside `internal/scenes/scenes.go`:

   ```go
   import "github.com/phlx0/drift/internal/scene/myscene"

   func All(cfg config.SceneConfig) []scene.Scene {
       return []scene.Scene{
           // existing scenes ...
           myscene.New(cfg.MyScene),
       }
   }
   ```

4. Add a config struct to `internal/config/config.go` if the scene has tunable knobs, and wire it to `SceneConfig`.

5. Add an entry to `CHANGELOG.md` under `## [Unreleased]` describing the new scene.

6. Test it:
   ```bash
   go build . && ./drift --scene myscene
   ```

### Scene guidelines

- **`Init` must be idempotent** — it is called again on every `Resize`.
- **`Draw` must not call `screen.Show()`** — the engine flushes once per frame.
- **Delta time**: `Update(dt float64)` receives seconds since last frame, capped at 100 ms by the engine. Use `dt` for all time-based motion.
- **Respect terminal color** — always use `tcell.StyleDefault` as the base and only override the foreground. Never hardcode a background color.
- **Handle all terminal sizes gracefully**, including very narrow (< 40 columns) or very short (< 10 rows) terminals.

---

## Adding a new theme

Open `internal/scene/scene.go` and add an entry to the `Themes` map:

```go
"mytheme": {
    Name: "mytheme",
    Palette: []RGBColor{
        {R, G, B},
        {R, G, B},
        {R, G, B},
        {R, G, B},
    },
    Dim: []RGBColor{
        // Darker / more muted versions of each Palette color.
        {R, G, B},
        {R, G, B},
        {R, G, B},
        {R, G, B},
    },
    Bright: RGBColor{R, G, B}, // near-white highlight
},
```

Also update the theme comment in `internal/config/config.go` to include your theme name in the inline list.

Add an entry to `CHANGELOG.md` under `## [Unreleased]` describing the new theme.

Run `./drift list themes` to confirm it appears.

---

## Code style

- Standard `gofmt -s` formatting — run `make fmt` before committing. CI will fail if any file is not formatted.
- No external linters beyond `go vet` are required, but PRs must pass the CI lint step.
- Keep files focused. If a scene file grows beyond ~300 lines, consider splitting helpers.
- Exported symbols need doc comments; unexported helpers are optional.

---

## Testing

```bash
make fmt        # format all Go files with gofmt -s
make lint       # run golangci-lint
make test       # unit tests with race detector
```

Because most of the interesting code is pixel-level rendering, visual smoke tests are done manually:

```bash
./drift --scene <name> --theme <name>
```

For automated tests, prefer testing pure functions (math helpers, config parsing, theme lookups) rather than trying to mock `tcell.Screen`.

---

## Submitting a pull request

1. Fork the repo and create a branch off `main` using the naming convention above.
2. Write commits following the Conventional Commits format.
3. Keep commits focused — one logical change per commit.
4. Add an entry to `CHANGELOG.md` under `## [Unreleased]` for every user-visible change.
5. Run `make test` and `go vet ./...` before opening the PR.
6. Fill in the PR description template.
7. Screenshots or terminal recordings of new scenes / visual changes are very welcome.

We review PRs as time allows. Patience is appreciated.

---

## Reporting bugs

Open an issue and include:

- Your OS and terminal emulator
- `drift version` output
- The theme and scene that triggered the bug
- What you expected vs. what happened
- A screenshot if the issue is visual

---

Made with care for the terminal community. ♥

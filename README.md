<div align="center">

<pre>
·  ·    ·        ·     ·    ·  ·
    ·      d r i f t      ·
  ·    ·        ·    ·       ·
</pre>

**A terminal screensaver that turns idle time into ambient art.**

Every OS has a screensaver. The terminal had nothing — until now.

[![Go](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Release](https://img.shields.io/github/v/release/phlx0/drift?style=flat-square&color=blueviolet)](https://github.com/phlx0/drift/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/phlx0/drift/ci.yml?style=flat-square&label=ci)](https://github.com/phlx0/drift/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/phlx0/drift?style=flat-square)](https://goreportcard.com/report/github.com/phlx0/drift)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen?style=flat-square)](CONTRIBUTING.md)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-brightgreen?style=flat-square)](https://github.com/phlx0/drift/releases)
[![AUR](https://img.shields.io/aur/version/drift-bin?style=flat-square&label=AUR&color=1793d1)](https://aur.archlinux.org/packages/drift-bin)

</div>

---

<img src="demo/waveform.gif" width="100%" />

---

## Scenes

drift ships nine animations. They cycle automatically or you can lock to one.

<table>
<tr>
<td width="50%">

**waveform** — braille sine waves that breathe

<img src="demo/waveform.gif" width="100%" />

</td>
<td width="50%">

**constellation** — stars drift and connect

<img src="demo/constellation.gif" width="100%" />

</td>
</tr>
<tr>
<td width="50%">

**rain** — katakana characters fall in columns

<img src="demo/rain.gif" width="100%" />

</td>
<td width="50%">

**particles** — a flow field of drifting glyphs

<img src="demo/particles.gif" width="100%" />

</td>
</tr>
<tr>
<td width="50%">

**pipes** — box-drawing pipes snake across the screen and wrap at the edges

<img src="demo/pipes.gif" width="100%" />

</td>
<td width="50%">

**maze** — a perfect maze builds itself, holds, then dissolves and regenerates

<img src="demo/maze.gif" width="100%" />

</td>
</tr>
<tr>
<td width="50%">

**life** — Conway's Game of Life; cells flash bright on birth, age through the palette, reset when the grid stagnates

<img src="demo/life.gif" width="100%" />

</td>
<td width="50%">

**clock** — current time in large braille digits, styled in the active theme, with the date below

<img src="demo/clock.gif" width="100%" />

</td>
</tr>
<tr>
<td width="50%">

**orrery** — a stylized solar system with a fixed sun, concentric orbit rings, and braille-rendered planets

<img src="demo/orrery.gif" width="100%" />

</td>
<td width="50%">

</td>
</tr>
</table>

## Themes

Nine built-in themes matched to popular terminal colorschemes.

`cosmic` · `nord` · `dracula` · `catppuccin` · `gruvbox` · `forest` · `wildberries` · `mono` · `rosepine`


```bash
drift list themes    # preview all themes with color swatches
```

---

## Installation

### Option 1 — AUR (Arch Linux)

```bash
yay -S drift-bin   # or: paru -S drift-bin
```

### Option 2 — Pre-built binary (no Go required)

1. Go to the [Releases](https://github.com/phlx0/drift/releases) page.
2. Download the archive for your platform:

   | OS | Chip | File |
   |---|---|---|
   | macOS | Apple Silicon (M1/M2/M3) | `drift_darwin_arm64.tar.gz` |
   | macOS | Intel | `drift_darwin_amd64.tar.gz` |
   | Linux | x86-64 | `drift_linux_amd64.tar.gz` |
   | Linux | ARM64 | `drift_linux_arm64.tar.gz` |
   | Windows | x86-64 | `drift_windows_amd64.zip` |

3. Extract and move it somewhere on your `PATH`:

   ```bash
   tar -xzf drift_darwin_arm64.tar.gz
   sudo mv drift /usr/local/bin/
   drift version
   ```

### Option 3 — Go install

```bash
go install github.com/phlx0/drift@latest
```

Make sure Go's bin directory is on your `PATH`:

```bash
# add to ~/.zshrc or ~/.bashrc
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Option 4 — Build from source

```bash
git clone https://github.com/phlx0/drift
cd drift
go install .
```

Requires **Go 1.23+**. No C compiler or CGO needed.

---

## Shell integration

Shell integration is what makes drift a real screensaver — your shell detects
idleness and launches drift automatically. Press any key to return to your prompt.

### Zsh

```zsh
# add to ~/.zshrc
export DRIFT_TIMEOUT=120   # seconds of inactivity (default: 120)
eval "$(drift shell-init zsh)"
```

### Bash

```bash
# add to ~/.bashrc
export DRIFT_TIMEOUT=120   # seconds of inactivity (default: 120)
eval "$(drift shell-init bash)"
```

### Fish

```fish
# add to ~/.config/fish/conf.d/drift.fish
set -x DRIFT_TIMEOUT 120   # seconds of inactivity (default: 120)
drift shell-init fish | source
```

### Windows

Shell integration is not available on Windows. Run `drift` directly from any terminal (Windows Terminal, PowerShell, cmd) to start it immediately, or use `drift --showcase` to browse interactively. Press `esc` to exit.

---

## Usage

```
drift                            start immediately (shell integration mode)
drift --scene waveform           lock to a specific scene
drift --theme catppuccin         override the color theme
drift --duration 30              cycle scenes every 30 seconds
drift --showcase                 browse all scenes and themes interactively

drift list scenes                list all available scenes
drift list themes                list themes with color swatches
drift shell-init zsh|bash|fish   print shell integration snippet
drift config                     show effective configuration
drift config --init              write default config to ~/.config/drift/config.toml
drift version                    print version info
```

| Flag | Default | Description |
|---|---|---|
| `--scene`, `-s` | cycle all | lock to a specific scene |
| `--theme`, `-t` | `cosmic` | color theme |
| `--fps` | `30` | target frame rate |
| `--duration` | `60` | seconds per scene, `0` = no cycling |
| `--showcase` | `false` | interactive browser: `↑↓`/`ws` scene · `←→`/`ad` theme · `esc` quit |

---

## Configuration

```bash
drift config --init   # writes ~/.config/drift/config.toml
```

```toml
[engine]
fps              = 30
cycle_seconds    = 60    # 0 = stay on one scene
scenes           = "all"
theme            = "cosmic"
shuffle          = true
hide_tmux_status = false

[scene.constellation]
star_count      = 80
connect_radius  = 0.18
twinkle         = true
max_connections = 4

[scene.rain]
charset = "ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉ0123456789"
density = 0.4
speed   = 1.0

[scene.particles]
count    = 120
gravity  = 0.0
friction = 0.98

[scene.waveform]
layers    = 3
amplitude = 0.70
speed     = 1.0

[scene.orrery]
bodies      = 8    # clamped to at least 4 for readability
trail_decay = 2.4

[scene.pipes]
heads         = 6
turn_chance   = 0.15
speed         = 1.0
reset_seconds = 45.0

[scene.maze]
pause_seconds = 3.0
fade_seconds  = 2.0
speed         = 1.0

[scene.life]
density       = 0.35
speed         = 1.0
reset_seconds = 30.0

[scene.clock]
show_date = true  # show date below the time
```

---

## Showcase mode

`drift --showcase` runs drift continuously — nothing exits it except `esc`. Use it to browse scenes and themes or leave it running on a visible window.

| Key | Action |
|---|---|
| `↑` / `w` | previous scene |
| `↓` / `s` | next scene |
| `←` / `a` | previous theme |
| `→` / `d` | next theme |
| `esc` / `q` / `ctrl+c` | quit |

A status bar shows the current scene and theme for 3 seconds after each keypress, then fades out.

---

## Troubleshooting

**Config changes have no effect**

Run `drift config` to check whether your config file is being found:

```
Config file: /Users/you/.config/drift/config.toml (not found — using defaults; run 'drift config --init' to create it)
```

If the file is missing, create it with `drift config --init`. If it exists but changes still don't apply, check for a TOML syntax error — `drift config` will print the parse error if there is one.

**drift doesn't activate automatically**

Make sure the shell integration is sourced in your rc file and that `DRIFT_TIMEOUT` (or `TMOUT` in zsh) is set:

```bash
export DRIFT_TIMEOUT=120
eval "$(drift shell-init zsh)"   # or bash / fish
```

Then open a new terminal session for it to take effect.

---

## Contributing

New scenes and themes are very welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for the development guide.

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=phlx0/drift&type=Date)](https://star-history.com/#phlx0/drift&Date)

---

<div align="center">

MIT License · made by [phlx0](https://github.com/phlx0)

*press any key to resume*

</div>

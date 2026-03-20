// Package drift implements the drift CLI.
package drift

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/phlx0/drift/internal/config"
	"github.com/phlx0/drift/internal/engine"
	"github.com/phlx0/drift/internal/scene"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

// SetBuildInfo is called from main with values injected by goreleaser ldflags.
func SetBuildInfo(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
}

func Execute() error {
	return rootCmd.Execute()
}

var (
	flagTheme    string
	flagScene    string
	flagFPS      int
	flagDuration float64
)

var rootCmd = &cobra.Command{
	Use:   "drift",
	Short: "Terminal screensaver and ambient visualiser",
	Long: `drift renders beautiful ASCII animations when your terminal is idle.

Set it up in your shell and it activates automatically after a configurable
period of inactivity — press any key to resume your session.

Shell integration:

  # Zsh — add to ~/.zshrc
  eval "$(drift shell-init zsh)"

  # Bash — add to ~/.bashrc
  eval "$(drift shell-init bash)"

  # Fish — add to ~/.config/fish/conf.d/drift.fish
  drift shell-init fish | source

Run 'drift list scenes' and 'drift list themes' to explore what's available.`,
	RunE:         runEngine,
	SilenceUsage: true, // don't print usage on error, it obscures the message
}

func init() {
	f := rootCmd.PersistentFlags()
	f.StringVarP(&flagTheme, "theme", "t", "", "color theme (overrides config)")
	f.StringVarP(&flagScene, "scene", "s", "", "lock to a specific scene (overrides config)")
	f.IntVar(&flagFPS, "fps", 0, "target frame rate in Hz (overrides config)")
	f.Float64Var(&flagDuration, "duration", -1, "seconds per scene when cycling, 0 = no cycling (overrides config)")

	rootCmd.AddCommand(shellInitCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
}

func runEngine(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
		cfg = config.Default()
	}

	if flagTheme != "" {
		cfg.Engine.Theme = flagTheme
	}

	if flagFPS > 0 {
		cfg.Engine.FPS = flagFPS
	}

	durationChanged := cmd.Flags().Changed("duration")
	if durationChanged {
		cfg.Engine.CycleSeconds = flagDuration
	}
	
	if flagScene != "" {
		cfg.Engine.Scenes = flagScene
		if !strings.Contains(flagScene, ",") {
			cfg.Engine.Shuffle = false
			if !durationChanged {
				cfg.Engine.CycleSeconds = 0
			}
		}
	}

	e := engine.New(*cfg)
	return e.Run()
}

var shellInitCmd = &cobra.Command{
	Use:       "shell-init <shell>",
	Short:     "Print shell integration code",
	Long:      "Print the shell integration snippet for the given shell.\nSource it from your rc file to enable idle activation.\n\nSupported shells: zsh, bash, fish",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"zsh", "bash", "fish"},
	RunE: func(_ *cobra.Command, args []string) error {
		snippet, err := shellSnippet(args[0])
		if err != nil {
			return err
		}
		fmt.Print(snippet)
		return nil
	},
}

func shellSnippet(shell string) (string, error) {
	switch shell {
	case "zsh":
		return zshSnippet, nil
	case "bash":
		return bashSnippet, nil
	case "fish":
		return fishSnippet, nil
	default:
		return "", fmt.Errorf("unsupported shell %q — choose from: zsh, bash, fish", shell)
	}
}

var listCmd = &cobra.Command{
	Use:   "list <scenes|themes>",
	Short: "List available scenes or themes",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		switch args[0] {
		case "scenes":
			fmt.Println("Available scenes:")
			for _, name := range scene.Names() {
				fmt.Printf("  %s\n", name)
			}
		case "themes":
			names := scene.ThemeNames()
			sort.Strings(names)
			fmt.Println("Available themes:")
			for _, name := range names {
				t := scene.Themes[name]
				swatches := make([]string, len(t.Palette))
				for i, c := range t.Palette {
					swatches[i] = colorSwatch(c)
				}
				fmt.Printf("  %-14s  %s\n", name, strings.Join(swatches, " "))
			}
		default:
			return fmt.Errorf("unknown argument %q — use 'scenes' or 'themes'", args[0])
		}
		return nil
	},
}

func colorSwatch(c scene.RGBColor) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm██\x1b[0m", c.R, c.G, c.B)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("drift %s (commit %s, built %s)\n", buildVersion, buildCommit, buildDate)
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or initialise the configuration file",
	Long: `Show the path to the active config file and print the effective configuration.

Use --init to write a default config file (will not overwrite an existing one).`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		initFlag, err := cmd.Flags().GetBool("init")
		if err != nil {
			return err
		}
		if initFlag {
			return initConfig()
		}
		return showConfig()
	},
}

func init() {
	configCmd.Flags().Bool("init", false, "write default config (no-op if file already exists)")
}

func showConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	path, _ := config.Path() //nolint:errcheck
	fmt.Printf("Config file: %s\n\n", path)
	fmt.Printf("theme:         %s\n", cfg.Engine.Theme)
	fmt.Printf("fps:           %d\n", cfg.Engine.FPS)
	fmt.Printf("cycle_seconds: %.0f\n", cfg.Engine.CycleSeconds)
	fmt.Printf("scenes:        %s\n", cfg.Engine.Scenes)
	fmt.Printf("shuffle:       %v\n", cfg.Engine.Shuffle)
	return nil
}

func initConfig() error {
	if err := config.WriteDefault(); err != nil {
		if os.IsExist(err) {
			path, _ := config.Path() //nolint:errcheck
			fmt.Printf("Config already exists at %s — not overwriting.\n", path)
			return nil
		}
		return err
	}
	path, _ := config.Path() //nolint:errcheck
	fmt.Printf("Created default config at %s\n", path)
	return nil
}

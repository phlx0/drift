package drift

import (
	"strings"
	"testing"
)

func TestShellSnippetReturnsCorrectShell(t *testing.T) {
	tests := []struct {
		shell   string
		wantSub string
	}{
		{"zsh", "TMOUT"},
		{"bash", "PROMPT_COMMAND"},
		{"fish", "fish_prompt"},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			snippet, err := shellSnippet(tt.shell)
			if err != nil {
				t.Fatalf("shellSnippet(%q) returned error: %v", tt.shell, err)
			}
			if !strings.Contains(snippet, tt.wantSub) {
				t.Errorf("shellSnippet(%q) missing expected substring %q", tt.shell, tt.wantSub)
			}
		})
	}
}

func TestShellSnippetReturnsErrorForUnsupported(t *testing.T) {
	unsupported := []string{"powershell", "nushell", "", "ZSH"}

	for _, shell := range unsupported {
		t.Run(shell, func(t *testing.T) {
			_, err := shellSnippet(shell)
			if err == nil {
				t.Errorf("shellSnippet(%q) should return error for unsupported shell", shell)
			}
		})
	}
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const zshInit = `aps() {
  case "$1" in
    init|completion|help|--help|-h)
      command aps "$@"
      ;;
    --unset)
      command aps "$@" >/dev/null || return $?
      unset AWS_PROFILE
      echo "AWS_PROFILE unset"
      ;;
    *)
      local __aps_profile
      __aps_profile="$(command aps "$@")" || return $?
      if [ -n "$__aps_profile" ]; then
        export AWS_PROFILE="$__aps_profile"
        echo "AWS_PROFILE=$__aps_profile"
      fi
      ;;
  esac
}
`

const bashInit = zshInit

const fishInit = `function aps
    switch "$argv[1]"
        case init completion help --help -h
            command aps $argv
        case --unset
            command aps $argv >/dev/null; or return $status
            set -e AWS_PROFILE
            echo "AWS_PROFILE unset"
        case '*'
            set -l __aps_profile (command aps $argv)
            set -l __aps_status $status
            if test $__aps_status -ne 0
                return $__aps_status
            end
            if test -n "$__aps_profile"
                set -gx AWS_PROFILE $__aps_profile
                echo "AWS_PROFILE=$__aps_profile"
            end
    end
end
`

const (
	markerBegin = "# >>> aps initialize >>>"
	markerEnd   = "# <<< aps initialize <<<"
)

var initInstall bool

var initCmd = &cobra.Command{
	Use:   "init [zsh|bash|fish]",
	Short: "Print shell integration code to eval in your rc file",
	Long: `Print a shell function named aps that shadows this binary. The
function passes admin subcommands (init, completion, help, ...) straight
through, turns --unset into an actual unset of AWS_PROFILE, and for
everything else captures this binary's stdout and exports it as
AWS_PROFILE.

Add one of these to your shell rc file:

    eval "$(aps init zsh)"    # ~/.zshrc
    eval "$(aps init bash)"   # ~/.bashrc
    aps init fish | source    # ~/.config/fish/config.fish

Or pass --install to append the line for you (idempotent, marked with an
"aps initialize" comment block so rerunning it is a no-op).`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"zsh", "bash", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		code, err := initCode(shell)
		if err != nil {
			return err
		}

		if !initInstall {
			fmt.Print(code)
			return nil
		}

		return installInit(shell)
	},
}

func init() {
	initCmd.Flags().BoolVar(&initInstall, "install", false, "append the eval line to your shell rc file instead of printing it")
	rootCmd.AddCommand(initCmd)
}

func initCode(shell string) (string, error) {
	switch shell {
	case "zsh":
		return zshInit, nil
	case "bash":
		return bashInit, nil
	case "fish":
		return fishInit, nil
	default:
		return "", fmt.Errorf("unsupported shell %q (want zsh, bash, or fish)", shell)
	}
}

// rcFile returns the shell rc file aps's integration line belongs in.
func rcFile(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	case "bash":
		return filepath.Join(home, ".bashrc"), nil
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell %q (want zsh, bash, or fish)", shell)
	}
}

// evalLine is the one line that belongs in the rc file; the shell function
// body itself is generated dynamically each time it's eval'd.
func evalLine(shell string) string {
	if shell == "fish" {
		return "aps init fish | source"
	}
	return fmt.Sprintf(`eval "$(aps init %s)"`, shell)
}

func installInit(shell string) error {
	path, err := rcFile(shell)
	if err != nil {
		return err
	}

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	if strings.Contains(string(existing), markerBegin) {
		fmt.Printf("aps is already wired into %s\n", path)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	block := fmt.Sprintf("\n%s\n%s\n%s\n", markerBegin, evalLine(shell), markerEnd)
	if _, err := f.WriteString(block); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	fmt.Printf("added aps shell integration to %s — restart your shell or run: source %s\n", path, path)
	return nil
}

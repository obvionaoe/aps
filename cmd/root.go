package cmd

import (
	"fmt"
	"os"

	"github.com/obvionaoe/aps/internal/awsprofile"
	"github.com/obvionaoe/aps/internal/tui"
	"github.com/spf13/cobra"
)

var unsetProfile bool

var rootCmd = &cobra.Command{
	Use:   "aps [profile]",
	Short: "AWS profile switcher",
	Long: `aps prints an AWS profile name to stdout: the one given as an
argument, or one chosen from a fuzzy-searchable list when no argument is
given.

aps itself can't change your shell's environment - it only prints a
profile name. Wire it up with:

    eval "$(aps init zsh)"   # or bash / fish

in your shell rc file. That defines a shell function named aps which
exports (or, with --unset, unsets) AWS_PROFILE using this binary's
output.`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if unsetProfile {
			if len(args) == 1 {
				return fmt.Errorf("--unset can't be combined with a profile argument")
			}
			return nil
		}

		profiles, err := awsprofile.List()
		if err != nil {
			return err
		}
		if len(profiles) == 0 {
			configPath, credentialsPath, err := awsprofile.Paths()
			if err != nil {
				return err
			}
			return fmt.Errorf("no AWS profiles found in %s or %s", configPath, credentialsPath)
		}

		if len(args) == 1 {
			name := args[0]
			found := false
			for _, p := range profiles {
				if p == name {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("no such AWS profile: %q", name)
			}
			fmt.Println(name)
			return nil
		}

		selected, err := tui.Pick(profiles, os.Getenv("AWS_PROFILE"))
		if err != nil {
			return err
		}
		if selected != "" {
			fmt.Println(selected)
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().BoolVar(&unsetProfile, "unset", false, "unset AWS_PROFILE (applied by the shell function from `aps init`)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

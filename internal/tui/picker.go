// Package tui provides the interactive profile picker.
package tui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// Pick renders a fuzzy-filterable list of profiles and returns the chosen
// one. If current is non-empty and present in profiles, it's marked in the
// list. Pick returns an empty string, nil error if the user aborted (esc/
// ctrl-c).
func Pick(profiles []string, current string) (string, error) {
	if len(profiles) == 0 {
		return "", fmt.Errorf("no AWS profiles found")
	}

	options := make([]huh.Option[string], 0, len(profiles))
	for _, p := range profiles {
		label := p
		if p == current {
			label = p + " (current)"
		}
		options = append(options, huh.NewOption(label, p))
	}

	var selected string
	field := huh.NewSelect[string]().
		Title("Select AWS profile").
		Options(options...).
		Filtering(true).
		Value(&selected)

	err := huh.NewForm(huh.NewGroup(field)).Run()
	if err != nil {
		if err == huh.ErrUserAborted {
			return "", nil
		}
		return "", err
	}
	return selected, nil
}

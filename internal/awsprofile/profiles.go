// Package awsprofile finds AWS named profiles from the standard AWS CLI
// config and credentials files.
package awsprofile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var sectionRe = regexp.MustCompile(`^\[(.+)\]\s*$`)

// List returns the sorted, deduplicated set of profile names found in
// ~/.aws/config and ~/.aws/credentials. Missing files are ignored.
func List() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving home directory: %w", err)
	}

	seen := map[string]struct{}{}
	for _, path := range []string{
		filepath.Join(home, ".aws", "config"),
		filepath.Join(home, ".aws", "credentials"),
	} {
		names, err := parseFile(path)
		if err != nil {
			continue
		}
		for _, n := range names {
			seen[n] = struct{}{}
		}
	}

	profiles := make([]string, 0, len(seen))
	for n := range seen {
		profiles = append(profiles, n)
	}
	sort.Strings(profiles)
	return profiles, nil
}

// Exists reports whether name is a known profile.
func Exists(name string) (bool, error) {
	profiles, err := List()
	if err != nil {
		return false, err
	}
	for _, p := range profiles {
		if p == name {
			return true, nil
		}
	}
	return false, nil
}

func parseFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var names []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		m := sectionRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		// ~/.aws/config prefixes non-default profile sections with
		// "profile "; ~/.aws/credentials does not.
		name := strings.TrimSpace(strings.TrimPrefix(m[1], "profile "))
		if name != "" {
			names = append(names, name)
		}
	}
	return names, scanner.Err()
}

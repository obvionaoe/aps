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

// List returns the sorted, deduplicated set of profile names found in the
// AWS config and credentials files. Locations follow the AWS CLI's own
// resolution: the AWS_CONFIG_FILE and AWS_SHARED_CREDENTIALS_FILE
// environment variables take precedence (this is how setups that keep
// these files under ~/.config/aws instead of ~/.aws point aps at them),
// falling back to ~/.aws/config and ~/.aws/credentials. Missing files are
// ignored.
func List() ([]string, error) {
	configPath, credentialsPath, err := Paths()
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	files := []struct {
		path     string
		isConfig bool
	}{
		{configPath, true},
		{credentialsPath, false},
	}
	for _, f := range files {
		names, err := parseFile(f.path, f.isConfig)
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

// Paths returns the config and credentials file paths aps reads profiles
// from, honoring AWS_CONFIG_FILE / AWS_SHARED_CREDENTIALS_FILE the same way
// List does.
func Paths() (configPath, credentialsPath string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("resolving home directory: %w", err)
	}

	configPath = os.Getenv("AWS_CONFIG_FILE")
	if configPath == "" {
		configPath = filepath.Join(home, ".aws", "config")
	}
	credentialsPath = os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	if credentialsPath == "" {
		credentialsPath = filepath.Join(home, ".aws", "credentials")
	}
	return configPath, credentialsPath, nil
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

func parseFile(path string, isConfig bool) ([]string, error) {
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
		name := profileName(strings.TrimSpace(m[1]), isConfig)
		if name != "" {
			names = append(names, name)
		}
	}
	return names, scanner.Err()
}

// profileName returns the profile name for a section header's contents, or
// an empty string if the section isn't a profile.
//
// In ~/.aws/config, profile sections are "[default]" or "[profile name]";
// other section types ("[sso-session x]", "[services x]", ...) are not
// profiles and are ignored. In ~/.aws/credentials every section is a
// profile and section names carry no prefix.
func profileName(section string, isConfig bool) string {
	if !isConfig {
		return section
	}
	if section == "default" {
		return "default"
	}
	if rest, ok := strings.CutPrefix(section, "profile "); ok {
		return strings.TrimSpace(rest)
	}
	return ""
}

package awsprofile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestListOnlyReturnsProfiles(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")
	credentialsPath := filepath.Join(dir, "credentials")

	config := `[default]
region = us-east-1

[profile dev]
sso_session = my-sso

[sso-session my-sso]
sso_start_url = https://example.com/start

[services my-services]
s3 =

[profile prod]
region = eu-west-1
`
	credentials := `[legacy]
aws_access_key_id = AKIAEXAMPLE
`
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(credentialsPath, []byte(credentials), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("AWS_CONFIG_FILE", configPath)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsPath)

	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"default", "dev", "legacy", "prod"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %v, want %v", got, want)
	}
}

func TestProfileName(t *testing.T) {
	tests := []struct {
		section  string
		isConfig bool
		want     string
	}{
		{"default", true, "default"},
		{"profile dev", true, "dev"},
		{"sso-session my-sso", true, ""},
		{"services my-services", true, ""},
		{"dev", true, ""}, // config profiles must use the "profile " prefix
		{"legacy", false, "legacy"},
		{"default", false, "default"},
	}
	for _, tt := range tests {
		if got := profileName(tt.section, tt.isConfig); got != tt.want {
			t.Errorf("profileName(%q, %v) = %q, want %q", tt.section, tt.isConfig, got, tt.want)
		}
	}
}

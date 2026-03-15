package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTokenFallsBackToLegacyCredentials(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	legacyPath := filepath.Join(home, LegacyDefaultDir, CredentialsFile)
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o700); err != nil {
		t.Fatalf("mkdir legacy dir: %v", err)
	}
	if err := os.WriteFile(legacyPath, []byte("legacy-token\n"), 0o600); err != nil {
		t.Fatalf("write legacy credentials: %v", err)
	}

	token, err := LoadToken("")
	if err != nil {
		t.Fatalf("LoadToken returned error: %v", err)
	}
	if token != "legacy-token" {
		t.Fatalf("LoadToken token = %q, want legacy-token", token)
	}
}

func TestClearTokenRemovesCurrentAndLegacyCredentials(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	currentPath := filepath.Join(home, DefaultDir, CredentialsFile)
	legacyPath := filepath.Join(home, LegacyDefaultDir, CredentialsFile)
	for _, path := range []string{currentPath, legacyPath} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatalf("mkdir credentials dir: %v", err)
		}
		if err := os.WriteFile(path, []byte("token"), 0o600); err != nil {
			t.Fatalf("write credentials: %v", err)
		}
	}

	if err := ClearToken(""); err != nil {
		t.Fatalf("ClearToken returned error: %v", err)
	}

	for _, path := range []string{currentPath, legacyPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("credentials file %q still exists, err=%v", path, err)
		}
	}
}

func TestResolveTokenPriority(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	currentPath := filepath.Join(home, DefaultDir, CredentialsFile)
	if err := os.MkdirAll(filepath.Dir(currentPath), 0o700); err != nil {
		t.Fatalf("mkdir credentials dir: %v", err)
	}
	if err := os.WriteFile(currentPath, []byte("credentials-token"), 0o600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	for _, tc := range []struct {
		name         string
		pat          string
		codeupToken  string
		yunxiaoToken string
		wantToken    string
		wantSource   string
	}{
		{
			name:         "prefer personal access token env",
			pat:          "pat-token",
			codeupToken:  "codeup-token",
			yunxiaoToken: "yunxiao-token",
			wantToken:    "pat-token",
			wantSource:   "env:CODEUP_PERSONAL_ACCESS_TOKEN",
		},
		{
			name:         "prefer codeup token over yunxiao token",
			codeupToken:  "codeup-token",
			yunxiaoToken: "yunxiao-token",
			wantToken:    "codeup-token",
			wantSource:   "env:CODEUP_TOKEN",
		},
		{
			name:         "fallback to yunxiao token",
			yunxiaoToken: "yunxiao-token",
			wantToken:    "yunxiao-token",
			wantSource:   "env:YUNXIAO_TOKEN",
		},
		{
			name:       "fallback to credentials file",
			wantToken:  "credentials-token",
			wantSource: "credentials",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CODEUP_PERSONAL_ACCESS_TOKEN", tc.pat)
			t.Setenv("CODEUP_TOKEN", tc.codeupToken)
			t.Setenv("YUNXIAO_TOKEN", tc.yunxiaoToken)

			token, source, err := ResolveToken("")
			if err != nil {
				t.Fatalf("ResolveToken returned error: %v", err)
			}
			if token != tc.wantToken || source != tc.wantSource {
				t.Fatalf("ResolveToken = (%q, %q), want (%q, %q)", token, source, tc.wantToken, tc.wantSource)
			}
		})
	}
}

func TestLoadFallsBackToLegacyConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	legacyPath := filepath.Join(home, LegacyDefaultDir, ConfigFile)
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o700); err != nil {
		t.Fatalf("mkdir legacy config dir: %v", err)
	}
	content := []byte("organization_id: legacy-org\ndomain: legacy.example.com\ndefault_repo: foo/bar\n")
	if err := os.WriteFile(legacyPath, content, 0o600); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.OrganizationID != "legacy-org" || cfg.Domain != "legacy.example.com" || cfg.DefaultRepo != "foo/bar" {
		t.Fatalf("Load returned unexpected config: %+v", cfg)
	}
}

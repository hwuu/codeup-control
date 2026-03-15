package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultDomain    = "openapi-rdc.aliyuncs.com"
	DefaultDir       = ".config/cuctl"
	LegacyDefaultDir = ".config/codeupcl"
	ConfigFile       = "config.yaml"
	CredentialsFile  = "credentials"
)

type Config struct {
	OrganizationID string `yaml:"organization_id,omitempty"`
	Domain         string `yaml:"domain,omitempty"`
	DefaultRepo    string `yaml:"default_repo,omitempty"`
}

func Dir(override string) string {
	if override != "" {
		return filepath.Dir(override)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, DefaultDir)
}

func legacyDir(override string) string {
	if override != "" {
		return filepath.Dir(override)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, LegacyDefaultDir)
}

func Path(override string) string {
	if override != "" {
		return override
	}
	return filepath.Join(Dir(""), ConfigFile)
}

func CredentialsPath(cfgOverride string) string {
	return filepath.Join(Dir(cfgOverride), CredentialsFile)
}

func Load(path string) (*Config, error) {
	p := Path(path)
	data, err := os.ReadFile(p)
	if err != nil {
		if path == "" && os.IsNotExist(err) {
			legacyPath := filepath.Join(legacyDir(""), ConfigFile)
			legacyData, legacyErr := os.ReadFile(legacyPath)
			if legacyErr == nil {
				var legacyCfg Config
				if err := yaml.Unmarshal(legacyData, &legacyCfg); err != nil {
					return nil, fmt.Errorf("解析配置文件失败: %w", err)
				}
				return &legacyCfg, nil
			}
		}
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	return &cfg, nil
}

func (c *Config) Save(path string) error {
	p := Path(path)
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := os.WriteFile(p, data, 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}

func LoadToken(cfgOverride string) (string, error) {
	p := CredentialsPath(cfgOverride)
	data, err := os.ReadFile(p)
	if err != nil {
		if cfgOverride == "" && os.IsNotExist(err) {
			legacyPath := filepath.Join(legacyDir(""), CredentialsFile)
			legacyData, legacyErr := os.ReadFile(legacyPath)
			if legacyErr == nil {
				return strings.TrimSpace(string(legacyData)), nil
			}
		}
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("读取凭证文件失败: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func SaveToken(cfgOverride, token string) error {
	p := CredentialsPath(cfgOverride)
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	if err := os.WriteFile(p, []byte(token), 0600); err != nil {
		return fmt.Errorf("写入凭证文件失败: %w", err)
	}
	return nil
}

func ClearToken(cfgOverride string) error {
	paths := []string{CredentialsPath(cfgOverride)}
	if cfgOverride == "" {
		paths = append(paths, filepath.Join(legacyDir(""), CredentialsFile))
	}

	for _, p := range paths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("清除凭证文件失败: %w", err)
		}
	}

	return nil
}

// ResolveToken returns token from env > credentials file, with source label.
func ResolveToken(cfgOverride string) (token, source string, err error) {
	if v := os.Getenv("CODEUP_PERSONAL_ACCESS_TOKEN"); v != "" {
		return v, "env:CODEUP_PERSONAL_ACCESS_TOKEN", nil
	}
	if v := os.Getenv("CODEUP_TOKEN"); v != "" {
		return v, "env:CODEUP_TOKEN", nil
	}
	if v := os.Getenv("YUNXIAO_TOKEN"); v != "" {
		return v, "env:YUNXIAO_TOKEN", nil
	}
	t, err := LoadToken(cfgOverride)
	if err != nil {
		return "", "", err
	}
	if t != "" {
		return t, "credentials", nil
	}
	return "", "", nil
}

func (c *Config) ResolveOrganizationID() (orgID, source string) {
	if v := os.Getenv("CODEUP_ORGANIZATION_ID"); v != "" {
		return v, "env:CODEUP_ORGANIZATION_ID"
	}
	if v := os.Getenv("YUNXIAO_ORGANIZATION_ID"); v != "" {
		return v, "env:YUNXIAO_ORGANIZATION_ID"
	}
	if c.OrganizationID != "" {
		return c.OrganizationID, "config"
	}
	return "", ""
}

func (c *Config) ResolveDomain() string {
	if v := os.Getenv("CODEUP_DOMAIN"); v != "" {
		return v
	}
	if v := os.Getenv("YUNXIAO_DOMAIN"); v != "" {
		return v
	}
	if c.Domain != "" {
		return c.Domain
	}
	return DefaultDomain
}

package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/render-oss/render-mcp-server/pkg/cfg"
)

const currentVersion = 1
const defaultDashboardURL = "https://dashboard.render.com"

var defaultConfigPath string

const configPathEnvKey = "RENDER_CONFIG_PATH"
const workspaceEnvKey = "RENDER_WORKSPACE"

var ErrNoWorkspace = errors.New("no workspace set. Prompt the user to select a workspace. Do NOT try to select a workspace for them, as it may be destructive")
var ErrLogin = errors.New("not authenticated; either set RENDER_API_KEY or use the `login` command")

type Config struct {
	Version   int    `yaml:"version"`
	Workspace string `yaml:"workspace"`

	APIConfig `yaml:"api"`
}

type APIConfig struct {
	APIKey       string `yaml:"api_key,omitempty"`
	ExpiresAt    int64  `yaml:"expires_at,omitempty"`
	Host         string `json:"host,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// This is used to store the workspace ID in memory if we can't access the config file.
// This has the downside of not persisting across sessions, but at least it's better than nothing.
var inMemoryWorkspaceID string

func init() {
	if workspaceID := os.Getenv(workspaceEnvKey); workspaceID != "" {
		inMemoryWorkspaceID = workspaceID
	}

	var defaultConfigBaseDir string
	defaultConfigBaseDir, err := os.UserHomeDir()
	// When launching an MCP server, we may not have a home directory. Try to find a good fallback.
	if err != nil {
		execPath, err := os.Executable()
		if err != nil {
			// We don't have a good fallback to write to, just try a temp dir
			defaultConfigBaseDir = os.TempDir()
		} else {
			defaultConfigBaseDir = filepath.Dir(execPath)
		}
	}
	defaultConfigPath = filepath.Join(defaultConfigBaseDir, ".render", "mcp-server.yaml")
}

func DefaultAPIConfig() (APIConfig, error) {
	if apiKey := os.Getenv("RENDER_API_KEY"); apiKey != "" {
		return APIConfig{
			APIKey: apiKey,
			Host:   cfg.GetHost(),
		}, nil
	}

	cfg, err := getAPIConfig()
	if err != nil {
		return APIConfig{}, err
	}
	if cfg.APIKey == "" {
		return APIConfig{}, ErrLogin
	}

	return cfg, nil
}

func DashboardURL() string {
	return defaultDashboardURL
}

func getConfigPath() string {
	if path := os.Getenv(configPathEnvKey); path != "" {
		return path
	}
	return defaultConfigPath
}

func expandPath(path string) (string, error) {
	if path == "~" || len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	return path, nil
}

func SelectWorkspace(workspaceID string) error {
	// First, try to load the config from disk and update the workspace.
	// This may fail if we're operating in an environment where we don't have disk access.
	conf, err := Load()
	if err == nil {
		conf.Workspace = workspaceID
		err = conf.Persist()
		if err == nil {
			return nil
		}
	}

	// If that fails, we'll fall back to the in memory workspace ID.
	inMemoryWorkspaceID = workspaceID

	return nil
}

func WorkspaceID() (string, error) {
	// First, try to load the config from disk.
	// If that fails, we'll fall back to the in memory workspace ID.
	//
	// We don't use the environment variable here because that's just considered the starting workspace.
	// The user may have changed workspaces since then, which would be reflected in the config file
	// and/or the in memory workspace ID.
	var workspaceID string

	cfg, err := Load()
	if err == nil && cfg.Workspace != "" {
		workspaceID = cfg.Workspace
	} else {
		workspaceID = inMemoryWorkspaceID
	}

	if workspaceID == "" {
		return "", ErrNoWorkspace
	}

	return workspaceID, nil
}

func getAPIConfig() (APIConfig, error) {
	cfg, err := Load()
	if err != nil {
		return APIConfig{}, err
	}

	return cfg.APIConfig, nil
}

func SetAPIConfig(input APIConfig) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	cfg.APIKey = input.APIKey
	cfg.Host = input.Host
	cfg.ExpiresAt = input.ExpiresAt
	cfg.RefreshToken = input.RefreshToken
	return cfg.Persist()
}

func Load() (*Config, error) {
	path, err := expandPath(getConfigPath())
	if err != nil {
		return nil, err
	}

	// Ignore the error if we can't chmod try to continue
	_ = os.Chmod(filepath.Dir(path), 0755)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Version: currentVersion}, nil
		}
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Persist() error {
	path, err := expandPath(getConfigPath())
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// Ignore the error if we can't chmod try to continue
	_ = os.Chmod(filepath.Dir(path), 0755)

	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

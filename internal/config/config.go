package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Instance struct {
	URL      string `yaml:"url"`
	TokenEnv string `yaml:"token_env"`
}

type OutputConfig struct {
	Dir     string   `yaml:"dir"`
	Formats []string `yaml:"formats"`
}

type DomainsConfig struct {
	Groups   []string `yaml:"groups"`
	Projects []string `yaml:"projects"`
}

type ProjectsConfig struct {
	Include          []string `yaml:"include"`
	Exclude          []string `yaml:"exclude"`
	IncludeSubgroups bool     `yaml:"include_subgroups"`
	IncludeArchived  bool     `yaml:"include_archived"`
}

type GroupsConfig struct {
	Include          []string `yaml:"include"`
	Exclude          []string `yaml:"exclude"`
	IncludeSubgroups bool     `yaml:"include_subgroups"`
}

type ConcurrencyConfig struct {
	Groups   int `yaml:"groups"`
	Projects int `yaml:"projects"`
}

type Config struct {
	Source      Instance          `yaml:"source"`
	Destination Instance          `yaml:"destination"`
	Groups      GroupsConfig      `yaml:"groups"`
	Projects    ProjectsConfig    `yaml:"projects"`
	Domains     DomainsConfig     `yaml:"domains"`
	Concurrency ConcurrencyConfig `yaml:"concurrency"`
	Output      OutputConfig      `yaml:"output"`
}

// Tier 1 defaults — well-defined REST, clean natural keys.
var DefaultGroupDomains = []string{
	"push_rules",
	"default_branch_name",
	"mr_settings",
	"mr_approval_settings",
	"protected_environments",
	"approval_rules",
}

var DefaultProjectDomains = []string{
	"topics",
	"environments",
	"protected_environments",
	"jira_integration",
	"pipeline_triggers",
	"deploy_keys",
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if len(c.Domains.Groups) == 0 {
		c.Domains.Groups = DefaultGroupDomains
	}
	if len(c.Domains.Projects) == 0 {
		c.Domains.Projects = DefaultProjectDomains
	}
	if c.Concurrency.Groups <= 0 {
		c.Concurrency.Groups = 5
	}
	if c.Concurrency.Projects <= 0 {
		c.Concurrency.Projects = 10
	}
	if c.Output.Dir == "" {
		c.Output.Dir = "./gl-copy-report"
	}
	if len(c.Output.Formats) == 0 {
		c.Output.Formats = []string{"terminal", "html", "json"}
	}
	if !c.Projects.IncludeSubgroups {
		c.Projects.IncludeSubgroups = true
	}
	if !c.Groups.IncludeSubgroups {
		c.Groups.IncludeSubgroups = true
	}
}

func (c *Config) validate() error {
	if c.Source.URL == "" {
		return fmt.Errorf("source.url is required")
	}
	if c.Source.TokenEnv == "" {
		return fmt.Errorf("source.token_env is required")
	}
	if c.Destination.URL == "" {
		return fmt.Errorf("destination.url is required")
	}
	if c.Destination.TokenEnv == "" {
		return fmt.Errorf("destination.token_env is required")
	}
	if len(c.Groups.Include) == 0 {
		return fmt.Errorf("at least one group must be specified under groups.include")
	}
	if os.Getenv(c.Source.TokenEnv) == "" {
		return fmt.Errorf("environment variable %q (source token) is not set", c.Source.TokenEnv)
	}
	if os.Getenv(c.Destination.TokenEnv) == "" {
		return fmt.Errorf("environment variable %q (destination token) is not set", c.Destination.TokenEnv)
	}
	return nil
}

func (i *Instance) Token() string {
	return os.Getenv(i.TokenEnv)
}

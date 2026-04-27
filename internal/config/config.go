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
	"description",
	"default_branch_name",
	"default_branch_protection",
	"mr_settings",
	"mr_approval_settings",
	"protected_environments",
	"approval_rules",
	"jira_integration",
	"badges",
	"compliance_frameworks",
	"compliance_assignments",
	"security_policy_project",
}

var DefaultProjectDomains = []string{
	"topics",
	"environments",
	"protected_environments",
	"jira_integration",
	"pipeline_triggers",
	"deploy_keys",
	"project_push_rules",
	"project_mr_approvals",
	"project_approval_rules",
	"badges",
	"project_protected_branches",
	"project_protected_tags",
}

func Load(path string) (*Config, error) {
	return LoadWithOverrides(path, "", "")
}

// LoadWithOverrides loads config and applies CLI flag overrides before validation.
// This allows -group and -project flags to satisfy the validation requirements
// even when groups.include is empty in the config file.
func LoadWithOverrides(path, groupOverride, projectOverride string) (*Config, error) {
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

	// Apply CLI overrides before defaults and validation
	if groupOverride != "" {
		cfg.Groups.Include = []string{groupOverride}
	}
	if projectOverride != "" {
		cfg.Projects.Include = []string{projectOverride}
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
	// groups.include is only required when we need to derive projects from groups.
	// If projects.include has explicit paths, or a -group/-project flag will be
	// passed at runtime, groups.include can be empty.
	if len(c.Groups.Include) == 0 && len(c.Projects.Include) == 0 {
		return fmt.Errorf("specify at least one of: groups.include or projects.include")
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

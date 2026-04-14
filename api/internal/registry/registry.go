package registry

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value.Value, err)
	}
	d.Duration = parsed
	return nil
}

type ImageRegistry struct {
	Images []ImageEntry `yaml:"images"`
}

type SSHEntry struct {
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type ImageEntry struct {
	Match        string             `yaml:"match"`
	InternalPort *int               `yaml:"internal_port,omitempty"`
	Env          map[string]string  `yaml:"env,omitempty"`
	PostStart    []ExecCommand      `yaml:"post_start,omitempty"`
	PreStop      []ExecCommand      `yaml:"pre_stop,omitempty"`
	HealthCheck  *HealthCheckConfig `yaml:"health_check,omitempty"`
	Labels       map[string]string  `yaml:"labels,omitempty"`
	SSH          *SSHEntry          `yaml:"ssh,omitempty"`
	Metadata     []MetadataItem     `yaml:"metadata,omitempty"`
	Logs         []LogSource        `yaml:"logs,omitempty"`
}

type MetadataItem struct {
	Key       string   `yaml:"key"                json:"key"`
	Label     string   `yaml:"label"              json:"label"`
	Type      string   `yaml:"type"               json:"type"`
	Value     string   `yaml:"value,omitempty"    json:"value,omitempty"`
	Input     string   `yaml:"input,omitempty"    json:"input,omitempty"`
	Required  bool     `yaml:"required,omitempty" json:"required,omitempty"`
	Options   []string `yaml:"options,omitempty"   json:"options,omitempty"`
	Variant   string   `yaml:"variant,omitempty"   json:"variant,omitempty"`
	Show      string   `yaml:"show,omitempty"      json:"show,omitempty"`
	Condition string   `yaml:"condition,omitempty" json:"condition,omitempty"`
	Icon      string   `yaml:"icon,omitempty"      json:"icon,omitempty"`
	Size      string   `yaml:"size,omitempty"      json:"size,omitempty"`
}

type ExecCommand struct {
	Command    []string `yaml:"command"`
	Label      string   `yaml:"label,omitempty"`
	Delay      Duration `yaml:"delay,omitempty"`
	Timeout    Duration `yaml:"timeout,omitempty"`
	Retries    int      `yaml:"retries,omitempty"`
	RetryDelay Duration `yaml:"retry_delay,omitempty"`
}

type HealthCheckConfig struct {
	Path     string   `yaml:"path"`
	Timeout  Duration `yaml:"timeout"`
	Interval Duration `yaml:"interval"`
}

type TemplateContext struct {
	Hostname       string
	URL            string
	Scheme         string
	Port           string
	SSHPort        string
	SSHUsername    string
	SSHPassword    string
	ContainerName  string
	TrustedProxies string
	DockerMode     string
	Network        string
	InternalPort   string
	ImageName      string
	ImageRepo      string
	ImageTag       string
	SandboxID      string
	HostSuffix     string
	TTL            string
	ExpiresAt      string
	ClientIP       string
	Meta           map[string]string
}

type LogSourceType string

const (
	LogSourceTypeDocker    LogSourceType = "docker"
	LogSourceTypeFile      LogSourceType = "file"
	LogSourceTypeLifecycle LogSourceType = "lifecycle"
)

type LogSource struct {
	Key   string        `yaml:"key"            json:"key"`
	Label string        `yaml:"label"          json:"label"`
	Type  LogSourceType `yaml:"type"           json:"type"`
	Path  string        `yaml:"path,omitempty" json:"path,omitempty"`
}

type ResolvedImage struct {
	InternalPort int
	Env          []string
	PostStart    []ExecCommand
	PreStop      []ExecCommand
	HealthCheck  *HealthCheckConfig
	Labels       map[string]string
	SSH          *SSHEntry
	Logs         []LogSource
}

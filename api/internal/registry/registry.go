package registry

import (
	"fmt"
	"text/template"
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
	Metadata     MetadataSchema     `yaml:"metadata,omitempty"`
	Logs         []LogSource        `yaml:"logs,omitempty"`
}

type MetadataSchema struct {
	Groups []MetadataGroup `yaml:"groups,omitempty" json:"groups,omitempty"`
	Items  []MetadataItem  `yaml:"items,omitempty"  json:"items"`
}

type MetadataGroup struct {
	Key         string `yaml:"key"                   json:"key"`
	Label       string `yaml:"label"                 json:"label"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type MetadataItem struct {
	Key        string          `yaml:"key"                  json:"key"`
	Label      string          `yaml:"label"                json:"label,omitempty"`
	Type       string          `yaml:"type"                 json:"type"`
	Icon       string          `yaml:"icon,omitempty"       json:"icon,omitempty"`
	Group      string          `yaml:"group,omitempty"      json:"group,omitempty"`
	Visibility *VisibilityRule `yaml:"visibility,omitempty" json:"visibility,omitempty"`

	Field   *FieldSpec   `yaml:"field,omitempty"   json:"field,omitempty"`
	Action  *ActionSpec  `yaml:"action,omitempty"  json:"action,omitempty"`
	Display *DisplaySpec `yaml:"display,omitempty" json:"display,omitempty"`
}

type VisibilityRule struct {
	Contexts  []string         `yaml:"contexts,omitempty"   json:"contexts,omitempty"`
	Condition string           `yaml:"condition,omitempty"  json:"condition,omitempty"`
	DependsOn *FieldDependency `yaml:"depends_on,omitempty" json:"dependsOn,omitempty"`
}

type FieldDependency struct {
	Field string `yaml:"field" json:"field"`
	Value string `yaml:"value" json:"value"`
}

type FieldSpec struct {
	Input       string         `yaml:"input"                 json:"input,omitempty"`
	Default     string         `yaml:"default,omitempty"     json:"default,omitempty"`
	Placeholder string         `yaml:"placeholder,omitempty" json:"placeholder,omitempty"`
	HelpText    string         `yaml:"help_text,omitempty"   json:"helpText,omitempty"`
	Required    bool           `yaml:"required,omitempty"    json:"required,omitempty"`
	ReadOnly    bool           `yaml:"read_only,omitempty"   json:"readOnly,omitempty"`
	Options     []SelectOption `yaml:"options,omitempty"     json:"options,omitempty"`
}

type SelectOption struct {
	Value string `yaml:"value" json:"value"`
	Label string `yaml:"label" json:"label"`
}

type ActionSpec struct {
	URL     string             `yaml:"url"               json:"url,omitempty"`
	Variant string             `yaml:"variant,omitempty" json:"variant,omitempty"`
	Size    string             `yaml:"size,omitempty"    json:"size,omitempty"`
	Target  string             `yaml:"target,omitempty"  json:"target,omitempty"`
	Confirm string             `yaml:"confirm,omitempty" json:"confirm,omitempty"`
	urlTmpl *template.Template `yaml:"-" json:"-"`
}

type DisplaySpec struct {
	Value    string `yaml:"value"              json:"value,omitempty"`
	Format   string `yaml:"format,omitempty"   json:"format,omitempty"`
	Copyable bool   `yaml:"copyable,omitempty" json:"copyable,omitempty"`

	valueTmpl *template.Template `yaml:"-" json:"-"`
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
	ContainerID    string
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
	Status         string
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

var ValidContexts = map[string]bool{
	"image.create":    true,
	"image.edit":      true,
	"image.card":      true,
	"sandbox.create":  true,
	"sandbox.card":    true,
	"sandbox.details": true,
}

var (
	ValidFieldInputs = map[string]bool{
		"text": true, "password": true, "number": true, "email": true, "url": true,
		"select": true, "multiselect": true, "toggle": true, "textarea": true,
	}
	ValidActionVariants = map[string]bool{"default": true, "outline": true, "destructive": true}
	ValidActionSizes    = map[string]bool{"default": true, "icon": true}
	ValidActionTargets  = map[string]bool{"_blank": true, "_self": true}
	ValidDisplayFormats = map[string]bool{"text": true, "code": true, "badge": true, "link": true, "password": true}
	ValidItemTypes      = map[string]bool{"field": true, "action": true, "display": true}
)

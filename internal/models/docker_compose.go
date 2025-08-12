package models

import (
	"fmt"
	"strings"
	"time"
)

// DockerCompose represents the structure of a docker-compose.yml file
type DockerCompose struct {
	Version  string                    `yaml:"version"`
	Services map[string]ComposeService `yaml:"services"`
	Networks map[string]ComposeNetwork `yaml:"networks,omitempty"`
	Volumes  map[string]ComposeVolume  `yaml:"volumes,omitempty"`
}

// ComposeService represents a service definition in docker-compose
type ComposeService struct {
	Build           *ComposeBuild         `yaml:"build,omitempty"`
	Image           string                `yaml:"image,omitempty"`
	Ports           []string              `yaml:"ports,omitempty"`
	Environment     []string              `yaml:"environment,omitempty"`
	EnvFile         []string              `yaml:"env_file,omitempty"`
	DependsOn       []string              `yaml:"depends_on,omitempty"`
	Networks        []string              `yaml:"networks,omitempty"`
	Volumes         []string              `yaml:"volumes,omitempty"`
	Command         string                `yaml:"command,omitempty"`
	WorkingDir      string                `yaml:"working_dir,omitempty"`
	HealthCheck     *ComposeHealthCheck   `yaml:"healthcheck,omitempty"`
	Restart         string                `yaml:"restart,omitempty"`
	Deploy          *ComposeDeploy        `yaml:"deploy,omitempty"`
	Labels          map[string]string     `yaml:"labels,omitempty"`
}

// ComposeBuild represents build configuration
type ComposeBuild struct {
	Context    string            `yaml:"context"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
	Args       map[string]string `yaml:"args,omitempty"`
	Target     string            `yaml:"target,omitempty"`
}

// ComposeHealthCheck represents health check configuration
type ComposeHealthCheck struct {
	Test        []string      `yaml:"test"`
	Interval    time.Duration `yaml:"interval,omitempty"`
	Timeout     time.Duration `yaml:"timeout,omitempty"`
	Retries     int           `yaml:"retries,omitempty"`
	StartPeriod time.Duration `yaml:"start_period,omitempty"`
	Disable     bool          `yaml:"disable,omitempty"`
}

// ComposeDeploy represents deployment configuration
type ComposeDeploy struct {
	Replicas  int                    `yaml:"replicas,omitempty"`
	Resources *ComposeResources      `yaml:"resources,omitempty"`
	Labels    map[string]string      `yaml:"labels,omitempty"`
}

// ComposeResources represents resource constraints
type ComposeResources struct {
	Limits       *ComposeResourceLimits `yaml:"limits,omitempty"`
	Reservations *ComposeResourceLimits `yaml:"reservations,omitempty"`
}

// ComposeResourceLimits represents CPU and memory limits
type ComposeResourceLimits struct {
	CPU    string `yaml:"cpus,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

// ComposeNetwork represents network configuration
type ComposeNetwork struct {
	Driver         string            `yaml:"driver,omitempty"`
	DriverOpts     map[string]string `yaml:"driver_opts,omitempty"`
	IPAM          *ComposeIPAM       `yaml:"ipam,omitempty"`
	External       bool              `yaml:"external,omitempty"`
	AttachableOpts map[string]string `yaml:"attachable_opts,omitempty"`
}

// ComposeIPAM represents IP Address Management configuration
type ComposeIPAM struct {
	Driver string                 `yaml:"driver,omitempty"`
	Config []ComposeIPAMConfig    `yaml:"config,omitempty"`
	Options map[string]string     `yaml:"options,omitempty"`
}

// ComposeIPAMConfig represents IPAM configuration
type ComposeIPAMConfig struct {
	Subnet  string `yaml:"subnet,omitempty"`
	Gateway string `yaml:"gateway,omitempty"`
}

// ComposeVolume represents volume configuration
type ComposeVolume struct {
	Driver     string            `yaml:"driver,omitempty"`
	DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	External   bool              `yaml:"external,omitempty"`
	Labels     map[string]string `yaml:"labels,omitempty"`
}

// DockerConfig represents Docker-specific configuration for a profile
type DockerConfig struct {
	ProfileID       string                 `json:"profileId"`
	BaseImages      map[string]string      `json:"baseImages"`      // service -> base image
	VolumeMappings  map[string][]string    `json:"volumeMappings"`  // service -> volume mappings
	NetworkSettings map[string]interface{} `json:"networkSettings"` // custom network config
	ResourceLimits  map[string]ResourceLimit `json:"resourceLimits"` // service -> resource limits
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

// ResourceLimit represents CPU and memory limits for a service
type ResourceLimit struct {
	CPULimit    string `json:"cpuLimit"`    // e.g., "0.5"
	MemoryLimit string `json:"memoryLimit"` // e.g., "512M"
	CPUReserve  string `json:"cpuReserve"`  // e.g., "0.1"
	MemoryReserve string `json:"memoryReserve"` // e.g., "128M"
}

// DockerComposeRequest represents a request to generate docker-compose
type DockerComposeRequest struct {
	Environment      string            `json:"environment"`      // development, staging, production
	IncludeExternal  bool              `json:"includeExternal"`  // include external services like DB
	CustomOverrides  map[string]interface{} `json:"customOverrides"`  // custom service overrides
	GenerateOverride bool              `json:"generateOverride"` // generate docker-compose.override.yml
}

// ToYAML converts DockerCompose struct to YAML string
func (dc *DockerCompose) ToYAML() string {
	var builder strings.Builder
	
	// Version
	builder.WriteString(fmt.Sprintf("version: '%s'\n\n", dc.Version))
	
	// Services
	if len(dc.Services) > 0 {
		builder.WriteString("services:\n")
		for name, service := range dc.Services {
			builder.WriteString(fmt.Sprintf("  %s:\n", name))
			builder.WriteString(service.ToYAML("    "))
		}
		builder.WriteString("\n")
	}
	
	// Networks
	if len(dc.Networks) > 0 {
		builder.WriteString("networks:\n")
		for name, network := range dc.Networks {
			builder.WriteString(fmt.Sprintf("  %s:\n", name))
			builder.WriteString(network.ToYAML("    "))
		}
		builder.WriteString("\n")
	}
	
	// Volumes
	if len(dc.Volumes) > 0 {
		builder.WriteString("volumes:\n")
		for name, volume := range dc.Volumes {
			builder.WriteString(fmt.Sprintf("  %s:\n", name))
			builder.WriteString(volume.ToYAML("    "))
		}
	}
	
	return builder.String()
}

// ToYAML converts ComposeService to YAML string with indentation
func (cs *ComposeService) ToYAML(indent string) string {
	var builder strings.Builder
	
	if cs.Build != nil {
		builder.WriteString(fmt.Sprintf("%sbuild:\n", indent))
		builder.WriteString(cs.Build.ToYAML(indent + "  "))
	} else if cs.Image != "" {
		builder.WriteString(fmt.Sprintf("%simage: %s\n", indent, cs.Image))
	}
	
	if len(cs.Ports) > 0 {
		builder.WriteString(fmt.Sprintf("%sports:\n", indent))
		for _, port := range cs.Ports {
			builder.WriteString(fmt.Sprintf("%s  - \"%s\"\n", indent, port))
		}
	}
	
	if len(cs.Environment) > 0 {
		builder.WriteString(fmt.Sprintf("%senvironment:\n", indent))
		for _, env := range cs.Environment {
			builder.WriteString(fmt.Sprintf("%s  - %s\n", indent, env))
		}
	}
	
	if len(cs.DependsOn) > 0 {
		builder.WriteString(fmt.Sprintf("%sdepends_on:\n", indent))
		for _, dep := range cs.DependsOn {
			builder.WriteString(fmt.Sprintf("%s  - %s\n", indent, dep))
		}
	}
	
	if len(cs.Networks) > 0 {
		builder.WriteString(fmt.Sprintf("%snetworks:\n", indent))
		for _, network := range cs.Networks {
			builder.WriteString(fmt.Sprintf("%s  - %s\n", indent, network))
		}
	}
	
	if len(cs.Volumes) > 0 {
		builder.WriteString(fmt.Sprintf("%svolumes:\n", indent))
		for _, volume := range cs.Volumes {
			builder.WriteString(fmt.Sprintf("%s  - %s\n", indent, volume))
		}
	}
	
	if cs.Command != "" {
		builder.WriteString(fmt.Sprintf("%scommand: %s\n", indent, cs.Command))
	}
	
	if cs.WorkingDir != "" {
		builder.WriteString(fmt.Sprintf("%sworking_dir: %s\n", indent, cs.WorkingDir))
	}
	
	if cs.Restart != "" {
		builder.WriteString(fmt.Sprintf("%srestart: %s\n", indent, cs.Restart))
	}
	
	if cs.HealthCheck != nil && !cs.HealthCheck.Disable {
		builder.WriteString(fmt.Sprintf("%shealthcheck:\n", indent))
		if len(cs.HealthCheck.Test) > 0 {
			builder.WriteString(fmt.Sprintf("%s  test: [%s]\n", indent, strings.Join(quoteStrings(cs.HealthCheck.Test), ", ")))
		}
		if cs.HealthCheck.Interval > 0 {
			builder.WriteString(fmt.Sprintf("%s  interval: %s\n", indent, cs.HealthCheck.Interval.String()))
		}
		if cs.HealthCheck.Timeout > 0 {
			builder.WriteString(fmt.Sprintf("%s  timeout: %s\n", indent, cs.HealthCheck.Timeout.String()))
		}
		if cs.HealthCheck.Retries > 0 {
			builder.WriteString(fmt.Sprintf("%s  retries: %d\n", indent, cs.HealthCheck.Retries))
		}
	}
	
	return builder.String()
}

// ToYAML converts ComposeBuild to YAML string
func (cb *ComposeBuild) ToYAML(indent string) string {
	var builder strings.Builder
	
	builder.WriteString(fmt.Sprintf("%scontext: %s\n", indent, cb.Context))
	
	if cb.Dockerfile != "" {
		builder.WriteString(fmt.Sprintf("%sdockerfile: %s\n", indent, cb.Dockerfile))
	}
	
	if len(cb.Args) > 0 {
		builder.WriteString(fmt.Sprintf("%sargs:\n", indent))
		for key, value := range cb.Args {
			builder.WriteString(fmt.Sprintf("%s  %s: %s\n", indent, key, value))
		}
	}
	
	if cb.Target != "" {
		builder.WriteString(fmt.Sprintf("%starget: %s\n", indent, cb.Target))
	}
	
	return builder.String()
}

// ToYAML converts ComposeNetwork to YAML string
func (cn *ComposeNetwork) ToYAML(indent string) string {
	var builder strings.Builder
	
	if cn.Driver != "" {
		builder.WriteString(fmt.Sprintf("%sdriver: %s\n", indent, cn.Driver))
	}
	
	if cn.External {
		builder.WriteString(fmt.Sprintf("%sexternal: true\n", indent))
	}
	
	return builder.String()
}

// ToYAML converts ComposeVolume to YAML string
func (cv *ComposeVolume) ToYAML(indent string) string {
	var builder strings.Builder
	
	if cv.Driver != "" {
		builder.WriteString(fmt.Sprintf("%sdriver: %s\n", indent, cv.Driver))
	}
	
	if cv.External {
		builder.WriteString(fmt.Sprintf("%sexternal: true\n", indent))
	}
	
	return builder.String()
}

// Helper function to quote strings for YAML arrays
func quoteStrings(strs []string) []string {
	quoted := make([]string, len(strs))
	for i, str := range strs {
		quoted[i] = fmt.Sprintf("\"%s\"", str)
	}
	return quoted
}

// Validate validates the DockerCompose configuration
func (dc *DockerCompose) Validate() error {
	if dc.Version == "" {
		return fmt.Errorf("version is required")
	}
	
	if len(dc.Services) == 0 {
		return fmt.Errorf("at least one service is required")
	}
	
	// Validate service names and dependencies
	serviceNames := make(map[string]bool)
	for name := range dc.Services {
		serviceNames[name] = true
	}
	
	for name, service := range dc.Services {
		// Check if dependencies exist
		for _, dep := range service.DependsOn {
			if !serviceNames[dep] {
				return fmt.Errorf("service %s depends on non-existent service %s", name, dep)
			}
		}
	}
	
	return nil
}
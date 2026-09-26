package services

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zechtz/vertex/internal/models"
)

// remoteRunConfiguration is the part of a run configuration file IntelliJ reads
// to attach its debugger.
type remoteRunConfiguration struct {
	XMLName       xml.Name `xml:"component"`
	Component     string   `xml:"name,attr"`
	Configuration struct {
		Name    string `xml:"name,attr"`
		Type    string `xml:"type,attr"`
		Options []struct {
			Name  string `xml:"name,attr"`
			Value string `xml:"value,attr"`
		} `xml:"option"`
	} `xml:"configuration"`
}

func (c remoteRunConfiguration) option(name string) string {
	for _, option := range c.Configuration.Options {
		if option.Name == name {
			return option.Value
		}
	}
	return ""
}

func parseRunConfiguration(t *testing.T, content string) remoteRunConfiguration {
	t.Helper()

	var config remoteRunConfiguration
	if err := xml.Unmarshal([]byte(content), &config); err != nil {
		t.Fatalf("run configuration is not valid XML: %v\n%s", err, content)
	}
	return config
}

func TestIntelliJAttachConfigTargetsTheDebugPort(t *testing.T) {
	config := parseRunConfiguration(t, intellijAttachConfig("gateway", 5007))

	if config.Component != "ProjectRunConfigurationManager" {
		t.Errorf("component = %q, want ProjectRunConfigurationManager", config.Component)
	}
	if config.Configuration.Type != "Remote" {
		t.Errorf("configuration type = %q, want Remote (IntelliJ's Remote JVM Debug)", config.Configuration.Type)
	}
	if config.Configuration.Name != "Attach gateway" {
		t.Errorf("configuration name = %q, want %q", config.Configuration.Name, "Attach gateway")
	}
	for option, want := range map[string]string{
		"HOST":        "localhost",
		"PORT":        "5007",
		"SERVER_MODE": "false", // attach to the JVM, rather than wait for it to connect
	} {
		if got := config.option(option); got != want {
			t.Errorf("option %s = %q, want %q", option, got, want)
		}
	}
}

// The name sits in an XML attribute, so a quote or ampersand in a service name
// must not break the file.
func TestIntelliJAttachConfigEscapesTheServiceName(t *testing.T) {
	config := parseRunConfiguration(t, intellijAttachConfig(`auth "v2" & co`, 5005))

	if want := `Attach auth "v2" & co`; config.Configuration.Name != want {
		t.Errorf("configuration name = %q, want %q", config.Configuration.Name, want)
	}
}

func TestWriteIntelliJAttachConfig(t *testing.T) {
	sm := newTestManager(t)
	projectsDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(projectsDir, "gateway"), 0o755); err != nil {
		t.Fatal(err)
	}
	sm.services["gw"] = &models.Service{ID: "gw", Name: "api/gateway", Dir: "gateway", DebugPort: 5005}

	path, err := sm.WriteIntelliJAttachConfig("gw", projectsDir)
	if err != nil {
		t.Fatalf("WriteIntelliJAttachConfig: %v", err)
	}

	// The slash in the name must not become a directory.
	if want := filepath.Join(projectsDir, "gateway", ".run", "Attach api-gateway.run.xml"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("configuration was not written: %v", err)
	}
	if got := parseRunConfiguration(t, string(content)).option("PORT"); got != "5005" {
		t.Errorf("written PORT = %q, want 5005", got)
	}

	// Writing again follows a change of port.
	sm.services["gw"].DebugPort = 5009
	if _, err := sm.WriteIntelliJAttachConfig("gw", projectsDir); err != nil {
		t.Fatalf("second WriteIntelliJAttachConfig: %v", err)
	}
	content, _ = os.ReadFile(path)
	if got := parseRunConfiguration(t, string(content)).option("PORT"); got != "5009" {
		t.Errorf("rewritten PORT = %q, want 5009", got)
	}
}

func TestWriteIntelliJAttachConfigNeedsADebugPort(t *testing.T) {
	sm := newTestManager(t)
	projectsDir := t.TempDir()
	sm.services["gw"] = &models.Service{ID: "gw", Name: "gateway", Dir: "."}

	_, err := sm.WriteIntelliJAttachConfig("gw", projectsDir)
	if err == nil || !strings.Contains(err.Error(), "turn on debugging") {
		t.Errorf("WriteIntelliJAttachConfig() error = %v, want one asking to turn on debugging", err)
	}
	if _, statErr := os.Stat(filepath.Join(projectsDir, ".run")); !os.IsNotExist(statErr) {
		t.Errorf("a .run directory was created for a service with no debug port")
	}
}

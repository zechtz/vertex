// Package services - IntelliJ run configurations for attaching a debugger
package services

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// intellijAttachConfigTemplate is a "Remote JVM Debug" run configuration in
// the form IntelliJ itself stores in .run/*.run.xml. The first verb is the
// configuration's name, the other two its port.
const intellijAttachConfigTemplate = `<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="%s" type="Remote">
    <option name="USE_SOCKET_TRANSPORT" value="true" />
    <option name="SERVER_MODE" value="false" />
    <option name="SHMEM_ADDRESS" />
    <option name="HOST" value="localhost" />
    <option name="PORT" value="%d" />
    <option name="AUTO_RESTART" value="false" />
    <RunnerSettings RunnerId="Debug">
      <option name="DEBUG_PORT" value="%d" />
      <option name="LOCAL" value="false" />
    </RunnerSettings>
    <method v="2" />
  </configuration>
</component>
`

// unsafeFileNameChars are the characters a service name may contain that a
// file name on some platform may not.
var unsafeFileNameChars = strings.NewReplacer(
	"/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", "\"", "-", "<", "-", ">", "-", "|", "-",
)

// intellijAttachConfigName is the name IntelliJ shows in its run menu.
func intellijAttachConfigName(serviceName string) string {
	return "Attach " + serviceName
}

// intellijAttachConfig renders the run configuration that attaches IntelliJ's
// debugger to a service listening on debugPort.
func intellijAttachConfig(serviceName string, debugPort int) string {
	var name bytes.Buffer
	// EscapeText also escapes quotes, which matters here: the name sits in an
	// attribute value.
	xml.EscapeText(&name, []byte(intellijAttachConfigName(serviceName)))
	return fmt.Sprintf(intellijAttachConfigTemplate, name.String(), debugPort, debugPort)
}

// WriteIntelliJAttachConfig writes a run configuration that attaches IntelliJ's
// debugger to the service, and returns the file's path. It goes in the
// service's .run directory, where IntelliJ picks up run configurations stored
// as project files, so it appears in the run menu without an import step; it
// can be committed to share it with the team.
//
// The file is rewritten on every call, so it follows a change of debug port.
// It uses the service's debug port whether or not debugging is on at the
// moment, since the port is kept while it is off.
func (sm *Manager) WriteIntelliJAttachConfig(serviceUUID, projectsDir string) (string, error) {
	sm.mutex.RLock()
	service, exists := sm.services[serviceUUID]
	sm.mutex.RUnlock()
	if !exists {
		return "", fmt.Errorf("service with UUID %s not found", serviceUUID)
	}

	service.Mutex.RLock()
	name, dir, debugPort := service.Name, service.Dir, service.DebugPort
	service.Mutex.RUnlock()

	if debugPort == 0 {
		return "", fmt.Errorf("service %s has no debug port yet; turn on debugging for it first", name)
	}

	serviceDir := filepath.Join(projectsDir, dir)
	if _, err := os.Stat(serviceDir); err != nil {
		return "", fmt.Errorf("service directory %s is not accessible: %w", serviceDir, err)
	}

	runDir := filepath.Join(serviceDir, ".run")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create %s: %w", runDir, err)
	}

	fileName := unsafeFileNameChars.Replace(intellijAttachConfigName(name)) + ".run.xml"
	path := filepath.Join(runDir, fileName)
	if err := os.WriteFile(path, []byte(intellijAttachConfig(name, debugPort)), 0o644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", path, err)
	}
	return path, nil
}

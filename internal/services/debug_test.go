package services

import (
	"net"
	"strings"
	"testing"

	"github.com/zechtz/vertex/internal/models"
)

const debugAgent5005 = "-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=127.0.0.1:5005"

// The debug agent must reach the application's JVM and nothing else. MAVEN_OPTS
// is Maven's own JVM: an agent there opens the port first, and the forked
// application then fails to bind it.
func TestMavenStartCommandGivesDebugAgentToApplicationOnly(t *testing.T) {
	cmd, err := GetStartCommand(t.TempDir(), "maven", "-Xmx512m", "", false, 5005)
	if err != nil {
		t.Fatalf("GetStartCommand: %v", err)
	}

	if !strings.Contains(cmd, `-Dspring-boot.run.jvmArguments="-Xmx512m `+debugAgent5005+`"`) {
		t.Errorf("application JVM arguments lack the java options and debug agent: %s", cmd)
	}
	if !strings.Contains(cmd, `MAVEN_OPTS="-Xmx512m" `) {
		t.Errorf("MAVEN_OPTS should carry the java options alone: %s", cmd)
	}
}

func TestMavenStartCommandWithOnlyDebugHasNoMavenOpts(t *testing.T) {
	cmd, err := GetStartCommand(t.TempDir(), "maven", "", "", false, 5005)
	if err != nil {
		t.Fatalf("GetStartCommand: %v", err)
	}

	if !strings.Contains(cmd, `-Dspring-boot.run.jvmArguments="`+debugAgent5005+`"`) {
		t.Errorf("application JVM arguments lack the debug agent: %s", cmd)
	}
	if strings.Contains(cmd, "MAVEN_OPTS") {
		t.Errorf("command sets MAVEN_OPTS without java options: %s", cmd)
	}
}

func TestGradleStartCommandGivesDebugAgentToApplicationOnly(t *testing.T) {
	useTempCacheDir(t)

	cmd, err := GetStartCommand(t.TempDir(), "gradle", "-Xmx512m", "", false, 5005)
	if err != nil {
		t.Fatalf("GetStartCommand: %v", err)
	}

	if !strings.Contains(cmd, `-PvertexAppJvmArgs="-Xmx512m `+debugAgent5005+`"`) {
		t.Errorf("bootRun JVM arguments lack the java options and debug agent: %s", cmd)
	}
	if !strings.Contains(cmd, `GRADLE_OPTS="-Xmx512m" `) {
		t.Errorf("GRADLE_OPTS should carry the java options alone: %s", cmd)
	}
}

// JDWP has no authentication, so the port must never listen beyond localhost.
func TestDebugAgentListensOnLocalhostOnly(t *testing.T) {
	if got := debugAgentArg(5005); !strings.HasSuffix(got, "address=127.0.0.1:5005") {
		t.Errorf("debugAgentArg(5005) = %q, want it bound to 127.0.0.1", got)
	}
	if got := debugAgentArg(0); got != "" {
		t.Errorf("debugAgentArg(0) = %q, want no agent", got)
	}
}

func TestActiveDebugPortIsZeroWhileDebuggingIsOff(t *testing.T) {
	service := &models.Service{DebugEnabled: false, DebugPort: 5005}
	if got := activeDebugPort(service); got != 0 {
		t.Errorf("activeDebugPort() = %d with debugging off, want 0", got)
	}

	service.DebugEnabled = true
	if got := activeDebugPort(service); got != 5005 {
		t.Errorf("activeDebugPort() = %d with debugging on, want 5005", got)
	}
}

func TestResolveDebugPort(t *testing.T) {
	sm := newTestManager(t)
	sm.services["gateway"] = &models.Service{ID: "gateway", Name: "gateway", Port: 5005}
	sm.services["registry"] = &models.Service{ID: "registry", Name: "registry", Port: 8761, DebugPort: 5006}

	t.Run("assigns a port no other service claims", func(t *testing.T) {
		port, err := sm.resolveDebugPort("auth", true, 0, 8080)
		if err != nil {
			t.Fatalf("resolveDebugPort: %v", err)
		}
		if port < firstDebugPort || port == 5005 || port == 5006 {
			t.Errorf("assigned port %d, want a port from %d not claimed by another service", port, firstDebugPort)
		}
	})

	t.Run("keeps a chosen port", func(t *testing.T) {
		port, err := sm.resolveDebugPort("auth", true, 5100, 8080)
		if err != nil || port != 5100 {
			t.Errorf("resolveDebugPort() = %d, %v; want 5100, nil", port, err)
		}
	})

	t.Run("keeps the port while debugging is off", func(t *testing.T) {
		port, err := sm.resolveDebugPort("auth", false, 5100, 8080)
		if err != nil || port != 5100 {
			t.Errorf("resolveDebugPort() = %d, %v; want 5100, nil", port, err)
		}
	})

	t.Run("does not assign a port while debugging is off", func(t *testing.T) {
		port, err := sm.resolveDebugPort("auth", false, 0, 8080)
		if err != nil || port != 0 {
			t.Errorf("resolveDebugPort() = %d, %v; want 0, nil", port, err)
		}
	})

	errorCases := []struct {
		name      string
		requested int
		want      string
	}{
		{"rejects the service's own port", 8080, "service's own port"},
		{"rejects another service's port", 5005, "already used by service gateway"},
		{"rejects another service's debug port", 5006, "already used by service registry"},
		{"rejects an impossible port", 70000, "not a valid port"},
	}
	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sm.resolveDebugPort("auth", true, tc.requested, 8080)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("resolveDebugPort(%d) error = %v, want one mentioning %q", tc.requested, err, tc.want)
			}
		})
	}
}

// The debug settings survive a save and a reload, through both the update path
// the config modal uses and the load path startup uses.
func TestDebugSettingsPersist(t *testing.T) {
	sm := newTestManager(t)
	service := &models.Service{ID: "gateway", Name: "gateway", Dir: "gateway", Port: 8080}
	if err := sm.insertServiceInDB(service); err != nil {
		t.Fatalf("insertServiceInDB: %v", err)
	}
	// The insert leaves last_started NULL, which the loader cannot scan. Any
	// status change writes it, so a real service always has one by now.
	if err := sm.UpdateServiceInDB(service); err != nil {
		t.Fatalf("UpdateServiceInDB: %v", err)
	}

	service.DebugEnabled = true
	service.DebugPort = 5010
	if err := sm.UpdateServiceConfigInDB(service); err != nil {
		t.Fatalf("UpdateServiceConfigInDB: %v", err)
	}

	sm.services = make(map[string]*models.Service)
	if err := sm.loadDynamicServices(); err != nil {
		t.Fatalf("loadDynamicServices: %v", err)
	}

	loaded, ok := sm.services["gateway"]
	if !ok {
		t.Fatal("service was not reloaded")
	}
	if !loaded.DebugEnabled || loaded.DebugPort != 5010 {
		t.Errorf("reloaded debugEnabled=%v debugPort=%d, want true, 5010", loaded.DebugEnabled, loaded.DebugPort)
	}
}

// A start whose debug port is taken fails up front, naming the port, rather
// than leaving the JVM to die on it partway through startup.
func TestEnsureDebugPortFreeRefusesATakenPort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to open a listener: %v", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	service := &models.Service{Name: "gateway", DebugEnabled: true, DebugPort: port}
	err = ensureDebugPortFree(service)
	if err == nil || !strings.Contains(err.Error(), "debug port") {
		t.Fatalf("ensureDebugPortFree() = %v, want an error naming the debug port", err)
	}

	service.DebugEnabled = false
	if err := ensureDebugPortFree(service); err != nil {
		t.Errorf("ensureDebugPortFree() = %v with debugging off, want nil", err)
	}
}

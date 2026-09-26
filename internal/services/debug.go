// Package services - attaching a debugger to a running service
package services

import (
	"fmt"

	"github.com/zechtz/vertex/internal/models"
)

const (
	// firstDebugPort is where port assignment starts: the JDWP port IDEs and
	// jdb assume by default, so the first debugged service needs no setup.
	firstDebugPort = 5005
	// debugPortSearchLimit bounds how far assignment looks past firstDebugPort.
	debugPortSearchLimit = 1000
)

// debugAgentArg returns the JVM option that opens a debugger port, or "" when
// debugPort is 0.
//
// The port binds to 127.0.0.1 only. JDWP has no authentication: anyone who can
// reach the port can run arbitrary code in the JVM, so it must not be open to
// the network. suspend=n lets the service start without waiting for a debugger
// to attach, which is what makes it reasonable to leave debugging on.
func debugAgentArg(debugPort int) string {
	if debugPort == 0 {
		return ""
	}
	return fmt.Sprintf("-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=127.0.0.1:%d", debugPort)
}

// activeDebugPort is the port the service's JVM should open for a debugger, or
// 0 when debugging is off.
func activeDebugPort(service *models.Service) int {
	if !service.DebugEnabled {
		return 0
	}
	return service.DebugPort
}

// resolveDebugPort decides which debug port to store for a service. Turning
// debugging on without a port assigns the first free one. A port the user chose
// must not be the service's own port or one another service claims. The port is
// kept while debugging is off, so turning it back on finds the same one.
// Callers must hold sm.mutex.
func (sm *Manager) resolveDebugPort(serviceID string, enabled bool, requested, servicePort int) (int, error) {
	if requested < 0 || requested > 65535 {
		return 0, fmt.Errorf("debug port %d is not a valid port", requested)
	}
	if !enabled {
		return requested, nil
	}
	if requested == 0 {
		return sm.nextFreeDebugPort(serviceID, servicePort)
	}

	if requested == servicePort {
		return 0, fmt.Errorf("debug port %d is the service's own port", requested)
	}
	if owner := sm.portOwner(serviceID, requested); owner != "" {
		return 0, fmt.Errorf("debug port %d is already used by service %s", requested, owner)
	}
	return requested, nil
}

// nextFreeDebugPort returns the first port from firstDebugPort that no service
// claims and nothing is listening on. Callers must hold sm.mutex.
func (sm *Manager) nextFreeDebugPort(serviceID string, servicePort int) (int, error) {
	for port := firstDebugPort; port < firstDebugPort+debugPortSearchLimit; port++ {
		if port == servicePort || sm.portOwner(serviceID, port) != "" {
			continue
		}
		if len(findProcessesOnPort(port)) > 0 {
			continue
		}
		return port, nil
	}
	return 0, fmt.Errorf("no free debug port between %d and %d", firstDebugPort, firstDebugPort+debugPortSearchLimit-1)
}

// portOwner names the service other than serviceID that uses port as its
// service port or its debug port, or returns "" if none does. A debug port
// counts while debugging is off, since it is kept for when it is turned back
// on. Callers must hold sm.mutex.
func (sm *Manager) portOwner(serviceID string, port int) string {
	for id, other := range sm.services {
		if id == serviceID {
			continue
		}
		if other.Port == port || other.DebugPort == port {
			return other.Name
		}
	}
	return ""
}

// ensureDebugPortFree fails a start whose debug port is already taken, rather
// than letting the JVM die on it partway through startup.
//
// Unlike the service port, whatever holds the debug port is not killed. It is
// usually another JVM being debugged, often one started from an IDE, and not a
// leftover of this service: a leftover also holds the service port, so the
// service port cleanup has already stopped it by the time this runs.
func ensureDebugPortFree(service *models.Service) error {
	port := activeDebugPort(service)
	if port == 0 {
		return nil
	}
	if pids := findProcessesOnPort(port); len(pids) > 0 {
		return fmt.Errorf("debug port %d for service %s is in use by process(es) %v; stop them or change the service's debug port",
			port, service.Name, pids)
	}
	return nil
}

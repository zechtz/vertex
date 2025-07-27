// Package config
package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/zechtz/nest-up/internal/models"
)

// getPortFromEnv gets a port from environment variable, returns defaultPort if not found or invalid
func getPortFromEnv(envVarName string, defaultPort int) int {
	if portStr := os.Getenv(envVarName); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			return port
		}
	}
	return defaultPort
}

func LoadDefaultConfig() models.Config {
	return models.Config{
		ProjectsDir:      "/Volumes/Work/sites/nest",
		JavaHomeOverride: "/Users/mtabe/.asdf/installs/java/openjdk-23",
		Services: []models.Service{
			{Name: "EUREKA", Dir: "nest-registry-server", Status: "stopped", HealthStatus: "unknown", Port: 8800, HealthURL: "http://localhost:8800/actuator/health", Order: 1},
			{Name: "CONFIG", Dir: "nest-config-server", ExtraEnv: "SPRING_PROFILES_ACTIVE=native &&", Status: "stopped", HealthStatus: "unknown", Port: 8801, HealthURL: "http://localhost:8801/actuator/health", Order: 2},
			{Name: "CACHE", Dir: "nest-cache", ExtraEnv: "export JAVA_HOME=\"/Users/mtabe/.asdf/installs/java/openjdk-23\" && export PATH=\"$JAVA_HOME/bin:$PATH\" &&", Status: "stopped", HealthStatus: "unknown", Port: 8814, HealthURL: "http://localhost:8814/actuator/health", Order: 3},
			{Name: "GATEWAY", Dir: "nest-gateway", Status: "stopped", HealthStatus: "unknown", Port: 8802, HealthURL: "http://localhost:8802/actuator/health", Order: 4},
			{Name: "UAA", Dir: "nest-uaa", Status: "stopped", HealthStatus: "unknown", Port: 8803, HealthURL: "http://localhost:8803/actuator/health", Order: 5},
			{Name: "APP", Dir: "nest-app", Status: "stopped", HealthStatus: "unknown", Port: 8805, HealthURL: "http://localhost:8805/actuator/health", Order: 6},
			{Name: "CONTRACT", Dir: "nest-contract-management", Status: "stopped", HealthStatus: "unknown", Port: 8818, HealthURL: "http://localhost:8818/actuator/health", Order: 7},
			{Name: "SUBMISSION", Dir: "nest-submission", Status: "stopped", HealthStatus: "unknown", Port: 8817, HealthURL: "http://localhost:8810/actuator/health", Order: 8},
			{Name: "DSMS", Dir: "nest-dsms", JavaOpts: "--add-exports=java.base/sun.nio.ch=ALL-UNNAMED --add-opens=java.base/java.lang=ALL-UNNAMED --add-opens=java.base/java.lang.reflect=ALL-UNNAMED --add-opens=java.base/java.io=ALL-UNNAMED --add-exports=jdk.unsupported/sun.misc=ALL-UNNAMED --add-opens=java.base/java.security=ALL-UNNAMED", Status: "stopped", HealthStatus: "unknown", Port: 8804, HealthURL: "http://localhost:8812/actuator/health", Order: 9},
			{Name: "MONITOR", Dir: "nest-monitor", Status: "stopped", HealthStatus: "unknown", Port: 8815, HealthURL: "http://localhost:8815/actuator/health", Order: 10},
		},
	}
}

func LoadEnvironmentVariables() error {
	fishFile := "env_vars.fish"
	file, err := os.Open(fishFile)
	if err != nil {
		return fmt.Errorf("could not open %s: %w", fishFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	envVarRegex := regexp.MustCompile(`^set -gx (\w+) (.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		matches := envVarRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			value := strings.Trim(matches[2], `"`)
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

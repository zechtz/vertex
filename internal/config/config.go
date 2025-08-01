// Package config
package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/zechtz/vertex/internal/models"
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
		ProjectsDir:      "",                 // Empty - user will configure
		JavaHomeOverride: "",                 // Empty - user will configure
		Services:         []models.Service{}, // Empty - user will add services
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
